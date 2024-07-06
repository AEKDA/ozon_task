package inmemory

import (
	"context"
	"errors"
	"fmt"
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
	if !ok {
		return nil, errors.New("post not found")
	}

	comment := model.Comment{
		ID:        db.generateCommentID(),
		Content:   commentInput.Content,
		Author:    commentInput.Author,
		CreatedAt: time.Now(),
		Replies:   model.CommentConnection{},
	}

	post.Comments.Edges = append(post.Comments.Edges, model.CommentEdge{
		Cursor: cursor.Encode(comment.ID),
		Node:   comment,
	})

	db.comments[comment.ID] = comment

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

	comment, ok := db.comments[commentInput.CommentID]
	if !ok {
		return nil, errors.New("comment not found")
	}

	newReply := model.Comment{
		ID:        db.generateCommentID(),
		Content:   commentInput.Content,
		Author:    commentInput.Author,
		CreatedAt: time.Now(),
		Replies:   model.CommentConnection{},
	}

	comment.Replies.Edges = append(comment.Replies.Edges, model.CommentEdge{
		Cursor: fmt.Sprintf("%d", newReply.ID),
		Node:   newReply,
	})

	db.comments[comment.ID] = comment

	return &newReply, nil
}

func (db *InMemoryDB) GetPosts(ctx context.Context, first int, after *string) (*model.PostConnection, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var startIdx int

	cursor, err := cursor.Decode(after)
	if err != nil {
		return nil, err
	}

	if cursor != nil {
		found := false
		for idx, post := range db.posts {
			if post.ID == *cursor {
				startIdx = int(idx) + 1
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("invalid after cursor")
		}
	}

	var endIdx int

	endIdx = startIdx + first
	if endIdx > len(db.posts) {
		endIdx = len(db.posts)
	}

	edges := make([]model.PostEdge, 0, endIdx-startIdx)
	for idx := startIdx; idx < endIdx; idx++ {
		post := db.posts[int64(idx)]
		edges = append(edges, model.PostEdge{
			Cursor: fmt.Sprintf("%d", post.ID),
			Node:   post,
		})
	}

	pageInfo := model.PageInfo{
		HasNextPage: endIdx < len(db.posts),
		StartCursor: edges[0].Cursor,
		EndCursor:   edges[len(edges)-1].Cursor,
	}

	return &model.PostConnection{
		Edges:    edges,
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

// Helper methods
func (db *InMemoryDB) generatePostID() int64 {
	db.postIDCounter++
	return db.postIDCounter
}

func (db *InMemoryDB) generateCommentID() int64 {
	db.commentIDCounter++
	return db.commentIDCounter
}
