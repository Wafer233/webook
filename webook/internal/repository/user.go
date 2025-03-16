package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

var (
	ErrUserDuplicated = dao.ErrUserDuplicated
	ErrUserNotFound   = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
}

type CacheUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDAO, c cache.UserCache) UserRepository {
	return &CacheUserRepository{
		dao:   dao,
		cache: c,
	}
}

func (r *CacheUserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (r *CacheUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {

	u, err := r.dao.FindByEmail(ctx, email)

	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func (r *CacheUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.cache.Get(ctx, id)

	//1.缓存里头有数据
	if err == nil {
		return u, nil
	}

	//没数据

	//	去数据库加载
	ue, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	u = domain.User{
		Id:       ue.Id,
		Email:    ue.Email,
		Password: ue.Password,
	}

	//go func() {
	//	err = r.cache.Set(ctx, u)
	//	if err != nil {
	//	}
	//}()
	_ = r.cache.Set(ctx, u)

	return u, nil

	//如果redis崩了呢？
	//我要加载，但是为了防止数据库也跟着崩，我需要限流

	//缓存出错
}

//func (repo *CachedUserRepository) UpdateNonZeroFields(ctx context.Context,
//	user domain.User) error {
//	return repo.dao.UpdateById(ctx, repo.toEntity(user))
//}
