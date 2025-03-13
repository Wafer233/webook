package web

import (
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
	"webook/internal/domain"
	"webook/internal/service"
	"webook/pkg/ginx"
	"webook/pkg/logger"
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
	group.POST("list", ginx.WrapBodyAndToken[ListReq, UserClaims](handler.List))
	group.POST("/detail/:id", ginx.WrapToken[UserClaims](handler.Detail))

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

func (handler *ArticleHandler) List(ctx *gin.Context, req ListReq, uc UserClaims) (ginx.Result, error) {
	res, err := handler.svc.List(ctx, uc.Uid, req.Offset, req.Limit)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "system error",
		}, nil
	}

	return ginx.Result{
		Data: slice.Map[domain.Article, ArticleVO](res,
			func(idx int, src domain.Article) ArticleVO {
				return ArticleVO{
					Id:    src.Id,
					Title: src.Title,
					//Content:  src.Content,
					Abstract: src.Abstract(),
					//Author:   src.Author.Name,
					Ctime: src.Ctime.Format(time.DateTime),
					Utime: src.Utime.Format(time.DateTime),
				}
			}),
	}, nil
}

func (handler *ArticleHandler) Detail(ctx *gin.Context, uc UserClaims) (ginx.Result, error) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		//h.l.Error("id input error", logger.Error(err))
		return ginx.Result{
			Code: 4,
			Msg:  "参数错误",
		}, err
	}
	art, err := handler.svc.GetById(ctx, id)
	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "system error",
		}, err
	}
	// 不借助数据库查询来判断
	if art.Author.Id != uc.Uid {
		return ginx.Result{
			Code: http.StatusBadRequest,
			Msg:  "input error",
		}, fmt.Errorf("非法访问文章，创作者 ID 不匹配 %d", uc.Uid)
	}
	return ginx.Result{
		Data: ArticleVO{
			Id:       art.Id,
			Title:    art.Title,
			Abstract: art.Abstract(),
			Content:  art.Content,
			Status:   art.Status.ToUint8(),
			Ctime:    art.Ctime.Format(time.DateTime),
			Utime:    art.Utime.Format(time.DateTime),
		},
	}, nil
}
