//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	"webook/ioc"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	InitRedis, InitDB,
	InitLogger)

var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewUserCache,
	repository.NewUserRepository,
	service.NewUserService)

var articlSvcProvider = wire.NewSet(
	//	//cache.NewArticleRedisCache,
	dao.NewGORMArticleDAO,
	repository.NewArticleRepository,
	service.NewArticleService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		articlSvcProvider,

		// cache 部分
		//cache.NewCodeCache,

		// repository 部分
		//repository.NewCodeRepository,

		// Service 部分
		//ioc.InitSMSService,
		//service.NewCodeService,
		//InitWechatService,

		// handler 部分
		web.NewUserHandler,
		web.NewArticleHandler,
		//web.NewOAuth2WechatHandler,
		//ijwt.NewRedisJWTHandler,
		ioc.InitMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}

func InitArticleHandler() *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		dao.NewGORMArticleDAO,
		service.NewArticleService,
		repository.NewArticleRepository,
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}

//		cache.NewArticleRedisCache,
//		service.NewArticleService,
//		web.NewArticleHandler)
//	return &web.ArticleHandler{}
//}
