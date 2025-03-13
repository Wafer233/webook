package web

import (
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
	"webook/internal/domain"
	"webook/internal/service"
)

const (
	emailRegexPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d).{8,}$`
)

type UserHandler struct {
	svc service.UserService
}

// JWT
type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{
		svc: svc,
	}
}

func (uh *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		ConfirmPassword string `json:"confirmPassword"`
		Email           string `json:"email"`
		Password        string `json:"password"`
	}
	var req SignUpReq

	if err := ctx.Bind(&req); err != nil {
		return
	}

	//emailRegexPattern
	emailReg := regexp.MustCompile(emailRegexPattern, 0)
	isMatch, err := emailReg.MatchString(req.Email)
	//do something
	if err != nil {
		ctx.String(http.StatusOK, "System error")
		return
	}
	if !isMatch {
		ctx.String(http.StatusOK, "Email format error")
		return
	}

	//passwordRegexPattern
	passwordReg := regexp.MustCompile(passwordRegexPattern, 0)

	isMatch, err = passwordReg.MatchString(req.Password)

	//do something
	if err != nil {
		ctx.String(http.StatusOK, "System error")
		return
	}
	if !isMatch {
		ctx.String(http.StatusOK,
			"At least 8 characters in length, containing at least one letter and one number")
		return
	}

	//confirmPassword
	if req.ConfirmPassword != req.Password {
		ctx.String(http.StatusOK, "Password error")
		return
	}

	fmt.Printf("%+v", req)

	err = uh.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	if err == service.ErrUserDuplicatedEmail {
		ctx.String(http.StatusOK, "Email already exists")
		return
	}

	if err != nil {
		ctx.String(http.StatusOK, "System error")
		return
	}

	//登陆成功

	ctx.String(http.StatusOK, "Sign up successful")
	//	数据库操作
}

//func (u *UserHandler) LogIn(ctx *gin.Context) {
//
//	type SignUpReq struct {
//		Email    string `json:"email"`
//		Password string `json:"password"`
//	}
//
//	var req SignUpReq
//
//	if err := ctx.Bind(&req); err != nil {
//		return
//	}
//
//	user, err := u.svc.LogIn(ctx, domain.User{
//		Email:    req.Email,
//		Password: req.Password,
//	})
//
//	if err == service.ErrInvalidUserOrPassword {
//		ctx.String(http.StatusOK, "Invalid email or password")
//		return
//	}
//
//	if err != nil {
//		ctx.String(http.StatusOK, "System error")
//		return
//	}
//
//	//登入成功
//	sess := sessions.Default(ctx)
//	sess.Set("user_id", user.Id)
//	sess.Options(sessions.Options{
//		//30秒过期
//		MaxAge: 60,
//	})
//	sess.Save()
//
//	ctx.String(http.StatusOK, "Sign in successful")
//	return
//}

func (uh *UserHandler) LogInJWT(ctx *gin.Context) {

	type SignUpReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req SignUpReq

	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := uh.svc.LogIn(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "Invalid email or password")
		return
	}

	if err != nil {
		ctx.String(http.StatusOK, "System error")
		return
	}

	//登入成功
	//sess := sessions.Default(ctx)
	//sess.Set("user_id", user.Id)
	//sess.Options(sessions.Options{
	//	//30秒过期
	//	MaxAge: 60,
	//})
	//sess.Save()

	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 1)),
		},
		Uid:       user.Id,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("f2d9e3c7b4a1f5d8e0c6b3a7d1f4e9a2"))
	if err != nil {
		ctx.String(http.StatusInternalServerError, "System error")
	}
	ctx.Header("x-jwt-token", tokenStr)
	fmt.Println(user)
	fmt.Println(tokenStr)

	ctx.String(http.StatusOK, "Sign in successful")
	return
}

func (u *UserHandler) LogOut(ctx *gin.Context) {

	sess := sessions.Default(ctx)

	sess.Options(sessions.Options{MaxAge: -1})

	ctx.String(http.StatusOK, "Sign out successful")
	return
}

func (h *UserHandler) Edit(ctx *gin.Context) {
	//	// 嵌入一段刷新过期时间的代码
	//	type Req struct {
	//		// 改邮箱，密码，或者能不能改手机号
	//
	//		Nickname string `json:"nickname"`
	//		// YYYY-MM-DD
	//		Birthday string `json:"birthday"`
	//		AboutMe  string `json:"aboutMe"`
	//	}
	//	var req Req
	//	if err := ctx.Bind(&req); err != nil {
	//		return
	//	}
	//	//sess := sessions.Default(ctx)
	//	//sess.Get("uid")
	//	uc, ok := ctx.MustGet("user").(UserClaims)
	//	if !ok {
	//		//ctx.String(http.StatusOK, "系统错误")
	//		ctx.AbortWithStatus(http.StatusUnauthorized)
	//		return
	//	}
	//	// 用户输入不对
	//	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	//	if err != nil {
	//		//ctx.String(http.StatusOK, "系统错误")
	//		ctx.String(http.StatusOK, "生日格式不对")
	//		return
	//	}
	//	err = h.svc.UpdateNonSensitiveInfo(ctx, domain.User{
	//		Id:       uc.Uid,
	//		Nickname: req.Nickname,
	//		Birthday: birthday,
	//		AboutMe:  req.AboutMe,
	//	})
	//	if err != nil {
	//		ctx.String(http.StatusOK, "系统异常")
	//		return
	//	}
	//	ctx.String(http.StatusOK, "更新成功")
	//}

	//func (u *UserHandler) Profile(ctx *gin.Context) {
	//
	//	ctx.String(http.StatusOK, "Sign in successful")
	//
}

func (h *UserHandler) ProfileJWT(ctx *gin.Context) {
	//
	//	//us := ctx.MustGet("user").(UserClaims)
	//	//ctx.String(http.StatusOK, "这是 profile")
	//	// 嵌入一段刷新过期时间的代码
	//
	//	uc, ok := ctx.MustGet("user").(UserClaims)
	//	if !ok {
	//		//ctx.String(http.StatusOK, "系统错误")
	//		ctx.AbortWithStatus(http.StatusUnauthorized)
	//		return
	//	}
	//	u, err := h.svc.FindById(ctx, uc.Uid)
	//	if err != nil {
	//		ctx.String(http.StatusOK, "系统异常")
	//		return
	//	}
	//	type User struct {
	//		Nickname string `json:"nickname"`
	//		Email    string `json:"email"`
	//		AboutMe  string `json:"aboutMe"`
	//		Birthday string `json:"birthday"`
	//	}
	//	ctx.JSON(http.StatusOK, User{
	//		Nickname: u.Nickname,
	//		Email:    u.Email,
	//		AboutMe:  u.AboutMe,
	//		Birthday: u.Birthday.Format(time.DateOnly),
	//	})
	//
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {

	ug := server.Group("/users")

	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.LogInJWT)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.ProfileJWT)

}
