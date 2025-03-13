package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"strings"
	"time"
	"webook/internal/web"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//
		//v1不需要登录校验
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		tokenHeader := ctx.GetHeader("Authorization")
		if tokenHeader == "" {
			//	没登陆
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		segs := strings.SplitN(tokenHeader, " ", 2)
		if len(segs) != 2 {
			//	没登陆
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := segs[1]
		claims := &web.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("f2d9e3c7b4a1f5d8e0c6b3a7d1f4e9a2"), nil
		})

		if err != nil {
			//	没登陆
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if token == nil || !token.Valid || claims.Uid == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims.UserAgent != ctx.Request.UserAgent() {
			//	严重的安全问题
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}

		//10秒刷新一次

		now := time.Now()
		if claims.ExpiresAt.Sub(now) < time.Second*50 {

			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute * 1))
			tokenStr, err = token.SignedString([]byte("f2d9e3c7b4a1f5d8e0c6b3a7d1f4e9a2"))
			if err != nil {
				log.Print("jwt signing error:", err)
			}

			ctx.Header("x-jwt-token", tokenStr)

		}
		ctx.Set("claims", claims)

	}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePath(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}
