-- name: GetPostByURL :one
SELECT id FROM posts WHERE url = $1 LIMIT 1;
