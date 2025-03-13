package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	"webook/internal/web"
	"webook/internal/web/middleware"
	"webook/pkg/ratelimit"
)

// articleHdl *web.ArticleHandler
func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler, articleHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHdlr(),
		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePath("/users/login").
			IgnorePath("/users/signup").
			Build(),

		ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
	}
}

func corsHdlr() gin.HandlerFunc {
	println("err")
	return cors.New(cors.Config{

		AllowHeaders:     []string{"authorization", "content-type"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"x-jwt-token"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	})
}
