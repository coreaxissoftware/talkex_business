/**
 * Client for the /events/stream SSE endpoint. One connection per
 * browser tab — components subscribe by event type.
 *
 * Auto-reconnects on drop with backoff. Uses the current access_token
 * from localStorage as a query param (EventSource can't set headers).
 */

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

  private baseUrl(): string {
    // Use same base as axios does
    return (import.meta.env.VITE_API_URL as string | undefined) || 'http://localhost:8080'
  }

  connect() {
    if (this.es) return
    const token = localStorage.getItem('access_token')
    if (!token) return

    const url = `${this.baseUrl()}/events/stream?token=${encodeURIComponent(token)}`
    const es = new EventSource(url)
    this.es = es

    es.addEventListener('hello', () => {
      this.connected = true
      this.backoff = 1000
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
      this.connected = false
      es.close()
      this.es = null
      // Reconnect with cap
      const delay = Math.min(this.backoff, 30_000)
      this.reconnectTimer = setTimeout(() => this.connect(), delay)
      this.backoff = Math.min(this.backoff * 2, 30_000)
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
