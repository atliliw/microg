package auth

import (
	"microg/examples/code"
	"microg/pkg/errors"
	"net/http"
	"strings"

	"microg/server/restserver/middlewares"

	"github.com/gin-gonic/gin"
)

const authHeaderCount = 2

// AutoStrategy defines authentication strategy which can automatically choose between Basic and Bearer
// according `Authorization` header.
type AutoStrategy struct {
	basic BasicStrategy
	jwt   JWTStrategy
}

var _ middlewares.AuthStrategy = &AutoStrategy{}

// NewAutoStrategy create auto strategy with basic strategy and jwt strategy.
func NewAutoStrategy(basic BasicStrategy, jwt JWTStrategy) AutoStrategy {
	return AutoStrategy{
		basic: basic,
		jwt:   jwt,
	}
}

// AuthFunc defines auto strategy as the gin authentication middleware.
func (a AutoStrategy) AuthFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		operator := middlewares.AuthOperator{}
		authHeader := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)

		if len(authHeader) != authHeaderCount {
			middlewares.WriteError(
				c,
				http.StatusUnauthorized,
				errors.WithCode(code.ErrInvalidAuthHeader, "Authorization header format is wrong."),
			)
			c.Abort()

			return
		}

		switch authHeader[0] {
		case "Basic":
			operator.SetStrategy(a.basic)
		case "Bearer":
			operator.SetStrategy(a.jwt)
		default:
			middlewares.WriteError(c, http.StatusUnauthorized, errors.WithCode(code.ErrSignatureInvalid, "unrecognized Authorization header."))
			c.Abort()

			return
		}

		operator.AuthFunc()(c)

		c.Next()
	}
}
