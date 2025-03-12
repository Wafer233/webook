package web

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"project/webook/internal/domain"
	"project/webook/internal/service"
	svcmock "project/webook/internal/service/mocks"
	"testing"
)

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.UserService
		reqBody  string
		wantCode int
		wantBody string
	}{
		{
			name: "sign up success",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmock.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@gmail.com",
					Password: "123hello123",
				}).Return(nil)
				return usersvc
			},
			reqBody: `{
				"email":"123@gmail.com",
				"password":"123hello123",
				"confirmPassword": "123hello123"
			}`,
			wantCode: http.StatusOK,
			wantBody: "Sign up successful",
		},
		{
			name: "bind fail",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmock.NewMockUserService(ctrl)
				//usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
				//	Email:    "123@gmail.com",
				//	Password: "123hello123",
				//}).Return(nil)
				return usersvc
			},
			reqBody: `{
				"email":"123@gmail.com",
				"password":"123hello123"
				`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "Email format error",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmock.NewMockUserService(ctrl)
				//usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
				//	Email:    "123@gmail.com",
				//	Password: "123hello123",
				//}).Return(nil)
				return usersvc
			},
			reqBody: `{
				"email":"123gmail.com",
				"password":"123hello123",
				"confirmPassword": "123hello123"
			}`,
			wantCode: http.StatusOK,
			wantBody: "Email format error",
		},
		{
			name: "password confirm error",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmock.NewMockUserService(ctrl)
				//usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
				//	Email:    "123@gmail.com",
				//	Password: "123hello123",
				//}).Return(nil)
				return usersvc
			},
			reqBody: `{
				"email":"123@gmail.com",
				"password":"123hello123",
				"confirmPassword": "123he123"
			}`,
			wantCode: http.StatusOK,
			wantBody: "Password error",
		},
		{
			name: "password format error",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmock.NewMockUserService(ctrl)
				//usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
				//	Email:    "123@gmail.com",
				//	Password: "123hello123",
				//}).Return(nil)
				return usersvc
			},
			reqBody: `{
				"email":"123@gmail.com",
				"password":"123123",
				"confirmPassword": "123123"
			}`,
			wantCode: http.StatusOK,
			wantBody: "At least 8 characters in length, containing at least one letter and one number",
		},
		{
			name: "email exist",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmock.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@gmail.com",
					Password: "123hello123",
				}).Return(service.ErrUserDuplicatedEmail)
				return usersvc
			},
			reqBody: `{
				"email":"123@gmail.com",
				"password":"123hello123",
				"confirmPassword": "123hello123"
			}`,
			wantCode: http.StatusOK,
			wantBody: "Email already exists",
		},
		{
			name: "other error",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmock.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@gmail.com",
					Password: "123hello123",
				}).Return(errors.New("other error"))
				return usersvc
			},
			reqBody: `{
				"email":"123@gmail.com",
				"password":"123hello123",
				"confirmPassword": "123hello123"
			}`,
			wantCode: http.StatusOK,
			wantBody: "System error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			h := NewUserHandler(tc.mock(ctrl))
			h.RegisterRoutes(server)
			req, err := http.NewRequest(http.MethodPost,
				"/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())

		})
	}
}

//func TestMock(t *testing.T) {
//	ctrl := gomock.NewController(t)
//
//	defer ctrl.Finish()
//
//	usersvc := svcmock.NewMockUserService(ctrl)
//
//	usersvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).
//		Return(errors.New("mock error"))
//
//	err := usersvc.SignUp(context.Background(), domain.User{
//		Email: "test@test.com",
//	})
//	t.Log(err)
//}
