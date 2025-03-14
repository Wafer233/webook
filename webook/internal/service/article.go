package service

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, art domain.Article) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)

	//PublishV1(ctx context.Context, art domain.Article) (int64, error)
}

type articleService struct {
	repo repository.ArticleRepository

	//v1
	authRepo repository.ArticleAuthorRepository
	readRepo repository.ArticleReaderRepository

	log logger.LoggerV1
}

func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
	}
}

func NewArticleServiceV1(authRepo repository.ArticleAuthorRepository, readRepo repository.ArticleReaderRepository, log logger.LoggerV1) ArticleService {
	return &articleService{
		authRepo: authRepo,
		readRepo: readRepo,
		log:      log,
	}
}

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, art)
}

func (a *articleService) Withdraw(ctx context.Context, art domain.Article) error {
	return a.repo.SyncStatus(ctx, art.Id, art.Author.Id, domain.ArticleStatusPrivate)
}

func (a *articleService) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.GetByAuthor(ctx, uid, offset, limit)
}

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetByID(ctx, id)
}

func (a *articleService) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetPubById(ctx, id)
}

//func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
//	var id = art.Id
//	var err error
//
//	if art.Id > 0 {
//		err = a.authRepo.Update(ctx, art)
//	} else {
//		id, err = a.authRepo.Create(ctx, art)
//	}
//	if err != nil {
//		return 0, err
//	}
//	art.Id = id
//
//	for i := 0; i < 3; i++ {
//		id, err = a.readRepo.Save(ctx, art)
//		if err == nil {
//			break
//		}
//		a.log.Error("save fail for reader db",
//			logger.Int64("article_id", art.Id),
//			logger.Error(err))
//	}
//	//
//	if err != nil {
//		a.log.Error("ALL save fail for reader db",
//			logger.Field{Key: "article_id", Value: art.Id},
//			logger.Field{Key: "error", Value: err})
//	}
//
//	return id, err
//}
