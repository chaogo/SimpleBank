package api

import (
	"database/sql"
	"errors"
	"net/http"

	db "github.com/chaogo/SimpleBank/db/sqlc"
	"github.com/chaogo/SimpleBank/token"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,currency"` // With ShouldBindJSON, Gin will validate the output object to make sure it satisfy the conditions we specified in the binding tag
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// the client has provided invalid data
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) // send a JSON response. status code for Bad request: 400
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload) // MustGet will return a general interface which need to be casted to token.Payload object so that we can use "authPayload.Username"

	arg := db.CreateAccountParams{
		Owner: authPayload.Username,
		Currency: req.Currency,
		Balance: 0,
	}

	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err)) // 403
				return
			}
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err)) // 500
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil { // ShouldBindUri function to bind all URI parameters into the struct
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) 
		return
	}

	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		}
		
		ctx.JSON(http.StatusInternalServerError, errorResponse(err)) // 500
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if account.Owner != authPayload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type listAccountRequest struct {
	PageID int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listAccount(ctx *gin.Context) {
	var req listAccountRequest
	if err := ctx.ShouldBindQuery(&req); err != nil { // ShouldBindQuery function to bind all query parameters into the struct
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) 
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.ListAccountsParams{
		Owner: authPayload.Username,
		Limit: req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	accounts, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}