package middleware

import (
	"encoding/gob"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (l *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//
		//v1不需要登录校验
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		//不需要登录校验
		//if ctx.Request.URL.Path == "/users/login" ||
		//	ctx.Request.URL.Path == "/users/signup" {
		//	return
		//}

		//	需要校验的话
		sess := sessions.Default(ctx)
		id := sess.Get("user_id")

		if id == nil {
			//	没有登陆
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		upDateTime := sess.Get("update_time")
		sess.Set("user_id", id)
		sess.Options(sessions.Options{
			MaxAge: 60,
		})

		now := time.Now()

		//第一次登录没更新
		gob.Register(time.Time{})
		if upDateTime == nil {
			sess.Set("update_time", now)
			if err := sess.Save(); err != nil {
				panic(err)
			}
			return
		}

		//断言
		updateTimeVal, ok := upDateTime.(time.Time)
		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		//60秒刷新一次
		if now.Sub(updateTimeVal) > 10*time.Second {
			sess.Set("update_time", now)
			if err := sess.Save(); err != nil {
				panic(err)
			}
		}

	}
}

func (l *LoginMiddlewareBuilder) IgnorePath(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}
