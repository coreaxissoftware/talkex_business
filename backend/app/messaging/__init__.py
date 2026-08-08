"""Message Queue -> Channel Router -> Delivery Engine (see CONTEXT.md
"Message flow"). MVP-blocking gaps to model here: idempotency keys, Dead
Letter Queue + retry policy, priority queues (OTP > transactional >
marketing). Not yet implemented.
"""
