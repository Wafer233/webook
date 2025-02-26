package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"project/webook/internal/domain"
	"project/webook/internal/service"
	"project/webook/pkg/logger"
)

//

type ArticleHandler struct {
	svc service.ArticleService
	log logger.LoggerV1
}

func NewArticleHandler(svc service.ArticleService, log logger.LoggerV1) *ArticleHandler {
	return &ArticleHandler{
		svc: svc,
		log: log,
	}
}

func (handler *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	group := server.Group("/articles")
	group.POST("withdraw", handler.Withdraw)
	group.POST("edit", handler.Edit)
	group.POST("publish", handler.Publish)
}

func (handler *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("claims")
	claims, ok := uc.(*UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "system error",
		})
		handler.log.Error("session no found")
		return
	}

	err := handler.svc.Withdraw(ctx, domain.Article{
		Id: req.Id,
		Author: domain.Author{
			Id: claims.Uid,
		},
	})

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "system error",
		})
		//log
		handler.log.Error("Publish fail",
			//logger.Int64("uid", uc.Uid),
			logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Msg: "ok",
	})

}

func (handler *ArticleHandler) Publish(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	//uc := ctx.MustGet("claims").(*UserClaims)
	uc := ctx.MustGet("claims")
	claims, ok := uc.(*UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		handler.log.Error("未发现session")
		return
	}

	id, err := handler.svc.Publish(ctx, req.toDomain(claims.Uid))

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "system error",
		})
		//log
		handler.log.Error("Publish fail",
			//logger.Int64("uid", uc.Uid),
			logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Data: id,
	})
}

func (handler *ArticleHandler) Edit(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	//uc := ctx.MustGet("claims").(*UserClaims)
	uc := ctx.MustGet("claims")
	claims, ok := uc.(*UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		handler.log.Error("未发现session")
		return
	}

	id, err := handler.svc.Save(ctx, req.toDomain(claims.Uid))

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		//log
		handler.log.Error("Save fail",
			//logger.Int64("uid", uc.Uid),
			logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Data: id,
	})
}

type Article struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (req ArticleReq) toDomain(uid int64) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
	}
}
