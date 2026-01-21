package api

import (
	"net/http"
	"sort"
	"time"

	db "github.com/ankurdas111111/simplebank/db/sqlc"
	"github.com/ankurdas111111/simplebank/token"
	"github.com/gin-gonic/gin"
)

type listTransfersRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=1,max=50"`
}

type transferHistoryItem struct {
	ID          int64     `json:"id"`
	FromAccount int64     `json:"from_account_id"`
	ToAccount   int64     `json:"to_account_id"`
	Amount      int64     `json:"amount"`
	FromCurrency string   `json:"from_currency"`
	ToCurrency   string   `json:"to_currency"`
	CreatedAt   time.Time `json:"created_at"`
}

// listTransfers returns transfer history for the authenticated user.
// Implementation note: transfers table doesn't store owner, so we:
// - list the user's accounts
// - fetch recent transfers per account
// - merge + de-duplicate + sort by created_at desc
func (server *Server) listTransfers(ctx *gin.Context) {
	var req listTransfersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Fetch all accounts for this owner (no handler-level page_size constraints here).
	accounts, err := server.store.ListAccounts(ctx, db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	accountCurrency := make(map[int64]string, len(accounts))
	owned := make(map[int64]struct{}, len(accounts))
	for _, a := range accounts {
		accountCurrency[a.ID] = a.Currency
		owned[a.ID] = struct{}{}
	}

	// We over-fetch up to page_id * page_size per account and then slice globally.
	need := int(req.PageID * req.PageSize)
	if need < 1 {
		need = 1
	}

	seen := make(map[int64]transferHistoryItem)
	for _, a := range accounts {
		transfers, err := server.store.ListTransfers(ctx, db.ListTransfersParams{
			FromAccountID: a.ID,
			ToAccountID:   a.ID,
			Limit:         int32(need),
			Offset:        0,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		for _, t := range transfers {
			// Deduplicate by transfer id across multiple accounts.
			if _, ok := seen[t.ID]; ok {
				continue
			}

			fromCur := accountCurrency[t.FromAccountID]
			toCur := accountCurrency[t.ToAccountID]
			// If transfer involves other user's account, currency won't be in map; resolve once.
			if fromCur == "" {
				acc, err := server.store.GetAccount(ctx, t.FromAccountID)
				if err == nil {
					fromCur = acc.Currency
					accountCurrency[t.FromAccountID] = fromCur
				}
			}
			if toCur == "" {
				acc, err := server.store.GetAccount(ctx, t.ToAccountID)
				if err == nil {
					toCur = acc.Currency
					accountCurrency[t.ToAccountID] = toCur
				}
			}

			seen[t.ID] = transferHistoryItem{
				ID:           t.ID,
				FromAccount:  t.FromAccountID,
				ToAccount:    t.ToAccountID,
				Amount:       t.Amount,
				FromCurrency: fromCur,
				ToCurrency:   toCur,
				CreatedAt:    t.CreatedAt,
			}
		}
	}

	items := make([]transferHistoryItem, 0, len(seen))
	for _, it := range seen {
		// Only include transfers that touch an owned account (belt-and-suspenders).
		if _, ok := owned[it.FromAccount]; ok {
			items = append(items, it)
			continue
		}
		if _, ok := owned[it.ToAccount]; ok {
			items = append(items, it)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	offset := int((req.PageID - 1) * req.PageSize)
	if offset > len(items) {
		ctx.JSON(http.StatusOK, []transferHistoryItem{})
		return
	}
	end := offset + int(req.PageSize)
	if end > len(items) {
		end = len(items)
	}

	ctx.JSON(http.StatusOK, items[offset:end])
}


