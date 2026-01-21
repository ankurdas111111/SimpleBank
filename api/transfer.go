package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	db "github.com/ankurdas111111/simplebank/db/sqlc"
	"github.com/ankurdas111111/simplebank/token"
	"github.com/ankurdas111111/simplebank/util"
	"github.com/gin-gonic/gin"
)



type transferRequest struct{
	FromAccountID   int64 `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   	int64 `json:"to_account_id" binding:"required,min=1"`
	Amount   		int64 `json:"amount" binding:"required,gt=0"`
	// Optional: kept for backward compatibility. If provided, it must match the
	// source account currency.
	Currency 		string `json:"currency" binding:"omitempty,currency"`
	// Optional: for transfers to other users, UI can send a recipient username
	// to validate account_id + username match.
	ToUsername 		string `json:"to_username" binding:"omitempty"`
}


func (server *Server) createTransfer(ctx *gin.Context){
	var req transferRequest
	err:= ctx.ShouldBindJSON(&req); 
	if err!=nil{
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return 
	}

	// From account must belong to the authenticated user.
	fromAccount, err := server.store.GetAccount(ctx, req.FromAccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if fromAccount.Owner != authPayload.Username {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("from account doesn't belong to the authenticated user")))
		return
	}

	// If request specifies currency, ensure it matches source account.
	if req.Currency != "" && fromAccount.Currency != req.Currency {
		ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("source account currency mismatch: %s vs %s", fromAccount.Currency, req.Currency)))
		return
	}

	toAccount, err := server.store.GetAccount(ctx, req.ToAccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// If provided, validate recipient username matches the destination account owner.
	if req.ToUsername != "" && toAccount.Owner != req.ToUsername {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("recipient username does not match destination account")))
		return
	}

	// Same-currency: old path. Cross-currency: convert and credit converted amount.
	if fromAccount.Currency == toAccount.Currency {
		arg := db.TransferTxParams{
			FromAccountID: req.FromAccountID,
			ToAccountID:   req.ToAccountID,
			Amount:        req.Amount,
		}
		result, err := server.store.TransferTx(ctx, arg)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusOK, result)
		return
	}

	toAmount, rate, ok := util.ConvertAmount(req.Amount, fromAccount.Currency, toAccount.Currency)
	if !ok {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("unsupported currency conversion")))
		return
	}
	if toAmount <= 0 {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("amount too small for conversion")))
		return
	}

	result, err := server.store.TransferTxFX(ctx, db.TransferTxFXParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		FromAmount:    req.Amount,
		ToAmount:      toAmount,
		Rate:          rate,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, result)
}

// validAccount removed: transfer validation now supports cross-currency and enforces ownership.