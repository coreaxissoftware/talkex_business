package widget

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// -----------------------------------------------------------------------
// Injected wires (from main.go) — keep this package free of contacts and
// conversations imports.
// -----------------------------------------------------------------------

// ContactCreator creates an anonymous webchat contact and returns its ID.
type ContactCreator func(ownerID, name, email string) (contactID string, err error)

// InboundRecorder records an inbound message from the visitor.
type InboundRecorder func(ownerID, contactID, body string) error

// OutboundStreamer returns a channel of outbound messages for a
// specific conversation ID plus a cancel func. The channel closes when
// cancel is called or the underlying pub/sub disconnects.
type OutboundStreamer func(conversationID string) (<-chan OutboundMsg, func())

// MessageLister returns all messages for a conversation, oldest first.
type MessageLister func(conversationID string) ([]OutboundMsg, error)

// OutboundMsg is the plain shape the widget streams to visitors.
type OutboundMsg struct {
	ID        string    `json:"id"`
	Direction string    `json:"direction"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	contactCreator  ContactCreator
	inboundRecorder InboundRecorder
	outboundStream  OutboundStreamer
	messageLister   MessageLister
	wireMu          sync.RWMutex
)

func RegisterContactCreator(f ContactCreator)   { wireMu.Lock(); contactCreator = f; wireMu.Unlock() }
func RegisterInboundRecorder(f InboundRecorder) { wireMu.Lock(); inboundRecorder = f; wireMu.Unlock() }
func RegisterOutboundStreamer(f OutboundStreamer) {
	wireMu.Lock()
	outboundStream = f
	wireMu.Unlock()
}
func RegisterMessageLister(f MessageLister) { wireMu.Lock(); messageLister = f; wireMu.Unlock() }

// -----------------------------------------------------------------------
// Routes
// -----------------------------------------------------------------------

func RegisterRoutes(r *gin.Engine) {
	// Public widget API — CORS is handled by the global middleware.
	g := r.Group("/widget")
	{
		g.GET("/config", handlePublicConfig) // ?key=<publicKey>
		g.POST("/init", handleInit)
		g.POST("/message", handleMessage)
		g.GET("/messages", handleListMessages)
		g.GET("/stream", handleStream)
		g.GET("/snippet.js", handleSnippet) // embeddable JS
	}

	// Authenticated owner config
	admin := r.Group("/settings/widget")
	admin.Use(auth.AuthRequired())
	{
		admin.GET("", handleGetConfig)
		admin.PATCH("", handleUpdateConfig)
		admin.POST("/rotate-key", handleRotateKey)
	}
}

// -----------------------------------------------------------------------
// Owner config endpoints
// -----------------------------------------------------------------------

func handleGetConfig(c *gin.Context) {
	cfg, err := GetOrCreateConfig(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func handleUpdateConfig(c *gin.Context) {
	cfg, err := GetOrCreateConfig(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	var in UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	updated, err := UpdateConfig(database.DB, cfg, &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func handleRotateKey(c *gin.Context) {
	cfg, err := GetOrCreateConfig(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	rotated, err := RotateKey(database.DB, cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, rotated)
}

// -----------------------------------------------------------------------
// Public widget endpoints
// -----------------------------------------------------------------------

// handlePublicConfig returns the customization fields (title, greeting,
// color) — nothing else, so a leaked key can't enumerate anything.
func handlePublicConfig(c *gin.Context) {
	cfg, err := FindConfigByKey(database.DB, c.Query("key"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Widget not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"title":       cfg.Title,
		"greeting":    cfg.Greeting,
		"theme_color": cfg.ThemeColor,
	})
}

type initReq struct {
	Key          string `json:"key" binding:"required"`
	VisitorName  string `json:"visitor_name"`
	VisitorEmail string `json:"visitor_email"`
	PageURL      string `json:"page_url"`
}

func handleInit(c *gin.Context) {
	var req initReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	cfg, err := FindConfigByKey(database.DB, req.Key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Widget not found"})
		return
	}

	wireMu.RLock()
	creator := contactCreator
	wireMu.RUnlock()
	if creator == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"detail": "Widget not wired"})
		return
	}

	name := req.VisitorName
	if name == "" {
		name = "Website visitor"
	}
	contactID, err := creator(cfg.OwnerID, name, req.VisitorEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create contact: " + err.Error()})
		return
	}

	// Record an initial inbound to open the conversation (empty body,
	// system-style greeting-return message). We use the greeting text
	// as the first CUSTOMER message so agents see context immediately.
	// Actually simpler: leave the thread empty and let the visitor's
	// first real message create the message row. But we still need a
	// Conversation row so /widget/stream can subscribe.

	// Kick a placeholder inbound so a Conversation exists — text is a
	// zero-width marker the frontend filters out on render.
	wireMu.RLock()
	recorder := inboundRecorder
	wireMu.RUnlock()
	if recorder != nil {
		_ = recorder(cfg.OwnerID, contactID, "[chat started]")
	}

	// Resolve the Conversation ID that was just created.
	// Cheapest way: our injected recorder returns via a hook; but a
	// simpler shortcut is to look up the conversation by contact +
	// channel through another injected fn. Keep it lean — we ask the
	// caller-side wire for the recent conversation.
	convID, err := lookupConv(cfg.OwnerID, contactID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to open conversation"})
		return
	}

	sess := &Session{
		OwnerID:        cfg.OwnerID,
		ContactID:      contactID,
		ConversationID: convID,
	}
	if req.VisitorName != "" {
		sess.VisitorName = &req.VisitorName
	}
	if req.VisitorEmail != "" {
		sess.VisitorEmail = &req.VisitorEmail
	}
	if req.PageURL != "" {
		sess.PageURL = &req.PageURL
	}
	ua := c.GetHeader("User-Agent")
	if ua != "" {
		sess.UserAgent = &ua
	}
	if err := database.DB.Create(sess).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session_id":      sess.ID,
		"conversation_id": convID,
		"greeting":        cfg.Greeting,
	})
}

// ConvLookup returns the most recent webchat conversation for a contact.
type ConvLookup func(ownerID, contactID string) (conversationID string, err error)

var convLookup ConvLookup

func RegisterConvLookup(f ConvLookup) { wireMu.Lock(); convLookup = f; wireMu.Unlock() }

func lookupConv(ownerID, contactID string) (string, error) {
	wireMu.RLock()
	fn := convLookup
	wireMu.RUnlock()
	if fn == nil {
		return "", fmt.Errorf("convLookup not wired")
	}
	return fn(ownerID, contactID)
}

// Message from visitor.
type messageReq struct {
	SessionID string `json:"session_id" binding:"required"`
	Key       string `json:"key" binding:"required"`
	Body      string `json:"body" binding:"required"`
}

func handleMessage(c *gin.Context) {
	var req messageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	cfg, err := FindConfigByKey(database.DB, req.Key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Widget not found"})
		return
	}
	sess, err := GetSession(database.DB, cfg.OwnerID, req.SessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Session not found"})
		return
	}

	wireMu.RLock()
	recorder := inboundRecorder
	wireMu.RUnlock()
	if recorder == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"detail": "Widget not wired"})
		return
	}
	if err := recorder(cfg.OwnerID, sess.ContactID, req.Body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// handleListMessages backfills the visitor's thread on reconnect. The
// visitor's own inbound messages plus every agent outbound reply.
func handleListMessages(c *gin.Context) {
	key := c.Query("key")
	sessionID := c.Query("session_id")
	if key == "" || sessionID == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "key and session_id required"})
		return
	}
	cfg, err := FindConfigByKey(database.DB, key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Widget not found"})
		return
	}
	sess, err := GetSession(database.DB, cfg.OwnerID, sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Session not found"})
		return
	}
	wireMu.RLock()
	lister := messageLister
	wireMu.RUnlock()
	if lister == nil {
		c.JSON(http.StatusOK, []OutboundMsg{})
		return
	}
	msgs, err := lister(sess.ConversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, msgs)
}

// handleStream sends agent replies to the visitor over SSE.
func handleStream(c *gin.Context) {
	key := c.Query("key")
	sessionID := c.Query("session_id")
	if key == "" || sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "key and session_id required"})
		return
	}
	cfg, err := FindConfigByKey(database.DB, key)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "invalid widget"})
		return
	}
	sess, err := GetSession(database.DB, cfg.OwnerID, sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "invalid session"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	wireMu.RLock()
	streamer := outboundStream
	wireMu.RUnlock()
	if streamer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"detail": "Streaming not wired"})
		return
	}

	ch, cancel := streamer(sess.ConversationID)
	defer cancel()

	fmt.Fprint(c.Writer, "event: hello\ndata: {\"ok\":true}\n\n")
	c.Writer.Flush()

	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()

	notify := c.Request.Context().Done()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-notify:
			return false
		case <-heartbeat.C:
			fmt.Fprintf(w, ": ping\n\n")
			return true
		case msg, ok := <-ch:
			if !ok {
				return false
			}
			// Only agent replies matter to the visitor
			if msg.Direction != "outbound" {
				return true
			}
			fmt.Fprintf(w, "event: message\ndata: {\"id\":%q,\"body\":%q,\"created_at\":%q}\n\n",
				msg.ID, msg.Body, msg.CreatedAt.Format(time.RFC3339))
			return true
		}
	})
}
