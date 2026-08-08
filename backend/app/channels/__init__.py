"""Pluggable channel connectors. `shared/` holds the common Channel
interface every connector implements (send, receive webhook, status
callback); `talkex/` and `whatsapp/` are first implementations. Adding a
new channel (Telegram, RCS, Email...) means adding a package here without
touching Contacts/Templates/Campaigns/Analytics. See CONTEXT.md.
"""
