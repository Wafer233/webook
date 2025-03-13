package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"webook/internal/domain"
	"webook/internal/repository"
)

var (
	ErrUserDuplicatedEmail   = repository.ErrUserDuplicated
	ErrInvalidUserOrPassword = errors.New("invalid email or password")
)

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	LogIn(ctx context.Context, u domain.User) (domain.User, error)
	//Profile(ctx context.Context, id int64) (domain.User, error)
}
type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (svc *userService) SignUp(ctx context.Context, u domain.User) error {
	//	你要考虑加密放在那里
	//	然后就是从存起来
	// 加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	//err = svc.repo.Create(ctx, u)
	//if err != nil {
	//	return err
	//}
	//val, err := json.Marshal(u)
	//if err != nil {
	//	return err
	//}
	//
	//svc.redis.Set(ctx, fmt.Sprintf("user:info:%d", u.Id), val, 30*time.Minute)
	////err = bcrypt.CompareHashAndPassword(hash, []byte(u.Password))
	////assert.NoError(t, err)
	//
	return svc.repo.Create(ctx, u)

}

func (svc *userService) LogIn(ctx context.Context, u domain.User) (domain.User, error) {
	//先找用户

	foundUser, err := svc.repo.FindByEmail(ctx, u.Email)

	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}

	if err != nil {
		return domain.User{}, err
	}

	// 比较密码
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(u.Password))
	if err != nil {
		//debug
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return foundUser, err
}

func (svc *userService) Profile(ctx context.Context, id int64) (domain.User, error) {
	//val, err := svc.redis.Get(ctx, fmt.Sprintf("user:info:%d", id)).Result()
	//if err != nil {
	//	return domain.User{}, err
	//}
	//var u domain.User
	//json.Unmarshal([]byte(val), &u)
	//return u, err
	u, err := svc.repo.FindById(ctx, id)
	return u, err
}

//func (svc *userService) UpdateNonSensitiveInfo(ctx context.Context,
//	user domain.User) error {
//	// UpdateNicknameAndXXAnd
//	return svc.repo.UpdateNonZeroFields(ctx, user)
//}

func (svc *userService) FindById(ctx context.Context, uid int64) (domain.User, error) {
	return svc.repo.FindById(ctx, uid)
}
