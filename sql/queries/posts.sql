-- name: CreatePost :one
  INSERT INTO posts(id, created_at, updated_at, title, url, description, published_at, feed_id)
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
  RETURNING *;

-- name: GetPostByFeedId :many
  SELECT * FROM posts WHERE feed_id = $1;  

-- name: GetPostsByUser :many
  SELECT * FROM posts JOIN feeds ON feed_id = feeds.id GROUP BY user_id HAVING user_id = $1 LIMIT $2;
