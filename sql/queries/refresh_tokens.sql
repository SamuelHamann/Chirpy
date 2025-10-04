-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token, expires_at, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens 
WHERE token = $1 
  AND revoked_at IS NULL 
  AND expires_at > NOW();
  
-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE token = $1
RETURNING *;
