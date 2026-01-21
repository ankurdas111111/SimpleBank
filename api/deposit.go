package api

import (
	"database/sql"
	"errors"
	"net/http"

	db "github.com/ankurdas111111/simplebank/db/sqlc"
	"github.com/ankurdas111111/simplebank/token"
	"github.com/gin-gonic/gin"
)

type depositRequest struct {
	ID     int64 `uri:"id" binding:"required,min=1"`
	Amount int64 `json:"amount" binding:"required,gt=0"`
}

func (server *Server) deposit(ctx *gin.Context) {
	var uriReq struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var bodyReq struct {
		Amount int64 `json:"amount" binding:"required,gt=0"`
	}
	if err := ctx.ShouldBindJSON(&bodyReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Ensure account exists and belongs to authenticated user.
	account, err := server.store.GetAccount(ctx, uriReq.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if account.Owner != authPayload.Username {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("account doesn't belong to the authenticated user")))
		return
	}

	updated, err := server.store.UpdateAccountBalance(ctx, db.UpdateAccountBalanceParams{
		ID:      uriReq.ID,
		Balance: bodyReq.Amount,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, updated)
}


