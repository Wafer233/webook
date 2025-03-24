package web

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
	"webook/internal/domain"
	"webook/internal/service"
	"webook/pkg/logger"
)

//

type ArticleHandler struct {
	svc      service.ArticleService
	interSvc service.InteractiveService
	biz      string

	log logger.LoggerV1
}

func NewArticleHandler(svc service.ArticleService, interSvc service.InteractiveService, log logger.LoggerV1) *ArticleHandler {
	return &ArticleHandler{
		svc:      svc,
		log:      log,
		interSvc: interSvc,
		biz:      "articles",
	}
}

func (handler *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	group := server.Group("/articles")
	group.POST("withdraw", handler.Withdraw)
	group.POST("edit", handler.Edit)
	group.POST("publish", handler.Publish)

	// 创作者接口
	group.GET("/detail/:id", handler.Detail)
	// 按照道理来说，这边就是 GET 方法
	// /list?offset=?&limit=?
	group.POST("/list", handler.List)

	pub := group.Group("/pub")
	pub.GET("/:id", handler.PubDetail)

	// 传入一个参数，true 就是点赞, false 就是不点赞
	pub.POST("/like", handler.Like)
	pub.POST("/collect", handler.Collect)
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

func (handler *ArticleHandler) List(ctx *gin.Context) {
	var page Page
	if err := ctx.Bind(&page); err != nil {
		return
	}
	// 我要不要检测一下？
	uc := ctx.MustGet("user").(UserClaims)
	arts, err := handler.svc.GetByAuthor(ctx, uc.Uid, page.Offset, page.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		handler.log.Error("查找文章列表失败",
			logger.Error(err),
			logger.Int("offset", page.Offset),
			logger.Int("limit", page.Limit),
			logger.Int64("uid", uc.Uid))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: slice.Map[domain.Article, ArticleVO](arts, func(idx int, src domain.Article) ArticleVO {
			return ArticleVO{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract(),

				//Content:  src.Content,
				AuthorId: src.Author.Id,
				// 列表，你不需要
				Status: src.Status.ToUint8(),
				Ctime:  src.Ctime.Format(time.DateTime),
				Utime:  src.Utime.Format(time.DateTime),
			}
		}),
	})
}

func (handler *ArticleHandler) Detail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "id 参数错误",
			Code: 4,
		})
		handler.log.Warn("查询文章失败，id 格式不对",
			logger.String("id", idstr),
			logger.Error(err))
		return
	}
	art, err := handler.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		handler.log.Error("查询文章失败",
			logger.Int64("id", id),
			logger.Error(err))
		return
	}
	uc := ctx.MustGet("user").(UserClaims)
	if art.Author.Id != uc.Uid {
		// 有人在搞鬼
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		handler.log.Error("非法查询文章",
			logger.Int64("id", id),
			logger.Int64("uid", uc.Uid))
		return
	}

	vo := ArticleVO{
		Id:    art.Id,
		Title: art.Title,
		//Abstract: art.Abstract(),

		Content:  art.Content,
		AuthorId: art.Author.Id,
		// 列表，你不需要
		Status: art.Status.ToUint8(),
		Ctime:  art.Ctime.Format(time.DateTime),
		Utime:  art.Utime.Format(time.DateTime),
	}
	ctx.JSON(http.StatusOK, Result{Data: vo})
}

func (handler *ArticleHandler) PubDetail(ctx *gin.Context) {

	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "id 参数错误",
			Code: 4,
		})
		handler.log.Warn("查询文章失败，id 格式不对",
			logger.String("id", idstr),
			logger.Error(err))
		return
	}

	var (
		eg   errgroup.Group
		art  domain.Article
		intr domain.Interactive
	)
	uc := ctx.MustGet("user").(UserClaims)
	eg.Go(func() error {

		var er error
		art, er = handler.svc.GetPubById(ctx, id, uc.Uid)
		return er
	})

	//uc := ctx.MustGet("user").(UserClaims)
	//eg.Go(func() error {
	//	var er error
	//	intr, er = handler.interSvc.Get(ctx, handler.biz, id, uc.Uid)
	//	return er
	//})

	// 等待结果
	err = eg.Wait()
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		handler.log.Error("查询文章失败，系统错误",
			logger.Int64("aid", id),
			logger.Int64("uid", uc.Uid),
			logger.Error(err))
		return
	}

	go func() {
		// 1. 如果你想摆脱原本主链路的超时控制，你就创建一个新的
		// 2. 如果你不想，你就用 ctx
		newCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := handler.interSvc.IncrReadCnt(newCtx, handler.biz, art.Id)
		if er != nil {
			handler.log.Error("更新阅读数失败",
				logger.Int64("aid", art.Id),
				logger.Error(err))
		}
	}()

	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVO{
			Id:    art.Id,
			Title: art.Title,

			Content:    art.Content,
			AuthorId:   art.Author.Id,
			AuthorName: art.Author.Name,
			ReadCnt:    intr.ReadCnt,
			CollectCnt: intr.CollectCnt,
			LikeCnt:    intr.LikeCnt,
			Liked:      intr.Liked,
			Collected:  intr.Collected,

			Status: art.Status.ToUint8(),
			Ctime:  art.Ctime.Format(time.DateTime),
			Utime:  art.Utime.Format(time.DateTime),
		},
	})
}

func (handler *ArticleHandler) Like(c *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
		// true 是点赞，false 是不点赞
		Like bool `json:"like"`
	}
	var req Req
	if err := c.Bind(&req); err != nil {
		return
	}
	uc := c.MustGet("user").(UserClaims)
	var err error
	if req.Like {
		// 点赞
		err = handler.interSvc.Like(c, handler.biz, req.Id, uc.Uid)
	} else {
		// 取消点赞
		err = handler.interSvc.CancelLike(c, handler.biz, req.Id, uc.Uid)
	}
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Code: 5, Msg: "系统错误",
		})
		handler.log.Error("点赞/取消点赞失败",
			logger.Error(err),
			logger.Int64("uid", uc.Uid),
			logger.Int64("aid", req.Id))
		return
	}
	c.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (handler *ArticleHandler) Collect(ctx *gin.Context) {
	type Req struct {
		Id  int64 `json:"id"`
		Cid int64 `json:"cid"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(UserClaims)

	err := handler.interSvc.Collect(ctx, handler.biz, req.Id, req.Cid, uc.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5, Msg: "系统错误",
		})
		handler.log.Error("收藏失败",
			logger.Error(err),
			logger.Int64("uid", uc.Uid),
			logger.Int64("aid", req.Id))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

type Page struct {
	Limit  int
	Offset int
}
