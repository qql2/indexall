-- name: CreateResource :one
INSERT INTO resources (id, source, external_id, title, description, url, open_with, metadata, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING *;

-- name: GetResource :one
SELECT * FROM resources WHERE id = ? AND status = 'active';

-- name: GetResourceBySourceAndExternalId :one
SELECT * FROM resources
WHERE source = ? AND external_id = ? AND status = 'active';

-- name: UpdateResource :exec
UPDATE resources
SET title = COALESCE(?, title),
    description = COALESCE(?, description),
    url = COALESCE(?, url),
    open_with = COALESCE(?, open_with),
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteResource :exec
DELETE FROM resources WHERE id = ?;

-- name: ListResources :many
SELECT r.id, r.source, r.external_id, r.title, r.description, r.url, r.open_with, r.metadata, r.status, r.synced_at, r.created_at, r.updated_at
FROM resources r
WHERE r.status = COALESCE(?, r.status)
ORDER BY r.created_at DESC
LIMIT ? OFFSET ?;

-- name: ListResourcesByTag :many
-- Get resources with a specific tag (direct only, recursion handled in app)
SELECT DISTINCT r.id, r.source, r.external_id, r.title, r.description, r.url, r.open_with, r.metadata, r.status, r.synced_at, r.created_at, r.updated_at
FROM resources r
JOIN resource_tags rt ON r.id = rt.resource_id
WHERE rt.tag_id = ?
  AND r.status = COALESCE(?, r.status)
ORDER BY r.created_at DESC
LIMIT ? OFFSET ?;

-- name: CountResources :one
SELECT COUNT(*) as count FROM resources WHERE status = ?;

-- name: CountResourcesByTag :one
-- Count resources with a specific tag (direct only, recursion handled in app)
SELECT COUNT(DISTINCT r.id) as count
FROM resources r
JOIN resource_tags rt ON r.id = rt.resource_id
WHERE rt.tag_id = ?
  AND r.status = COALESCE(?, r.status);

-- name: SearchResources :many
-- Full-text search resources by title and description
SELECT
  r.id,
  r.source,
  r.title,
  r.description,
  r.url,
  r.created_at
FROM resources r
WHERE r.status = 'active'
  AND (r.title LIKE ? OR r.description LIKE ?)
ORDER BY r.created_at DESC
LIMIT ? OFFSET ?;

-- name: CountSearchResults :one
SELECT COUNT(DISTINCT r.id) as count
FROM resources r
WHERE r.status = 'active'
  AND (r.title LIKE ? OR r.description LIKE ?);

-- name: GetResourceByUrl :one
SELECT id, title FROM resources WHERE url = ? AND status = 'active' LIMIT 1;

-- name: AddTagToResource :exec
INSERT INTO resource_tags (resource_id, tag_id, created_at)
VALUES (?, ?, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;

-- name: RemoveTagFromResource :exec
DELETE FROM resource_tags WHERE resource_id = ? AND tag_id = ?;

-- name: RemoveAllTagsFromResource :exec
DELETE FROM resource_tags WHERE resource_id = ?;

-- name: GetResourceTags :many
SELECT rt.tag_id, t.name, t.color
FROM resource_tags rt
JOIN tags t ON rt.tag_id = t.id
WHERE rt.resource_id = ?
ORDER BY t.name ASC;

-- name: MarkResourceAsStale :exec
UPDATE resources
SET status = 'stale', updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: MarkResourceAsActive :exec
UPDATE resources
SET status = 'active', updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdateSyncTime :exec
UPDATE resources
SET synced_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: GetResourcesNeedingSync :many
-- Get resources that haven't been synced in a while
SELECT id, source, external_id, title, synced_at
FROM resources
WHERE source != 'manual'
  AND status = 'active'
  AND (synced_at IS NULL OR synced_at < datetime('now', '-1 day'))
ORDER BY synced_at ASC
LIMIT ?;

-- name: CountResourcesBySource :many
SELECT source, COUNT(*) as count
FROM resources
WHERE status = 'active'
GROUP BY source
ORDER BY count DESC;

-- name: GetStaleResources :many
SELECT id, source, external_id, title
FROM resources
WHERE status = 'stale'
ORDER BY updated_at DESC
LIMIT ? OFFSET ?;

-- name: CountStaleResources :one
SELECT COUNT(*) as count FROM resources WHERE status = 'stale';

-- name: GetDeletedResources :many
SELECT id, source, external_id, title, updated_at
FROM resources
WHERE status = 'deleted'
ORDER BY updated_at DESC
LIMIT ? OFFSET ?;

-- name: CountDeletedResources :one
SELECT COUNT(*) as count FROM resources WHERE status = 'deleted';

-- name: PermanentlyDeleteResource :exec
-- Permanently remove a resource (used for cleanup)
DELETE FROM resources WHERE id = ? AND status = 'deleted';

-- name: GetResourcesWithoutTags :many
SELECT r.id, r.title, r.source
FROM resources r
LEFT JOIN resource_tags rt ON r.id = rt.resource_id
WHERE rt.resource_id IS NULL
  AND r.status = 'active'
ORDER BY r.created_at DESC
LIMIT ? OFFSET ?;
