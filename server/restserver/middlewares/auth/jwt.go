package auth

import (
	ginjwt "github.com/appleboy/gin-jwt/v2"
	"github.com/atliliw/microg/server/restserver/middlewares"
	"github.com/gin-gonic/gin"
)

// AuthzAudience defines the value of jwt audience field.
const AuthzAudience = "microg.example.com"

// JWTStrategy defines jwt bearer authentication strategy.
type JWTStrategy struct {
	ginjwt.GinJWTMiddleware
}

var _ middlewares.AuthStrategy = &JWTStrategy{}

// NewJWTStrategy create jwt bearer strategy with GinJWTMiddleware.
func NewJWTStrategy(gjwt ginjwt.GinJWTMiddleware) JWTStrategy {
	return JWTStrategy{gjwt}
}

// AuthFunc defines jwt bearer strategy as the gin authentication middleware.
func (j JWTStrategy) AuthFunc() gin.HandlerFunc {
	return j.MiddlewareFunc()
}
