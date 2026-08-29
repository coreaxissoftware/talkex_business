/**
 * Client for the /events/stream SSE endpoint. One connection per
 * browser tab — components subscribe by event type.
 *
 * Auto-reconnects on drop with backoff. Uses the current access_token
 * from localStorage as a query param (EventSource can't set headers).
 * Refreshes the token on repeated failures so long-lived sessions
 * survive JWTAccessMinutes expiring mid-stream.
 */

import api from './api'

export type EventType =
  | 'message.inbound'
  | 'message.outbound'
  | 'message.status'
  | 'notification.new'
  | 'campaign.changed'
  | 'conversation.update'

type Listener = (data: any) => void

class EventStream {
  private es: EventSource | null = null
  private listeners = new Map<EventType, Set<Listener>>()
  private connected = false
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private backoff = 1000
  // Consecutive errors before the last successful `hello`. Two in a
  // row usually means the token expired; on the second failure we
  // try to refresh before reconnecting so we don't loop forever
  // with a dead token.
  private consecutiveErrors = 0
  private refreshing = false

  private baseUrl(): string {
    // Use same base as axios does
    return (import.meta.env.VITE_API_URL as string | undefined) || 'http://localhost:8080'
  }

  connect() {
    // Cancel any pending reconnect so we don't spawn duplicate streams
    // when connect() is called before the timer fires (React StrictMode
    // remounts, manual reconnect, re-login).
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.es) return
    const token = localStorage.getItem('access_token')
    if (!token) return

    const url = `${this.baseUrl()}/events/stream?token=${encodeURIComponent(token)}`
    const es = new EventSource(url)
    this.es = es

    es.addEventListener('hello', () => {
      this.connected = true
      this.backoff = 1000
      this.consecutiveErrors = 0
    })

    // Wire each known event type as a distinct listener so we can
    // dispatch cleanly (SSE 'event:' names are the type strings).
    const types: EventType[] = [
      'message.inbound',
      'message.outbound',
      'message.status',
      'notification.new',
      'campaign.changed',
      'conversation.update',
    ]
    for (const t of types) {
      es.addEventListener(t, (e: MessageEvent) => {
        try {
          const parsed = JSON.parse(e.data)
          this.emit(t, parsed.data ?? parsed)
        } catch {
          /* ignore malformed payloads */
        }
      })
    }

    es.onerror = () => {
      const wasConnected = this.connected
      this.connected = false
      es.close()
      this.es = null
      // Count failures since the last successful `hello`. If we
      // never even got hello twice in a row, the JWT is probably
      // dead — try refreshing before the next reconnect.
      if (!wasConnected) this.consecutiveErrors++
      const shouldRefresh = this.consecutiveErrors >= 2 && !this.refreshing
      const delay = Math.min(this.backoff, 30_000)
      this.reconnectTimer = setTimeout(() => {
        this.reconnectTimer = null
        if (shouldRefresh) {
          void this.refreshAndReconnect()
        } else {
          this.connect()
        }
      }, delay)
      this.backoff = Math.min(this.backoff * 2, 30_000)
    }
  }

  // Ask the backend for a fresh access token using the stored refresh
  // token, then reconnect with the new access token. If the refresh
  // itself fails, fall back to a normal reconnect attempt (the axios
  // interceptor may still recover, and if not the user will see 401
  // on the next API call and be redirected to /login).
  private async refreshAndReconnect() {
    this.refreshing = true
    try {
      const refresh = localStorage.getItem('refresh_token')
      if (refresh) {
        const res = await api.post('/auth/refresh', { refresh_token: refresh })
        if (res.data?.access_token) {
          localStorage.setItem('access_token', res.data.access_token)
        }
        if (res.data?.refresh_token) {
          localStorage.setItem('refresh_token', res.data.refresh_token)
        }
        // Reset the failure counter — we're about to try a fresh token.
        this.consecutiveErrors = 0
      }
    } catch {
      /* fall through — reconnect will fail again and back off */
    } finally {
      this.refreshing = false
      this.connect()
    }
  }

  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    this.es?.close()
    this.es = null
    this.connected = false
  }

  isConnected(): boolean {
    return this.connected
  }

  on(type: EventType, fn: Listener): () => void {
    let set = this.listeners.get(type)
    if (!set) {
      set = new Set()
      this.listeners.set(type, set)
    }
    set.add(fn)
    return () => set!.delete(fn)
  }

  private emit(type: EventType, data: any) {
    this.listeners.get(type)?.forEach(fn => {
      try { fn(data) } catch (e) { console.error('sse listener error', e) }
    })
  }
}

export const eventStream = new EventStream()
