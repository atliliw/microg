package auth

import (
	"encoding/base64"
	"microg/examples/code"
	"microg/pkg/errors"
	"net/http"
	"strings"

	"microg/server/restserver/middlewares"

	"github.com/gin-gonic/gin"
)

// BasicStrategy defines Basic authentication strategy.
type BasicStrategy struct {
	compare func(username string, password string) bool
}

var _ middlewares.AuthStrategy = &BasicStrategy{}

// NewBasicStrategy create basic strategy with compare function.
func NewBasicStrategy(compare func(username string, password string) bool) BasicStrategy {
	return BasicStrategy{
		compare: compare,
	}
}

// AuthFunc defines basic strategy as the gin authentication middleware.
func (b BasicStrategy) AuthFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			middlewares.WriteError(
				c,
				http.StatusUnauthorized,
				errors.WithCode(code.ErrSignatureInvalid, "Authorization header format is wrong."),
			)
			c.Abort()

			return
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)

		if len(pair) != 2 || !b.compare(pair[0], pair[1]) {
			middlewares.WriteError(
				c,
				http.StatusUnauthorized,
				errors.WithCode(code.ErrSignatureInvalid, "Authorization header format is wrong."),
			)
			c.Abort()

			return
		}

		c.Set(middlewares.UsernameKey, pair[0])

		c.Next()
	}
}
