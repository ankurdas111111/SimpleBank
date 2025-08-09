package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ankurdas111111/simplebank/token"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

// authMiddleware creates a gin middleware for authorization
func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)

		// Check if the authorization header is provided at all.
		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			// Use AbortWithStatusJSON to send the error and stop the request chain.
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// The header should be in the format "Bearer <token>".
		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			// Abort the request.
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// Check if the authorization type is "bearer".
		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			// Abort the request.
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// Verify the token itself.
		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			// Abort the request.
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// If everything is okay, set the payload in the context and proceed.
		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
