package inmemory

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/AEKDA/ozon_task/internal/api/graph/model"
	"github.com/AEKDA/ozon_task/internal/repository/cursor"
)

var (
	ErrNotFound error = errors.New("error not found")
)

type InMemoryDB struct {
	posts            map[int64]model.Post
	comments         map[int64]model.Comment
	mu               sync.RWMutex
	postIDCounter    int64
	commentIDCounter int64
}

func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{
		posts:    make(map[int64]model.Post),
		comments: make(map[int64]model.Comment),
	}
}

func (db *InMemoryDB) AddPost(ctx context.Context, postInput model.AddPostInput) (*model.Post, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	post := model.Post{
		ID:            db.generatePostID(),
		CreatedAt:     time.Now(),
		Title:         postInput.Title,
		Content:       postInput.Content,
		Author:        postInput.Author,
		AllowComments: postInput.AllowComments,
		Comments:      model.CommentConnection{},
	}

	db.posts[post.ID] = post
	return &post, nil
}

func (db *InMemoryDB) AddCommentToPost(ctx context.Context, commentInput model.AddCommentInput) (*model.Comment, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	post, ok := db.posts[commentInput.PostID]
	if !ok || !post.AllowComments {
		return nil, errors.New("the post was not found or comments cannot be left under it")
	}

	comment := model.Comment{
		ID:        db.generateCommentID(),
		Content:   commentInput.Content,
		Author:    commentInput.Author,
		CreatedAt: time.Now(),
		ReplyTo:   nil,
	}

	post.Comments.Edges = append(post.Comments.Edges, model.CommentEdge{
		Cursor: cursor.Encode(comment.ID),
		Node:   comment,
	})

	db.comments[comment.ID] = comment
	db.posts[post.ID] = post

	return &comment, nil
}
func (db *InMemoryDB) SetCommentPremission(ctx context.Context, postID int64, allow bool) (*model.Post, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if post, ok := db.posts[postID]; ok {
		post.AllowComments = allow
		db.posts[postID] = post
		return &post, nil
	}

	return nil, ErrNotFound
}

func (db *InMemoryDB) AddReplyToComment(ctx context.Context, commentInput model.AddReplyInput) (*model.Comment, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	post, ok := db.posts[commentInput.PostID]
	if !ok || !post.AllowComments {
		return nil, fmt.Errorf("the post was not found or comments cannot be left under it")
	}

	comment, ok := db.comments[commentInput.CommentID]
	if !ok {
		return nil, fmt.Errorf("comment not found")
	}

	newReply := model.Comment{
		ID:        db.generateCommentID(),
		Content:   commentInput.Content,
		Author:    commentInput.Author,
		CreatedAt: time.Now(),
		ReplyTo:   &commentInput.CommentID,
	}

	post.Comments.Edges = append(post.Comments.Edges, model.CommentEdge{Cursor: cursor.Encode(comment.ID), Node: comment})

	db.comments[comment.ID] = comment
	db.posts[post.ID] = post

	return &newReply, nil
}

func (db *InMemoryDB) GetPosts(ctx context.Context, first int, afterCursor *string) (*model.PostConnection, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var startIdx int

	after, err := cursor.Decode(afterCursor)
	if err != nil {
		return nil, err
	}

	if after != nil {
		found := false
		for idx, post := range db.posts {
			if post.ID == *after {
				startIdx = int(idx) + 1
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("invalid after cursor")
		}
	}

	endIdx := startIdx + first
	if endIdx > len(db.posts) {
		endIdx = len(db.posts)
	}

	edges := make([]model.PostEdge, 0, first)
	for idx := startIdx; idx <= endIdx; idx++ {
		post, ok := db.posts[int64(idx)]
		if !ok {
			continue
		}
		edges = append(edges, model.PostEdge{
			Cursor: cursor.Encode(post.ID),
			Node:   post,
		})
	}

	pageInfo := model.PageInfo{
		HasNextPage: endIdx < len(db.posts),
	}

	if len(edges) > 0 {
		pageInfo.StartCursor = edges[0].Cursor
		pageInfo.EndCursor = edges[len(edges)-1].Cursor
	}

	return &model.PostConnection{
		Edges:    edges,
		PageInfo: pageInfo,
	}, nil
}

func (db *InMemoryDB) GetCommentsByPostID(ctx context.Context, postID int64, first int, after *string) (model.CommentConnection, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	post, ok := db.posts[postID]
	if !ok {
		return model.CommentConnection{}, ErrNotFound
	}
	var filteredComments []model.CommentEdge
	if after != nil {
		afterID, err := cursor.Decode(after)
		if err != nil {
			return model.CommentConnection{}, err
		}
		for _, comment := range post.Comments.Edges {
			if comment.Node.ID > *afterID {
				filteredComments = append(filteredComments, comment)
			}
		}
	} else {
		filteredComments = post.Comments.Edges
	}
	slices.SortFunc(filteredComments, func(a model.CommentEdge, b model.CommentEdge) int { return int(a.Node.ID - b.Node.ID) })

	pageInfo := model.PageInfo{
		HasNextPage: len(filteredComments) > first,
	}
	if len(filteredComments) > first {
		filteredComments = filteredComments[:first]
	}

	if len(filteredComments) > 0 {
		pageInfo.StartCursor = filteredComments[0].Cursor
		pageInfo.EndCursor = filteredComments[len(filteredComments)-1].Cursor
	}

	return model.CommentConnection{
		Edges:    filteredComments,
		PageInfo: pageInfo,
	}, nil

}

func (db *InMemoryDB) GetPostByID(ctx context.Context, id int64) (*model.Post, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	post, ok := db.posts[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &post, nil
}

func (db *InMemoryDB) GetReplyByCommentID(ctx context.Context, commentID int64, first int, after *string) (*model.CommentConnection, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var comments []model.Comment

	for _, comment := range db.comments {
		if comment.ReplyTo != nil && *comment.ReplyTo == commentID {
			comments = append(comments, comment)
		}
	}

	if after != nil {
		afterID, err := cursor.Decode(after)
		if err != nil {
			return nil, err
		}
		var filteredComments []model.Comment
		for _, comment := range comments {
			if comment.ID > *afterID {
				filteredComments = append(filteredComments, comment)
			}
		}
		comments = filteredComments
	}

	if len(comments) > first {
		comments = comments[:first]
	}

	edges := make([]model.CommentEdge, len(comments))
	for i, comment := range comments {
		edges[i] = model.CommentEdge{
			Cursor: cursor.Encode(comment.ID),
			Node:   comment,
		}
	}

	pageInfo := model.PageInfo{
		HasNextPage: len(comments) > first,
	}
	if len(edges) > 0 {
		pageInfo.StartCursor = edges[0].Cursor
		pageInfo.EndCursor = edges[len(edges)-1].Cursor
	}

	return &model.CommentConnection{
		Edges:    edges,
		PageInfo: pageInfo,
	}, nil
}

func (db *InMemoryDB) generatePostID() int64 {
	db.postIDCounter++
	return db.postIDCounter
}

func (db *InMemoryDB) generateCommentID() int64 {
	db.commentIDCounter++
	return db.commentIDCounter
}
