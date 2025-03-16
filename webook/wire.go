//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/internal/event"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	"webook/ioc"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	ioc.InitRedis,
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitSaramaClient,
	ioc.InitSyncProducer,
	ioc.InitConsumers,
)

var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewUserCache,
	repository.NewUserRepository,
	service.NewUserService)

var articlSvcProvider = wire.NewSet(
	dao.NewGORMArticleDAO,
	repository.NewArticleRepository,
	service.NewArticleService,
	cache.NewRedisArticleCache,

	dao.NewGORMInteractiveDAO,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
	cache.NewInteractiveRedisCache,

	event.NewInteractiveReadEventConsumer,
	event.NewSaramaSyncProducer,
)

func InitWebServer() *App {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		articlSvcProvider,

		// handler 部分
		web.NewUserHandler,
		web.NewArticleHandler,
		ioc.InitMiddlewares,
		ioc.InitWeb,
		wire.Struct(new(App), "*"),
	)

	return new(App)
}
