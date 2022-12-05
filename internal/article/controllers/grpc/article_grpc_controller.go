package articleGrpc

import (
	"context"
	articleException "github.com/infranyx/go-grpc-template/internal/article/exception"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	articleDomain "github.com/infranyx/go-grpc-template/internal/article/domain"
	articleDto "github.com/infranyx/go-grpc-template/internal/article/dto"
	articleV1 "github.com/infranyx/protobuf-template-go/golang-grpc-template/article/v1"
)

type ArticleGrpcController struct {
	articleUC articleDomain.ArticleUseCase
}

func New(uc articleDomain.ArticleUseCase) *ArticleGrpcController {
	return &ArticleGrpcController{
		articleUC: uc,
	}
}

func (ac *ArticleGrpcController) CreateArticle(ctx context.Context, req *articleV1.CreateArticleRequest) (*articleV1.CreateArticleResponse, error) {
	aDto := &articleDto.CreateArticle{
		Name:        req.Name,
		Description: req.Desc,
	}
	err := aDto.ValidateCreateArticleDto()
	if err != nil {
		return nil, articleException.CreateArticleValidationExc(err)
	}
	article, err := ac.articleUC.Create(ctx, aDto)
	if err != nil {
		return nil, err
	}
	return &articleV1.CreateArticleResponse{
		Id:   article.ID.String(),
		Name: article.Name,
		Desc: article.Description,
	}, nil
}

func (ac *ArticleGrpcController) GetArticleById(ctx context.Context, req *articleV1.GetArticleByIdRequest) (*articleV1.GetArticleByIdResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
