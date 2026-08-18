-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
    RETURNING *
)
SELECT * FROM inserted_feed_follow
LEFT JOIN users ON inserted_feed_follow.user_id = users.id
LEFT JOIN feeds ON inserted_feed_follow.feed_id = feeds.id;