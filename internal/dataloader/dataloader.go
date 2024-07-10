package dataloader

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AEKDA/ozon_task/internal/api/graph/model"
	"github.com/AEKDA/ozon_task/internal/service"
	"github.com/vikstrous/dataloadgen"
)

type ctxKey string

const (
	loadersKey = ctxKey("dataloaders")
	paramKey   = ctxKey("param")
)

type commentReader struct {
	db service.CommentRepository
}

func multiplyError(err error, n int) []error {
	errs := make([]error, n)
	for i := range errs {
		errs[i] = err
	}
	return errs
}

func (u commentReader) getComments(ctx context.Context, postIDs []int64) ([]model.CommentConnection, []error) {
	param, ok := ctx.Value(paramKey).(commentParam)
	if !ok {
		return nil, []error{fmt.Errorf("error on commentParam")}
	}
	comments, err := u.db.GetCommentsByPostIDs(ctx, postIDs, param.Limit, param.After)

	if err != nil {
		return make([]model.CommentConnection, len(postIDs)), multiplyError(
			fmt.Errorf("repo error %w", err),
			len(postIDs),
		)
	}

	res := make([]model.CommentConnection, len(postIDs))
	for i, v := range postIDs {
		res[i] = comments[v]
	}

	return res, nil
}

type Loaders struct {
	CommentLoader *dataloadgen.Loader[int64, model.CommentConnection]
}

func NewLoaders(repo service.CommentRepository) *Loaders {

	ur := &commentReader{db: repo}
	return &Loaders{
		CommentLoader: dataloadgen.NewLoader(ur.getComments, dataloadgen.WithWait(time.Millisecond)),
	}
}

func Middleware(conn service.CommentRepository, next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loader := NewLoaders(conn)
		r = r.WithContext(context.WithValue(r.Context(), loadersKey, loader))
		next.ServeHTTP(w, r)
	})
}

func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}

type commentParam struct {
	Limit int
	After *string
}

func GetComments(ctx context.Context, postID int64, limit int, after *string) (model.CommentConnection, error) {

	loaders := For(ctx)
	queryParam := commentParam{
		Limit: limit,
		After: after,
	}

	comments, err := loaders.CommentLoader.Load(context.WithValue(ctx, paramKey, queryParam), postID)
	if err != nil {
		return model.CommentConnection{}, fmt.Errorf("load from context loader %w", err)
	}

	return comments, nil
}
