package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type lookupAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type lookupAccountResponse struct {
	ID       int64  `json:"id"`
	Owner    string `json:"owner"`
	Currency string `json:"currency"`
}

// lookupAccount returns limited metadata needed for transfers to other users.
// It intentionally does not return balance or timestamps.
func (server *Server) lookupAccount(ctx *gin.Context) {
	var req lookupAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, lookupAccountResponse{
		ID:       account.ID,
		Owner:    account.Owner,
		Currency: account.Currency,
	})
}


