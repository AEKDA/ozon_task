package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.49

import (
	"context"

	"github.com/AEKDA/ozon_task/internal/api/graph/model"
)

// AddPost is the resolver for the addPost field.
func (r *mutationResolver) AddPost(ctx context.Context, input model.AddPostInput) (*model.Post, error) {
	return r.PostService.AddPost(ctx, input)
}

// AddCommentToPost is the resolver for the addCommentToPost field.
func (r *mutationResolver) AddCommentToPost(ctx context.Context, input model.AddCommentInput) (*model.Comment, error) {
	return r.PostService.AddCommentToPost(ctx, input)
}

// AddReplyToComment is the resolver for the addReplyToComment field.
func (r *mutationResolver) AddReplyToComment(ctx context.Context, input model.AddReplyInput) (*model.Comment, error) {
	return r.PostService.AddReplyToComment(ctx, input)
}

// SetCommentPremission is the resolver for the setCommentPremission field.
func (r *mutationResolver) SetCommentPremission(ctx context.Context, postID int64, allow bool) (*model.Post, error) {
	return r.PostService.SetCommentPremission(ctx, postID, allow)
}

// Posts is the resolver for the posts field.
func (r *queryResolver) Posts(ctx context.Context, first int, after *string) (*model.PostConnection, error) {
	return r.PostService.Posts(ctx, first, after)
}

// Post is the resolver for the post field.
func (r *queryResolver) Post(ctx context.Context, id int64) (*model.Post, error) {
	return r.PostService.Post(ctx, id)
}

// CommentAdded is the resolver for the commentAdded field.
func (r *subscriptionResolver) CommentAdded(ctx context.Context, postID int64) (<-chan *model.Comment, error) {
	return r.PostService.SubscriptionOnPost(ctx, postID)
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Subscription returns SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
