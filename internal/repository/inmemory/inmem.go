package inmemory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/AEKDA/ozon_task/internal/api/graph/model"
	"golang.org/x/exp/maps"
)

var (
	ErrNotFound error = errors.New("error not found")
)

type InMemoryDB struct {
	posts            map[int64]Post
	comments         map[int64]Comment
	mu               sync.RWMutex
	postIDCounter    int64
	commentIDCounter int64
}

func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{
		posts:    make(map[int64]Post),
		comments: make(map[int64]Comment),
	}
}

func (db *InMemoryDB) AddPost(ctx context.Context, postInput model.AddPostInput) (*model.Post, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	post := Post{
		ID:            db.generatePostID(),
		CreatedAt:     time.Now(),
		Title:         postInput.Title,
		Content:       postInput.Content,
		Author:        postInput.Author,
		AllowComments: postInput.AllowComments,
	}

	db.posts[post.ID] = post

	modelPost := post.toModel()
	return &modelPost, nil
}

func (db *InMemoryDB) AddCommentToPost(ctx context.Context, commentInput model.AddCommentInput) (*model.Comment, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	post, ok := db.posts[commentInput.PostID]
	if !ok || !post.AllowComments {
		return nil, errors.New("the post was not found or comments cannot be left under it")
	}

	comment := Comment{
		ID:        db.generateCommentID(),
		Content:   commentInput.Content,
		Author:    commentInput.Author,
		CreatedAt: time.Now(),
		PostID:    post.ID,
		ReplyTo:   nil,
	}

	db.comments[comment.ID] = comment

	modelComment := comment.toModel()
	return &modelComment, nil
}

func (db *InMemoryDB) SetCommentPremission(ctx context.Context, postID int64, allow bool) (*model.Post, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if post, ok := db.posts[postID]; ok {
		post.AllowComments = allow
		db.posts[postID] = post

		modelPost := post.toModel()
		return &modelPost, nil
	}

	return nil, ErrNotFound
}

func (db *InMemoryDB) AddReplyToComment(ctx context.Context, commentInput model.AddReplyInput) (*model.Comment, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	comment, ok := db.comments[commentInput.CommentID]
	if !ok {
		return nil, fmt.Errorf("comment not found")
	}

	post, ok := db.posts[comment.PostID]
	if !ok || !post.AllowComments {
		return nil, fmt.Errorf("the post was not found or comments cannot be left under it")
	}

	reply := Comment{
		ID:        db.generateCommentID(),
		ReplyTo:   &commentInput.CommentID,
		Content:   commentInput.Content,
		Author:    commentInput.Author,
		PostID:    comment.PostID,
		CreatedAt: time.Now(),
	}

	db.comments[reply.ID] = reply

	modelComment := reply.toModel()
	return &modelComment, nil
}

func (db *InMemoryDB) GetPosts(ctx context.Context, first int, after *string) (*model.PostConnection, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	connection, err := PostsToCursorPagination(maps.Values(db.posts), first, after)

	return &connection, err
}

func (db *InMemoryDB) GetCommentsByPostID(ctx context.Context, postID int64, first int, after *string) (model.CommentConnection, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	_, ok := db.posts[postID]
	if !ok {
		return model.CommentConnection{}, ErrNotFound
	}

	comments := []Comment{}
	for _, v := range db.comments {
		if v.PostID == postID {
			comments = append(comments, v)
		}
	}

	connection, err := commentsToCursorPagination(comments, first, after)

	return connection, err
}

func (db *InMemoryDB) GetPostByID(ctx context.Context, id int64) (*model.Post, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	post, ok := db.posts[id]
	if !ok {
		return nil, ErrNotFound
	}

	modelPost := post.toModel()
	return &modelPost, nil
}

func (db *InMemoryDB) generatePostID() int64 {
	db.postIDCounter++
	return db.postIDCounter
}

func (db *InMemoryDB) generateCommentID() int64 {
	db.commentIDCounter++
	return db.commentIDCounter
}
