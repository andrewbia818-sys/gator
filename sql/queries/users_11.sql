-- name: GetPostsForUser :many
SELECT id, created_at, updated_at, title, description, url, feed_id
FROM posts
WHERE feed_id IN (
    SELECT feed_id FROM feed_follows WHERE user_id = $1
)
ORDER BY created_at DESC
LIMIT $2;

