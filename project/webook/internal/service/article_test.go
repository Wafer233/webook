package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"project/webook/internal/domain"
	"project/webook/internal/repository"
	repov1mocks "project/webook/internal/repository/mocks"
	"project/webook/pkg/logger"
	"testing"
)

func Test_articleService_Publish(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository,
			repository.ArticleReaderRepository)
		art domain.Article

		wantErr error
		wantId  int64
	}{
		{
			name: "publish success",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				readRepo := repov1mocks.NewMockArticleReaderRepository(ctrl)
				authRepo := repov1mocks.NewMockArticleAuthorRepository(ctrl)

				authRepo.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 233,
					},
				}).Return(int64(1), nil)

				readRepo.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 233,
					},
				}).Return(int64(1), nil)
				return authRepo, readRepo
			},

			art: domain.Article{
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 233,
				},
			},
			wantId: 1,
		},
		{
			name: "publish success with an edit",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				readRepo := repov1mocks.NewMockArticleReaderRepository(ctrl)
				authRepo := repov1mocks.NewMockArticleAuthorRepository(ctrl)

				authRepo.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 233,
					},
				}).Return(nil)

				readRepo.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 233,
					},
				}).Return(int64(2), nil)
				return authRepo, readRepo
			},

			art: domain.Article{
				Id:      2,
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 233,
				},
			},
			wantId: 2,
		},
		{
			name: "fail to save into author repo",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				readRepo := repov1mocks.NewMockArticleReaderRepository(ctrl)
				authRepo := repov1mocks.NewMockArticleAuthorRepository(ctrl)

				authRepo.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 233,
					},
				}).Return(errors.New("mock db error"))

				//readRepo.EXPECT().Save(gomock.Any(), domain.Article{
				//	Id:      2,
				//	Title:   "my title",
				//	Content: "my content",
				//	Author: domain.Author{
				//		Id: 233,
				//	},
				//}).Return(int64(0), errors.New("mock db error"))
				return authRepo, readRepo
			},

			art: domain.Article{
				Id:      2,
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 233,
				},
			},
			wantId:  0,
			wantErr: errors.New("mock db error"),
		},
		{
			name: "fail to save into reader repo, but remake success",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				readRepo := repov1mocks.NewMockArticleReaderRepository(ctrl)
				authRepo := repov1mocks.NewMockArticleAuthorRepository(ctrl)

				authRepo.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 233,
					},
				}).Return(nil)

				readRepo.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 233,
					},
				}).Return(int64(0), errors.New("mock db error"))
				//
				readRepo.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 233,
					},
				}).Return(int64(2), nil)
				return authRepo, readRepo
			},

			art: domain.Article{
				Id:      2,
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 233,
				},
			},
			wantId:  2,
			wantErr: nil,
		},
		{
			name: "ALL fail to save into reader repo",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				readRepo := repov1mocks.NewMockArticleReaderRepository(ctrl)
				authRepo := repov1mocks.NewMockArticleAuthorRepository(ctrl)

				authRepo.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 233,
					},
				}).Return(nil)

				readRepo.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 233,
					},
				}).Times(3).Return(int64(0), errors.New("mock db error"))

				return authRepo, readRepo
			},

			art: domain.Article{
				Id:      2,
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 233,
				},
			},
			wantId:  0,
			wantErr: errors.New("mock db error"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			authRepo, readRepo := testCase.mock(ctrl)
			svc := NewArticleServiceV1(authRepo, readRepo, &logger.NopLogger{})
			id, err := svc.PublishV1(context.Background(), testCase.art)
			assert.Equal(t, testCase.wantId, id)
			assert.Equal(t, testCase.wantErr, err)
		})
	}
}
