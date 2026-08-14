package wallet

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

type createTxnReq struct {
	Type           TransactionType `json:"type" binding:"required"`
	Amount         float64         `json:"amount" binding:"required,gt=0"`
	Reference      *string         `json:"reference"`
	IdempotencyKey string          `json:"idempotency_key" binding:"required"`
}

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/wallet")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleGetWallet)
		g.GET("/transactions", handleListTransactions)
		g.POST("/transactions", handleCreateTransaction)
	}
}

func handleGetWallet(c *gin.Context) {
	userID := auth.GetUserID(c)
	w, err := GetOrCreateWallet(database.DB, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, w)
}

func handleListTransactions(c *gin.Context) {
	userID := auth.GetUserID(c)
	w, err := GetOrCreateWallet(database.DB, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	txns, err := ListTransactions(database.DB, w.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, txns)
}

func handleCreateTransaction(c *gin.Context) {
	var req createTxnReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	userID := auth.GetUserID(c)
	w, err := GetOrCreateWallet(database.DB, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}

	txn, err := ApplyTransaction(database.DB, w, req.Type, req.Amount, req.IdempotencyKey, req.Reference)
	if err == ErrInsufficientBalance {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}

	// Notify post-credit hook (e.g. auto-resume paused campaigns)
	if req.Type == Credit && onCredit != nil {
		onCredit(userID, txn.BalanceAfter)
	}

	c.JSON(http.StatusCreated, txn)
}

// CreditHook is fired after a successful wallet credit.
type CreditHook func(ownerID string, newBalance float64)

var onCredit CreditHook

func RegisterCreditHook(h CreditHook) { onCredit = h }
