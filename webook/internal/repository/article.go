package repository

import (
	"context"
	"gorm.io/gorm"
	"project/webook/internal/domain"
	"project/webook/internal/repository/dao"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, aid int64, status domain.ArticleStatus) error
}

type CachedArticleRepository struct {
	dao dao.ArticleDAO

	// sync v1  with 2-dao
	readDAO dao.ReaderDAO
	authDAO dao.AuthorDAO

	// sync v2
	db *gorm.DB
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, id int64, aid int64, status domain.ArticleStatus) error {
	return c.dao.SyncStatus(ctx, id, aid, status.ToUint8())
}

func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Insert(ctx, dao.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
}

func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	return c.dao.UpdateById(ctx, dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {

	return c.dao.Sync(ctx, c.toEntity(art))
}

func (c *CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	//操作两个dao
	var id = art.Id
	var err error
	var artDAO = c.toEntity(art)
	if id > 0 {
		err = c.authDAO.UpdateById(ctx, artDAO)
	} else {
		id, err = c.authDAO.Insert(ctx, artDAO)
	}
	if err != nil {
		return id, err
	}
	err = c.readDAO.Upsert(ctx, artDAO)
	return id, err
}

func (c *CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	defer tx.Rollback()

	author := dao.NewAuthorDAO(tx)
	reader := dao.NewReaderDAO(tx)

	var id = art.Id
	var err error
	var artDAO = c.toEntity(art)
	if id > 0 {
		err = author.UpdateById(ctx, artDAO)
	} else {
		id, err = author.Insert(ctx, artDAO)
	}
	if err != nil {
		return id, err
	}
	err = reader.UpsertV2(ctx, dao.PublishedArticle{Article: artDAO})
	tx.Commit()
	return id, err

}

func (c *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func NewArticleRepository(dao dao.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{
		dao: dao,
	}
}
