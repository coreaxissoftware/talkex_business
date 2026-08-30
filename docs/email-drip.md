# Onboarding email drip — 5 emails, 7 days

The moment a new tenant signs up, they land in a paved-path drip. Every
email has one job — one action to take, one link to click. Copy is
intentionally short and unpolished. Send from `aditi@business.talkex.in`.

---

## Day 0 · 2 minutes after signup

**Subject:** Welcome — one thing to do next
**Preheader:** Wire up your first channel, then send a message to yourself.

Hey {{name}},

Aditi here — one of the two founders. Thanks for signing up for TalkEx.

The fastest way to know if this is going to work for you: wire up one channel and send a message to your own number. Here's the 90-second version:

1. Open the [Channels page](https://app.business.talkex.in/channels)
2. Click **TalkEx (sandbox)** — no setup, works instantly
3. Under Contacts, add your own phone number
4. Send yourself a message

That's it. If any step doesn't work, hit reply — my inbox is real, not automation.

— Aditi

P.S. If you're moving off Gupshup / Wati / Interakt, tell me — we've done migration hand-holding for a dozen teams and I'll set aside 30 minutes.

---

## Day 1

**Subject:** Have you tried the AI Assist button?
**Preheader:** ✦ next to the send box — Claude drafts your reply.

Hey {{name}},

Quick one — the ✦ button next to the send box in any conversation is our AI Assist. Click it, Claude reads the last 30 messages, drafts a reply. You edit before sending.

We built it after watching our own agents write "Great, thanks!" 40 times a day. If your team is doing the same, this will save real hours.

Free tier includes 100 AI requests / month. Growth plan bumps that to 1,000.

Try it → https://app.business.talkex.in/conversations

— Aditi

---

## Day 3

**Subject:** Chatbot or campaign — which is on your list?
**Preheader:** Two 5-minute setup guides inside.

Hey {{name}},

Most teams' first serious project on TalkEx is one of two things:

- **A chatbot** for "where's my order" / "what are your hours" queries — the [flow builder](https://app.business.talkex.in/flows) covers this
- **A campaign** — bulk WhatsApp / SMS to a contact list — the [campaigns page](https://app.business.talkex.in/campaigns) is the entry point

If you're not sure which fits, reply and tell me what your customers ask most. I'll point you at the shorter path.

— Aditi

---

## Day 5

**Subject:** How is TalkEx going?
**Preheader:** Honest question — I read every reply.

Hey {{name}},

You've had TalkEx for five days. Two questions:

1. What's the one thing you wish worked differently?
2. What's the one feature you thought you needed and now realise you don't?

I'm asking because we're a three-person team and the roadmap is set almost entirely by what customers tell us in emails like this one.

— Aditi

---

## Day 7

**Subject:** Ready to upgrade? Here's what you unlock
**Preheader:** Growth plan — ₹2,499/mo, no per-seat fees.

Hey {{name}},

You're on the free tier — 500 messages / month, all 8 channels, unlimited agents. Good place to start.

If you're ready to send more, the Growth plan is ₹2,499/mo and unlocks:

- 1,000 AI Assist requests
- Unlimited chatbot flows (free tier caps at 3)
- Campaign approval workflow
- Per-agent CSAT dashboard
- Priority email support (< 4 hr response)

[Upgrade in one click](https://app.business.talkex.in/billing) → no card is charged for 14 days on the Growth trial.

— Aditi

---

## Sending mechanics

- Store templates in a small YAML file (`emails/*.yml`) and render with a Go template at send-time.
- Use SendGrid or AWS SES with a dedicated subdomain (`mail.business.talkex.in`) — set up SPF + DKIM + DMARC before sending anything.
- Track opens + clicks in the SendGrid dashboard — never inline pixels in the copy itself.
- Every email has a plain-text alternative auto-generated from the markdown source.
- Suppress-list respected across the whole drip — if someone unsubscribes on Day 1, none of Days 3, 5, 7 fire.
