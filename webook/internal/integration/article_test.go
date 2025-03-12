package integration

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"project/webook/internal/domain"
	"project/webook/internal/integration/startup"
	"project/webook/internal/repository/dao"
	"project/webook/internal/web"
	"testing"
)

type ArticleTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func (suite *ArticleTestSuite) SetupSuite() {
	//suite.route = startup.InitWebServer()
	suite.db = startup.InitDB()
	suite.server = gin.Default()
	suite.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &web.UserClaims{
			Uid: 233,
		})
	})
	artHdl := startup.InitArticleHandler()
	artHdl.RegisterRoutes(suite.server)
}

//func (suite *ArticleTestSuite) TestFirst() {
//	suite.T().Log("hello, this is the first test")
//}

func (suite *ArticleTestSuite) TestArticle_Publish() {
	t := suite.T()

	testCases := []struct {
		name string
		// 要提前准备数据
		before func(t *testing.T)
		// 验证并且删除数据
		after func(t *testing.T)
		req   Article

		// 预期响应
		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子并发表",
			before: func(t *testing.T) {
				// 什么也不需要做
			},
			after: func(t *testing.T) {
				// 验证一下数据
				var art dao.Article
				err := suite.db.Where("author_id = ?", 233).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Id > 0)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Utime = 0
				art.Ctime = 0
				art.Id = 0
				assert.Equal(t, dao.Article{
					Title:    "my title",
					Content:  "my content",
					Status:   domain.ArticleStatusPublished.ToUint8(),
					AuthorId: 233,
				}, art)

				var publishedArt dao.PublishedArticle
				err = suite.db.Where("author_id = ?", 233).First(&publishedArt).Error
				assert.NoError(t, err)
				assert.True(t, publishedArt.Id > 0)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
				publishedArt.Utime = 0
				publishedArt.Ctime = 0
				publishedArt.Id = 0
				assert.Equal(t, dao.PublishedArticle{
					Article: dao.Article{
						Title:    "my title",
						Content:  "my content",
						Status:   domain.ArticleStatusPublished.ToUint8(),
						AuthorId: 233,
					},
				}, publishedArt)
			},
			req: Article{
				Title:   "my title",
				Content: "my content",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 1,
			},
		},
		{
			// 制作库有，但是线上库没有
			name: "更新帖子并新发表",
			before: func(t *testing.T) {
				// 模拟已经存在的帖子
				err := suite.db.Create(&dao.Article{
					Id:       2,
					Title:    "my title",
					Content:  "my content",
					Ctime:    10000,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
					Utime:    20000,
					AuthorId: 233,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证一下数据
				var art dao.Article
				err := suite.db.Where("id = ?", 2).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 0)
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "new title",
					Content:  "new content",
					Status:   domain.ArticleStatusPublished.ToUint8(),
					AuthorId: 233,
					Ctime:    10000,
				}, art)
				var publishedArt dao.PublishedArticle
				err = suite.db.Where("id = ?", 2).First(&publishedArt).Error
				assert.NoError(t, err)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
				publishedArt.Utime = 0
				publishedArt.Ctime = 0
				assert.Equal(t, dao.PublishedArticle{
					Article: dao.Article{
						Id:       2,
						Title:    "new title",
						Content:  "new content",
						Status:   domain.ArticleStatusPublished.ToUint8(),
						AuthorId: 233,
					},
				}, publishedArt)
			},
			req: Article{
				Id:      2,
				Title:   "new title",
				Content: "new content",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 2,
			},
		},
		{
			name: "更新帖子，并且重新发表",
			before: func(t *testing.T) {
				art := dao.Article{
					Id:       3,
					Title:    "my title",
					Content:  "my content",
					Ctime:    10000,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Utime:    20000,
					AuthorId: 233,
				}
				suite.db.Create(&art)
				part := dao.PublishedArticle{Article: art}
				suite.db.Create(&part)
			},
			after: func(t *testing.T) {
				var art dao.Article
				err := suite.db.Where("id = ?", 3).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 20000)
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       3,
					Title:    "new title",
					Content:  "new content",
					Ctime:    10000,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					AuthorId: 233,
				}, art)

				var part dao.PublishedArticle
				err = suite.db.Where("id = ?", 3).First(&part).Error
				assert.NoError(t, err)
				assert.True(t, part.Utime > 20000)
				part.Utime = 0
				assert.Equal(t, dao.PublishedArticle{
					Article: dao.Article{
						Id:       3,
						Title:    "new title",
						Content:  "new content",
						Ctime:    10000,
						Status:   domain.ArticleStatusPublished.ToUint8(),
						AuthorId: 233,
					},
				}, part)

			},
			req: Article{
				Id:      3,
				Title:   "new title",
				Content: "new content",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 3,
			},
		},
		{
			name: "更新别人的帖子，并且发表失败",
			before: func(t *testing.T) {
				art := dao.Article{
					Id:      4,
					Title:   "my title",
					Content: "my content",
					Ctime:   10000,
					Utime:   20000,
					Status:  domain.ArticleStatusPublished.ToUint8(),
					// 注意。这个 AuthorID 我们设置为另外一个人的ID
					AuthorId: 100,
				}
				suite.db.Create(&art)
				part := dao.PublishedArticle{
					Article: dao.Article{
						Id:      4,
						Title:   "my title",
						Content: "my content",
						Ctime:   10000,
						Utime:   20000,
						Status:  domain.ArticleStatusPublished.ToUint8(),
						// 注意。这个 AuthorID 我们设置为另外一个人的ID
						AuthorId: 100,
					},
				}
				suite.db.Create(&part)
			},
			after: func(t *testing.T) {
				// 更新应该是失败了，数据没有发生变化
				var art dao.Article
				err := suite.db.Where("id = ?", 4).First(&art).Error
				assert.NoError(t, err)
				assert.Equal(t, dao.Article{
					Id:       4,
					Title:    "my title",
					Content:  "my content",
					Ctime:    10000,
					Utime:    20000,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					AuthorId: 100,
				}, art)

				var part dao.PublishedArticle
				// 数据没有变化
				err = suite.db.Where("id = ?", 4).First(&part).Error
				assert.NoError(t, err)
				assert.Equal(t, dao.PublishedArticle{
					Article: dao.Article{
						Id:       4,
						Title:    "my title",
						Content:  "my content",
						Ctime:    10000,
						Utime:    20000,
						Status:   domain.ArticleStatusPublished.ToUint8(),
						AuthorId: 100,
					},
				}, part)
			},
			req: Article{
				Id:      4,
				Title:   "new title",
				Content: "new content",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Code: 5,
				Msg:  "system error",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			data, err := json.Marshal(tc.req)
			// 不能有 error
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/publish", bytes.NewReader(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type",
				"application/json")
			recorder := httptest.NewRecorder()

			suite.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}
			// 反序列化为结果
			// 利用泛型来限定结果必须是 int64
			var result Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult, result)
			tc.after(t)
		})
	}
}

func (suite *ArticleTestSuite) TestEdit() {
	t := suite.T()
	testCases := []struct {
		name string
		//prepare data
		before func(t *testing.T)
		after  func(t *testing.T)

		//预期输入
		art Article

		//http response with code
		wantCode int
		//response with article id
		wantRes Result[int64]
	}{
		{
			name: "新建帖子",
			before: func(t *testing.T) {
				//suite.b
			},
			after: func(t *testing.T) {
				var art dao.Article
				err := suite.db.Where("id=?", 1).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       1,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 233,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}, art)
			},
			art: Article{
				Title:   "我的标题",
				Content: "我的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 1,
			},
		},
		{
			name: "修改帖子",
			before: func(t *testing.T) {
				err := suite.db.Create(dao.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 233,
					Ctime:    100000,
					Utime:    200000,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}).Error
				assert.NoError(t, err)
			},

			after: func(t *testing.T) {
				var art dao.Article
				err := suite.db.Where("id=?", 2).First(&art).Error
				assert.NoError(t, err)
				//assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 200000)
				//art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorId: 233,
					Ctime:    100000,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}, art)
			},
			art: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 2,
			},
		},
		{
			name: "他人篡改帖子",
			before: func(t *testing.T) {
				err := suite.db.Create(dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 233000,
					Ctime:    100000,
					Utime:    200000,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}).Error
				assert.NoError(t, err)
			},

			after: func(t *testing.T) {
				var art dao.Article
				err := suite.db.Where("id=?", 3).First(&art).Error
				assert.NoError(t, err)
				//assert.True(t, art.Ctime > 0)
				//assert.True(t, art.Utime > 200000)
				//art.Ctime = 0
				//art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 233000,
					Ctime:    100000,
					Utime:    200000,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, art)
			},
			art: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//请求
			tc.before(t)
			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/edit", bytes.NewBuffer([]byte(reqBody)))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()

			suite.server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			var result Result[int64]
			err = json.Unmarshal(resp.Body.Bytes(), &result)

			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, result)
			tc.after(t)

		})
	}
}

func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleTestSuite{})
}

func (suite *ArticleTestSuite) TearDownTest() {
	suite.db.Exec("TRUNCATE TABLE articles")
	suite.db.Exec("TRUNCATE TABLE published_articles")
}

type Article struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Id      int64  `json:"id"`
}

type Result[T any] struct {
	// 这个叫做业务错误码
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
