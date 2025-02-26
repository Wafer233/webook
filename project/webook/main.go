package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"project/webook/internal/integration/startup"
)

func main() {
	server := startup.InitWebServer()
	server.GET("hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hello world me")
	})
	server.Run(":8080")

}

//
//db := initDB()
//server := initWebServer()
//
//u := initUser(db)
//u.RegisterRoutes(server)
//
////server := gin.Default()
//server.GET("hello", func(c *gin.Context) {
//	c.String(http.StatusOK, "hello world me")
//})
//server.Run(":8080")
//server := InitWebServer()
//
//server.GET("hello", func(c *gin.Context) {
//	c.String(http.StatusOK, "hello world me")
//})
//server.Run(":8080")
//
//func initWebServer() *gin.Engine {
//	server := gin.Default()

//解决CORS问题
//server.Use(cors.New(cors.Config{
//
//	AllowHeaders:     []string{"authorization", "content-type"},
//	AllowCredentials: true,
//	ExposeHeaders:    []string{"x-jwt-token"},
//	AllowOriginFunc: func(origin string) bool {
//		if strings.HasPrefix(origin, "http://localhost") {
//			return true
//		}
//		return origin == "https://github.com"
//	},
//	MaxAge: 12 * time.Hour,
//}))

//解决session问题

//基于cookie 不好！
//store := cookie.NewStore([]byte("secret"))

////v1 based on memstore
//store := memstore.NewStore(
//	[]byte("a3f8d2e7c6b5a4d1e9f0c3b2d7e8a1f6"),
//	[]byte("f2d9e3c7b4a1f5d8e0c6b3a7d1f4e9a2"),
//)

//store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
//	[]byte("a3f8d2e7c6b5a4d1e9f0c3b2d7e8a1f6"),
//	[]byte("f2d9e3c7b4a1f5d8e0c6b3a7d1f4e9a2"))
//
//if err != nil {
//	panic(err)
//}
//redisClient := redis.NewClient(&redis.Options{
//	//Addr: "webook-redis:6381",
//	Addr: config.Config.Redis.Addr,
//})
//
//server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())
//server.Use(sessions.Sessions("session_id", store))
//server.Use(middleware.NewLoginMiddlewareBuilder().IgnorePath("/users/login").
//	IgnorePath("/users/signup").CheckLogin())

//server.Use(middleware.NewLoginMiddlewareBuilder().CheckLogin())

//server.Use(sessions.Sessions("session_id", store))
//server.Use(middleware.NewLoginJWTMiddlewareBuilder().
//	IgnorePath("/users/login").IgnorePath("/users/signup").CheckLogin())
//
//return server

//}

//func initUser(db *gorm.DB) *web.UserHandler {
//	ud := dao.NewUserDAO(db)
//	repo := repository.NewUserRepository(ud)
//	svc := service.NewUserService(repo)
//	u := web.NewUserHandler(svc)
//	return u
//}

//func initDB() *gorm.DB {
//	//初始化数据库
//	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
//	//db, err := gorm.Open(mysql.Open("root:root@tcp(webook-mysql:3308)/webook"))
//	if err != nil {
//		panic(err)
//	}
//
//	err = dao.InitTable(db)
//	if err != nil {
//		panic(err)
//	}
//
//	return db
//}
