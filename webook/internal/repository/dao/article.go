package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, aid int64, stat uint8) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
	//Transaction(ctx context.Context, bizFunc func(txDAO ArticleDAO) error) error
	//Upsert(ctx context.Context, article PublishedArticle) error
}

type Article struct {
	Id      int64  `gorm:"primary_key,autoIncrement"`
	Title   string `gorm:"type=varchar(1024)"`
	Content string `gorm:"type=BLOB"`

	//SLECT * FROM articles WHERE author_id = 1 ORDER BY `ctime`
	//SLECT * FROM articles WHERE id = 1
	AuthorId int64 `gorm:"index"`
	//AuthorId int64 `gorm:"index=aid_ctime"`
	//Ctime    int64
	Ctime  int64
	Utime  int64
	Status uint8
}
type GORMArticleDAO struct {
	db *gorm.DB
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

func (dao *GORMArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (dao *GORMArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	//art.Ctime = now
	art.Utime = now
	res := dao.db.WithContext(ctx).Model(&art).Where("id=? AND author_id=?", art.Id, art.AuthorId).
		Updates(map[string]any{
			"status":  art.Status,
			"title":   art.Title,
			"content": art.Content,
			"utime":   art.Utime,
		})

	err := res.Error
	if err != nil {
		return err
	}

	if res.RowsAffected == 0 {
		return errors.New("更新数据失败")
	}

	return nil
}

func (dao *GORMArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var (
			err error
		)
		dao := NewGORMArticleDAO(tx)
		if id > 0 {
			err = dao.UpdateById(ctx, art)
		} else {
			id, err = dao.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id
		now := time.Now().UnixMilli()
		pubArt := PublishedArticle(art)
		pubArt.Ctime = now
		pubArt.Utime = now
		err = tx.Clauses(clause.OnConflict{
			// 对MySQL不起效，但是可以兼容别的方言
			// INSERT xxx ON DUPLICATE KEY SET `title`=?
			// 别的方言：
			// sqlite INSERT XXX ON CONFLICT DO UPDATES WHERE
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":   pubArt.Title,
				"content": pubArt.Content,
				"utime":   now,
				"status":  pubArt.Status,
			}),
		}).Create(&pubArt).Error
		return err
	})
	return id, err
}

func (dao *GORMArticleDAO) SyncStatus(ctx context.Context, uid int64, aid int64, stat uint8) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id = ? and author_id = ?", uid, aid).
			Updates(map[string]any{
				"utime":  now,
				"status": stat,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return errors.New("ID 不对或者创作者不对")
		}
		return tx.Model(&PublishedArticle{}).
			Where("id = ?", uid).
			Updates(map[string]any{
				"utime":  now,
				"status": stat,
			}).Error
	})
}

func (dao *GORMArticleDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	err := dao.db.WithContext(ctx).
		Where("author_id = ?", uid).
		Offset(offset).Limit(limit).
		// a ASC, B DESC
		Order("utime DESC").
		Find(&arts).Error
	return arts, err
}

func (dao *GORMArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := dao.db.WithContext(ctx).
		Where("id = ?", id).First(&art).Error
	return art, err
}

func (dao *GORMArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var res PublishedArticle
	err := dao.db.WithContext(ctx).
		Where("id = ?", id).
		First(&res).Error
	return res, err
}

//func (dao *GORMArticleDAO) Transaction(ctx context.Context, bizFunc func(txDAO ArticleDAO) error) error {
//	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
//		txDAO := NewGORMArticleDAO(tx)
//		return bizFunc(txDAO)
//
//	})
//}

//func (dao *GORMArticleDAO) Upsert(ctx context.Context, article PublishedArticle) error {
//	now := time.Now().UnixMilli()
//	article.Ctime = now
//	article.Utime = now
//	err := dao.db.WithContext(ctx).Clauses(clause.OnConflict{
//		DoUpdates: clause.Assignments(map[string]interface{}{
//			"title":   article.Title,
//			"content": article.Content,
//			"utime":   now,
//			"status":  article.Status,
//		}),
//	}).Create(&article).Error
//	return err
//}
