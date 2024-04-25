-- name: UsersAll :many
SELECT * FROM users;

-- name: UsersByID :one
SELECT * FROM users WHERE id = $1;
