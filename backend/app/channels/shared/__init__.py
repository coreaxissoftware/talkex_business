"""Common Channel interface (send, receive webhook, status callback) that
every connector under channels/{talkex,whatsapp,...} implements, so the
rest of the app (Campaigns, Message Queue, Delivery Engine) is
channel-agnostic. See CONTEXT.md "Message flow". Not yet implemented.
"""
