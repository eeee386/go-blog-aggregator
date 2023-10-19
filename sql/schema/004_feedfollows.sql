
-- +goose Up
CREATE TABLE feedfollows(
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  user_id UUID NOT NULL,
  feed_id UUID NOT NULL,
  CONSTRAINT userID
  FOREIGN KEY(user_id) REFERENCES users(id)
  ON DELETE CASCADE,
  CONSTRAINT feedId
  FOREIGN KEY(feed_id) REFERENCES feeds(id)
  ON DELETE CASCADE
);


-- +goose Down  
DROP TABLE feedfollows;
