package service

import (
	"context"
	"sync"

	"github.com/AEKDA/ozon_task/internal/api/graph/model"
)

type PostRepository interface {
	AddPost(ctx context.Context, post model.AddPostInput) (*model.Post, error)
	GetPostByID(ctx context.Context, id int64) (*model.Post, error)
	GetPosts(ctx context.Context, first int, after *string) (*model.PostConnection, error)
	SetCommentPremission(ctx context.Context, postID int64, allow bool) (*model.Post, error)
}

type CommentRepository interface {
	AddCommentToPost(ctx context.Context, commentInput model.AddCommentInput) (*model.Comment, error)
	AddReplyToComment(ctx context.Context, commentInput model.AddReplyInput) (*model.Comment, error)
	GetCommentsByPostID(ctx context.Context, postID int64, first int, after *string) (model.CommentConnection, error)
}

type PostService struct {
	postRepo    PostRepository
	commentRepo CommentRepository

	mu          sync.Mutex
	subscribers map[int64][]chan *model.Comment
}

func NewPostService(post PostRepository, comment CommentRepository) *PostService {
	return &PostService{
		postRepo:    post,
		commentRepo: comment,

		subscribers: make(map[int64][]chan *model.Comment),
	}
}

func (s *PostService) SubscriptionOnPost(ctx context.Context, postID int64) (<-chan *model.Comment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	commentChan := make(chan *model.Comment, 1)
	s.subscribers[postID] = append(s.subscribers[postID], commentChan)

	return commentChan, nil
}

func (r *PostService) NotifySubscribers(postID int64, comment *model.Comment) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if chans, found := r.subscribers[postID]; found {
		for _, ch := range chans {

			ch <- comment
		}
	}
}

func (s *PostService) AddPost(ctx context.Context, input model.AddPostInput) (*model.Post, error) {
	return s.postRepo.AddPost(ctx, input)
}

func (s *PostService) AddCommentToPost(ctx context.Context, input model.AddCommentInput) (*model.Comment, error) {

	comment, err := s.commentRepo.AddCommentToPost(ctx, input)

	s.NotifySubscribers(input.PostID, comment)

	return comment, err
}

func (s *PostService) AddReplyToComment(ctx context.Context, input model.AddReplyInput) (*model.Comment, error) {
	return s.commentRepo.AddReplyToComment(ctx, input)
}

func (s *PostService) SetCommentPremission(ctx context.Context, postID int64, allow bool) (*model.Post, error) {
	return s.postRepo.SetCommentPremission(ctx, postID, allow)
}

func (s *PostService) Posts(ctx context.Context, first int, after *string) (*model.PostConnection, error) {
	return s.postRepo.GetPosts(ctx, first, after)
}

func (s *PostService) Comments(ctx context.Context, postID int64, first int, after *string) (*model.CommentConnection, error) {
	comments, err := s.commentRepo.GetCommentsByPostID(ctx, postID, first, after)
	if err != nil {
		return nil, err
	}

	return &comments, nil
}

func (s *PostService) Post(ctx context.Context, id int64) (*model.Post, error) {
	return s.postRepo.GetPostByID(ctx, id)
}
