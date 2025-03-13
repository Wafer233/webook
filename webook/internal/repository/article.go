package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/pkg/logger"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, aid int64, status domain.ArticleStatus) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetByID(ctx context.Context, id int64) (domain.Article, error)
}

type CachedArticleRepository struct {
	dao dao.ArticleDAO

	// sync v1  with 2-dao
	readDAO dao.ReaderDAO
	authDAO dao.AuthorDAO

	// sync v2
	db *gorm.DB

	cache cache.ArticleCache

	log logger.LoggerV1
}

func NewArticleRepository(dao dao.ArticleDAO, log logger.LoggerV1) ArticleRepository {

	return &CachedArticleRepository{
		dao: dao,
		log: log,
	}
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, id int64, aid int64, status domain.ArticleStatus) error {

	return c.dao.SyncStatus(ctx, id, aid, status.ToUint8())
}

func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	defer func() {
		c.cache.DelFirstPage(ctx, art.Author.Id)
	}()

	return c.dao.Insert(ctx, dao.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
}

func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	defer func() {
		c.cache.DelFirstPage(ctx, art.Author.Id)
	}()

	return c.dao.UpdateById(ctx, dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	defer func() {
		c.cache.DelFirstPage(ctx, art.Author.Id)
	}()

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

func (c *CachedArticleRepository) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {

	// 你在这个地方，集成记得复杂的缓存方案
	// 缓存第一页
	if offset == 0 && limit <= 100 {
		data, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			go func() {
				c.preCache(ctx, data)
			}()
			return data, err
		}
	}

	res, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	data := slice.Map[dao.Article, domain.Article](res, func(idx int, src dao.Article) domain.Article {
		return c.toDomain(src)
	})

	defer func() {
		c.cache.SetFirstPage(ctx, uid, data)
		c.log.Error("回写缓存失败", logger.Error(err))
		c.preCache(ctx, data)

	}()

	return data, nil
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

func (c *CachedArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Status: domain.ArticleStatus(art.Status),
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
	}
}

func (c *CachedArticleRepository) GetByID(ctx context.Context, id int64) (domain.Article, error) {
	data, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return c.toDomain(data), nil
}

func (c *CachedArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	const contentSizeThreshold = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) <= contentSizeThreshold {
		err := c.cache.Set(ctx, arts[0].Id, arts[0])
		if err != nil {
			c.log.Error("提取预加载缓存失败", logger.Error(err))
		}
	}
}
