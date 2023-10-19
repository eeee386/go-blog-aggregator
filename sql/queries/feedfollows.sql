-- name: CreateFeedFollow :one
  INSERT INTO feedfollows(id, created_at, updated_at, user_id, feed_id)
  VALUES ($1, $2, $3, $4, $5)
  RETURNING *;

-- name: DeleteFeedFollowById :exec
  DELETE FROM feedFollows WHERE id = $1;

-- name: GetAllFeedFollowsByUserID :many
  SELECT * FROM feedFollows WHERE user_id = $1;  
