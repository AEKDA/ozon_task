package pgrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AEKDA/ozon_task/internal/api/graph/model"
	"github.com/AEKDA/ozon_task/internal/repository/cursor"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Repository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func New(db *pgxpool.Pool, log *zap.Logger) *Repository {
	return &Repository{db: db, logger: log}
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
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &post, nil
}

func (r *Repository) GetCommentsByPostID(ctx context.Context, postID int64, first int, after *string) (model.CommentConnection, error) {
	var comments []model.Comment
	var args []interface{}

	args = append(args, postID)
	query := fmt.Sprintf("SELECT id, content, author, created_at, reply_to FROM comments WHERE post_id=$%d ", len(args))

	if after != nil {
		startID, err := cursor.Decode(after)
		if err != nil {
			return model.CommentConnection{}, fmt.Errorf("invalid cursor: %v", err)
		}
		args = append(args, *startID)
		query = fmt.Sprintf("%s AND id > $%d", query, len(args))
	}

	args = append(args, first+1)
	query = fmt.Sprintf("%s ORDER BY id ASC LIMIT $%d", query, len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.CommentConnection{}, nil
		}
		return model.CommentConnection{}, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var comment model.Comment
		if err := rows.Scan(&comment.ID, &comment.Content, &comment.Author, &comment.CreatedAt, &comment.ReplyTo); err != nil {
			return model.CommentConnection{}, fmt.Errorf("row scan failed: %v", err)
		}
		comments = append(comments, comment)
	}

	if rows.Err() != nil {
		return model.CommentConnection{}, fmt.Errorf("rows error: %v", rows.Err())
	}

	edges := make([]model.CommentEdge, 0, len(comments))
	for _, comment := range comments {
		edges = append(edges, model.CommentEdge{
			Cursor: cursor.Encode(comment.ID),
			Node:   comment,
		})
	}

	var pageInfo model.PageInfo
	if len(edges) > first {
		pageInfo = model.PageInfo{
			HasNextPage: true,
			StartCursor: edges[0].Cursor,
			EndCursor:   edges[first-1].Cursor,
		}
		edges = edges[:first] // Return only the first 'first' comments
	} else if len(edges) > 0 {
		pageInfo = model.PageInfo{
			HasNextPage: false,
			StartCursor: edges[0].Cursor,
			EndCursor:   edges[len(edges)-1].Cursor,
		}
	}

	return model.CommentConnection{
		Edges:    edges,
		PageInfo: pageInfo,
	}, nil
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
		"INSERT INTO comments (post_id, content, author) VALUES ($1, $2, $3) RETURNING id, content, author, created_at",
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
		"INSERT INTO comments (content, author, reply_to) VALUES ((SELECT post_id FROM comments WHERE id=$1), $2, $3, $1) RETURNING id, content, author, created_at",
		commentInput.CommentID, commentInput.Content, commentInput.Author).Scan(
		&newComment.ID, &newComment.Content, &newComment.Author, &newComment.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &newComment, nil
}

func (r *Repository) GetReplyByCommentID(ctx context.Context, commentID int64, first int, after *string) (*model.CommentConnection, error) {
	query := `
        SELECT id, author, content, created_at, reply_to
        FROM comments
        WHERE reply_to = $1 %s
        ORDER BY created_at ASC
        LIMIT $2
    `

	var rows pgx.Rows
	var err error

	if after != nil {
		var afterID *int64
		afterID, err = cursor.Decode(after)
		if err != nil {
			return nil, err
		}
		query = fmt.Sprintf(query, "AND id > $3")
		rows, err = r.db.Query(ctx, query, commentID, first, afterID)
	} else {
		query = fmt.Sprintf(query, "")
		rows, err = r.db.Query(ctx, query, commentID, first)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []model.Comment
	for rows.Next() {
		var comment model.Comment
		var replyTo sql.NullInt64

		if err := rows.Scan(&comment.ID, &comment.Author, &comment.Content, &comment.CreatedAt, &replyTo); err != nil {
			return nil, err
		}
		if replyTo.Valid {
			comment.ReplyTo = &replyTo.Int64
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	edges := make([]model.CommentEdge, len(comments))
	for i, comment := range comments {
		edges[i] = model.CommentEdge{
			Cursor: cursor.Encode(comment.ID),
			Node:   comment,
		}
	}

	pageInfo := model.PageInfo{
		HasNextPage: len(comments) == first,
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
