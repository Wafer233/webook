package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"project/webook/internal/domain"
	"project/webook/internal/service"
	svcmock "project/webook/internal/service/mocks"
	"project/webook/pkg/logger"
	"testing"
)

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.ArticleService
		reqBody  string
		wantCode int
		wantRes  Result
	}{
		{
			name: "新建并发表",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmock.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 233,
					},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody: `{
						"title":"my title",
						"content":"my content",
						}`,
			wantCode: http.StatusOK,
			wantRes: Result{
				Data: 1,
			},
		},
		{
			name: "publish fail",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmock.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 233,
					},
				}).Return(int64(0), errors.New("publish fail"))
				return svc
			},
			reqBody: `{
						"title":"my title",
						"content":"my content",
						}`,
			wantCode: http.StatusOK,
			wantRes: Result{
				//Data: 1,
				Code: 5,
				Msg:  "system error",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			server.Use(func(ctx *gin.Context) {
				ctx.Set("claims", &UserClaims{
					Uid: 233,
				})
				h := NewArticleHandler(tc.mock(ctrl), &logger.NopLogger{})
				h.RegisterRoutes(server)
				req, err := http.NewRequest(http.MethodPost,
					"/articles/publish", bytes.NewBuffer([]byte(tc.reqBody)))
				require.NoError(t, err)

				req.Header.Set("Content-Type", "application/json")

				resp := httptest.NewRecorder()

				server.ServeHTTP(resp, req)
				assert.Equal(t, tc.wantCode, resp.Code)
				assert.Equal(t, tc.wantRes, resp.Body.String())

				assert.Equal(t, tc.wantCode, resp.Code)
				if resp.Code != http.StatusOK {
					return
				}
				var result Result
				err = json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, tc.wantRes, result)

			})
		})
	}
}
