# TalkEx Messenger — /api/v1/inbound endpoint

The TalkEx Business inbound poller ([`internal/channels/talkex/poller.go`](../internal/channels/talkex/poller.go))
polls TalkEx every 30s to bridge customer replies into the merchant's
Conversations inbox. It needs one endpoint added to the TalkEx Messenger
FastAPI backend (`D:\WebProject\APP\TalkEx Messenger\backend\main.py`).

## Contract the poller expects

```
GET /api/v1/inbound?since=<unix_seconds>&limit=<int 1..500>
  Header: Authorization: Bearer <talkex_bulk_api_key>
  200 → {
    "messages": [
      {
        "message_id":    "<str>",
        "chat_id":       "<str>",
        "from_username": "<str>",
        "text":          "<str>",
        "kind":          "text|photo|voice|document|location|contact",
        "created_at":    <float unix seconds>,
        "seq":           <int64>
      },
      …
    ],
    "next_since": <float>,   // pass on next call
    "count":      <int>
  }

  401 → invalid or missing bulk key
  429 → rate-limited (poller backs off + retries)
```

Rules mirroring `bulk_send_message`:

- **Only DMs** (`chats.type = 'dm'`) — bulk keys never see a group / channel /
  community.
- **Only inbound** (`m.sender_id != sender.id`) — the merchant's own outbox
  stays private to the interactive session-token path.
- **Skip deleted-for-everyone rows** (`m.deleted_for_everyone = 0`).
- **Skip system / poll / meeting frames** — only substantive kinds surface.
- Rate limit: 120 pulls/minute per key (a 30s poller uses 4x headroom).

## The patch — paste into `backend/main.py`

Insert immediately after the closing `return` of `bulk_send_message`
(currently around line 2113), before `# ── Profile ────`:

```python
bulk_inbound_rate_limiter = RateLimiter(max_events=120, window_seconds=60)


@app.get("/api/v1/inbound")
async def bulk_inbound_messages(
    since: float = 0.0,
    limit: int = 100,
    sender: dict = Depends(require_api_key),
):
    """
    Return DM messages this business account has RECEIVED since a timestamp.

    Companion to POST /api/v1/messages so an external CRM (TalkEx Business,
    Zapier bridge, etc.) can pull replies into its own inbox — the WhatsApp
    Business "webhook receive" pattern, but pull-based so the CRM doesn't
    need a public URL for us to reach.

    Deliberately narrow, like the send endpoint:
      - Only DMs (chats.type = 'dm').
      - Only inbound (m.sender_id != sender.id).
      - Only substantive kinds — poll/system/meeting frames excluded.
      - Only non-deleted rows.
      - Capped at 500 rows/call; 120 pulls/min/key.

    The `next_since` cursor is tied to the newest row's created_at, so a
    poller that stops for an hour and resumes doesn't miss messages.
    """
    bulk_inbound_rate_limiter.check(sender["id"])

    if limit < 1:
        limit = 1
    if limit > 500:
        limit = 500

    rows = db.query_all(
        """
        SELECT m.id, m.chat_id, m.sender_id, m.text, m.kind,
               m.created_at, m.seq,
               u.username AS from_username
        FROM messages m
        JOIN chats c        ON c.id = m.chat_id
        JOIN chat_members cm ON cm.chat_id = m.chat_id
        JOIN users u        ON u.id = m.sender_id
        WHERE cm.user_id = ?
          AND c.type = 'dm'
          AND m.sender_id != ?
          AND m.created_at > ?
          AND m.deleted_for_everyone = 0
          AND m.kind IN ('text','photo','voice','document','location','contact')
        ORDER BY m.created_at ASC
        LIMIT ?
        """,
        (sender["id"], sender["id"], since, limit),
    )

    messages = []
    newest = since
    for r in rows:
        messages.append({
            "message_id":    r["id"],
            "chat_id":       r["chat_id"],
            "from_username": r["from_username"],
            "text":          r["text"] or "",
            "kind":          r["kind"],
            "created_at":    r["created_at"],
            "seq":           r["seq"],
        })
        if r["created_at"] > newest:
            newest = r["created_at"]

    return {"messages": messages, "next_since": newest, "count": len(messages)}
```

**Column-name check** — the SQL above assumes the following columns on the
tables it joins; adjust names if your `db.py` differs:

- `messages.id`, `.chat_id`, `.sender_id`, `.text`, `.kind`, `.created_at`,
  `.seq`, `.deleted_for_everyone`
- `chats.id`, `.type`
- `chat_members.chat_id`, `.user_id`
- `users.id`, `.username`

## After deploy

1. Restart the TalkEx backend on Render — the endpoint is live.
2. In TalkEx Business, enable the TalkEx channel with the merchant's bulk
   API key (already documented in the connector docstring).
3. Within 30 seconds of a customer replying on TalkEx, the message shows
   up in the merchant's Conversations inbox exactly the same way an
   inbound WhatsApp / Telegram / SMS message would — same hooks, same
   labels, same AI auto-tag, same SSE stream.

## Test end-to-end

From the merchant's terminal after enabling the channel:

```bash
# Confirm the poller sees the endpoint
curl -s -H "Authorization: Bearer $TALKEX_KEY" \
  "https://talkex-backend.onrender.com/api/v1/inbound?since=0&limit=5" | jq

# Have a customer reply on TalkEx to the merchant's account…
# …then watch the TalkEx Business server log:
#   "talkex poller: started (interval 30s)"
#   (30s later, when the reply lands)
#   (nothing — quiet is good)
#   (reply arrives — poller enqueues via RecordInbound)

# And the reply now shows up in the Conversations tab of TalkEx Business.
```

The poller's state is visible per-merchant at:

```sql
SELECT owner_id, last_since, last_run_at, last_error
FROM talkex_poller_state;
```

so a stalled tenant is trivially diagnosable.
