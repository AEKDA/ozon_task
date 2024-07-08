package inmemory

import (
	"slices"
	"time"

	"github.com/AEKDA/ozon_task/internal/api/graph/model"
	"github.com/AEKDA/ozon_task/internal/repository/cursor"
)

type Post struct {
	ID            int64
	Title         string
	Content       string
	Author        string
	CreatedAt     time.Time
	AllowComments bool
}

type Comment struct {
	ID        int64
	PostID    int64
	Content   string
	Author    string
	CreatedAt time.Time
	ReplyTo   *int64
}

func commentsToCursorPagination(comments []Comment, first int, after *string) (model.CommentConnection, error) {

	var filteredComments []model.CommentEdge

	afterID, err := cursor.Decode(after)
	if err != nil {
		return model.CommentConnection{}, err
	}
	for _, comment := range comments {
		if afterID != nil && comment.ID <= *afterID {
			continue
		}
		filteredComments = append(filteredComments, model.CommentEdge{
			Node:   comment.toModel(),
			Cursor: cursor.Encode(comment.ID),
		})
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

func PostsToCursorPagination(posts []Post, first int, after *string) (model.PostConnection, error) {

	var filteredPosts []model.PostEdge

	afterID, err := cursor.Decode(after)
	if err != nil {
		return model.PostConnection{}, err
	}

	for _, post := range posts {
		if afterID != nil && post.ID <= *afterID {
			continue
		}
		filteredPosts = append(filteredPosts, model.PostEdge{
			Node:   post.toModel(),
			Cursor: cursor.Encode(post.ID),
		})
	}

	slices.SortFunc(filteredPosts, func(a model.PostEdge, b model.PostEdge) int { return int(a.Node.ID - b.Node.ID) })

	pageInfo := model.PageInfo{
		HasNextPage: len(filteredPosts) > first,
	}
	if len(filteredPosts) > first {
		filteredPosts = filteredPosts[:first]
	}

	if len(filteredPosts) > 0 {
		pageInfo.StartCursor = filteredPosts[0].Cursor
		pageInfo.EndCursor = filteredPosts[len(filteredPosts)-1].Cursor
	}

	return model.PostConnection{
		Edges:    filteredPosts,
		PageInfo: pageInfo,
	}, nil
}

func (c Comment) toModel() model.Comment {
	return model.Comment{
		ID:        c.ID,
		Content:   c.Content,
		Author:    c.Author,
		CreatedAt: c.CreatedAt,
		ReplyTo:   c.ReplyTo,
	}
}

func (c Post) toModel() model.Post {
	return model.Post{
		ID:            c.ID,
		Content:       c.Content,
		Author:        c.Author,
		CreatedAt:     c.CreatedAt,
		Title:         c.Title,
		AllowComments: c.AllowComments,
		Comments:      model.CommentConnection{},
	}
}
