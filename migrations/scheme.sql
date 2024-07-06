CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    author TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    allow_comments BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    author TEXT NOT NULL,
    post_id integer REFERENCES posts,
    content TEXT NOT NULL check (length(content) < 2000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reply_to integer REFERENCES comments
);