-- name: DeleteFeedFollow :one
WITH deleted_feed_follow AS (
    DELETE FROM feed_follows
    WHERE feed_follows.user_id = $1
      AND feed_follows.feed_id = $2
    RETURNING
        feed_follows.id,
        feed_follows.created_at,
        feed_follows.updated_at,
        feed_follows.user_id,
        feed_follows.feed_id
)
SELECT
    deleted_feed_follow.id AS feed_follow_id,
    deleted_feed_follow.created_at AS feed_follow_created_at,
    deleted_feed_follow.updated_at AS feed_follow_updated_at,
    deleted_feed_follow.user_id AS feed_follow_user_id,
    deleted_feed_follow.feed_id AS feed_follow_feed_id,
    users.id AS user_id,
    users.name AS user_name,
    feeds.id AS feed_id,
    feeds.name AS feed_name,
    feeds.url AS feed_url
FROM deleted_feed_follow
LEFT JOIN users ON deleted_feed_follow.user_id = users.id
LEFT JOIN feeds ON deleted_feed_follow.feed_id = feeds.id;


