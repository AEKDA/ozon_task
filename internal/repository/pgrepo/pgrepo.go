package pgrepo

import (
	"context"

	"github.com/AEKDA/ozon_task/internal/api/graph/model"
	"github.com/AEKDA/ozon_task/internal/repository/cursor"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AddPost(ctx context.Context, post model.AddPostInput) (*model.Post, error) {
	var newPost model.Post
	err := r.db.QueryRow(ctx,
		"INSERT INTO posts (title, content, author, allow_comments) VALUES ($1, $2, $3, $4) RETURNING id, title, content, created_at, author, allow_comments",
		post.Title, post.Content, post.Author, post.AllowComments).Scan(
		&newPost.ID, &newPost.Title, &newPost.Content, &newPost.CreatedAt, &newPost.Author, &newPost.AllowComments)
	if err != nil {
		return nil, err
	}
	return &newPost, nil
}

func (r *Repository) GetPostByID(ctx context.Context, id int64) (*model.Post, error) {
	var post model.Post
	err := r.db.QueryRow(ctx,
		"SELECT id, title, content, author, allow_comments,created_at FROM posts WHERE id=$1", id).Scan(
		&post.ID, &post.Title, &post.Content, &post.Author, &post.AllowComments, &post.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

func (r *Repository) GetPosts(ctx context.Context, first int, after *string) (*model.PostConnection, error) {
	var rows pgx.Rows
	var err error

	cursorID, err := cursor.Decode(after)
	if err != nil {
		return nil, err
	}

	if cursorID != nil {
		rows, err = r.db.Query(ctx,
			"SELECT id, title, content, author, allow_comments, created_at FROM posts WHERE id > $1 ORDER BY id ASC LIMIT $2",
			*cursorID, first)
	} else {
		rows, err = r.db.Query(ctx,
			"SELECT id, title, content, author, allow_comments, created_at FROM posts ORDER BY id ASC LIMIT $1",
			first)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var post model.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.AllowComments, &post.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	edges := make([]model.PostEdge, len(posts))
	for i, post := range posts {
		edges[i] = model.PostEdge{
			Cursor: cursor.Encode(post.ID),
			Node:   post,
		}
	}

	pageInfo := model.PageInfo{
		HasNextPage: len(posts) == first,
		StartCursor: edges[0].Cursor,
		EndCursor:   edges[len(edges)-1].Cursor,
	}

	return &model.PostConnection{
		Edges:    edges,
		PageInfo: pageInfo,
	}, nil
}

func (r *Repository) SetCommentPremission(ctx context.Context, postID int64, allow bool) (*model.Post, error) {
	var post model.Post
	err := r.db.QueryRow(ctx,
		"UPDATE posts SET allow_comments = $1 WHERE id = $2 RETURNING id, title, content, author, allow_comments, created_at",
		allow, postID).Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.AllowComments, &post.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *Repository) AddCommentToPost(ctx context.Context, commentInput model.AddCommentInput) (*model.Comment, error) {
	var newComment model.Comment
	err := r.db.QueryRow(ctx,
		"INSERT INTO comments (post_id, content, author) VALUES ($1, $2, $3) RETURNING id, post_id, content, author, created_at",
		commentInput.PostID, commentInput.Content, commentInput.Author).Scan(
		&newComment.ID, &newComment.Content, &newComment.Author, &newComment.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &newComment, nil
}

func (r *Repository) AddReplyToComment(ctx context.Context, commentInput model.AddReplyInput) (*model.Comment, error) {
	var newComment model.Comment
	err := r.db.QueryRow(ctx,
		"INSERT INTO comments (content, author, reply_to) VALUES ((SELECT post_id FROM comments WHERE id=$1), $2, $3, $1) RETURNING id, post_id, content, author, created_at, reply_to",
		commentInput.CommentID, commentInput.Content, commentInput.Author).Scan(
		&newComment.ID, &newComment.Content, &newComment.Author, &newComment.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &newComment, nil
}
