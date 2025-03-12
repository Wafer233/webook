package dao

import (
	"context"
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicated = errors.New("email conflict")
	ErrUserNotFound   = gorm.ErrRecordNotFound
)

type UserDAO interface {
	FindByEmail(ctx context.Context, email string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
	Insert(ctx context.Context, u User) error
}

// 负责数据库对接 要有gorm的标签
type User struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Email    string `gorm:"unique"`
	Password string

	Ctime int64
	Utime int64
}
type GORMUserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {

	return &GORMUserDAO{
		db: db,
	}
}

func (dao *GORMUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {

	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) FindById(ctx context.Context, id int64) (User, error) {

	var u User
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) Insert(ctx context.Context, u User) error {
	//有的公司存毫秒数
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if mysqlError, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlError.Number == uniqueConflictsErrNo {
			return ErrUserDuplicated
		}
	}
	return err
}
