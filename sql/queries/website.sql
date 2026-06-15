-- name: GetWebsitePageBySlug :one
SELECT id, slug, title, content FROM website_pages WHERE slug = $1 AND deleted_at IS NULL;
