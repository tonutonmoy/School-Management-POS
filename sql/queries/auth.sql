-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetValidPasswordResetToken :one
SELECT * FROM password_reset_tokens
WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()
ORDER BY created_at DESC
LIMIT 1;

-- name: MarkPasswordResetTokenUsed :exec
UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1;

-- name: RevokeToken :exec
INSERT INTO revoked_tokens (jti, expires_at)
VALUES ($1, $2)
ON CONFLICT (jti) DO NOTHING;

-- name: IsTokenRevoked :one
SELECT EXISTS(SELECT 1 FROM revoked_tokens WHERE jti = $1) AS revoked;

-- name: CleanupExpiredRevokedTokens :exec
DELETE FROM revoked_tokens WHERE expires_at < NOW();

-- name: CleanupExpiredPasswordResetTokens :exec
DELETE FROM password_reset_tokens WHERE expires_at < NOW() OR used_at IS NOT NULL;
