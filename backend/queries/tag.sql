-- name: CreateTag :one
INSERT INTO tags (id, name, color, created_at, updated_at)
VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING *;

-- name: GetTag :one
SELECT * FROM tags WHERE id = ?;

-- name: GetTagByName :one
SELECT * FROM tags WHERE name = ?;

-- name: UpdateTag :exec
UPDATE tags
SET name = COALESCE(?, name),
    color = COALESCE(?, color),
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteTag :exec
DELETE FROM tags WHERE id = ?;

-- name: ListTags :many
SELECT * FROM tags ORDER BY name ASC;

-- name: GetTagWithAliases :one
SELECT
  t.id,
  t.name,
  t.color,
  t.created_at,
  t.updated_at,
  COUNT(DISTINCT ta.id) as alias_count
FROM tags t
LEFT JOIN tag_aliases ta ON t.id = ta.tag_id
WHERE t.id = ?
GROUP BY t.id;

-- name: GetTagsWithCounts :many
SELECT
  t.id,
  t.name,
  t.color,
  t.created_at,
  t.updated_at,
  COUNT(DISTINCT rt.resource_id) as resource_count
FROM tags t
LEFT JOIN resource_tags rt ON t.id = rt.tag_id
GROUP BY t.id
ORDER BY t.name ASC;

-- name: ListAliasesByTag :many
SELECT id, alias FROM tag_aliases WHERE tag_id = ? ORDER BY alias ASC;

-- name: CreateAlias :one
INSERT INTO tag_aliases (id, tag_id, alias, created_at)
VALUES (?, ?, ?, CURRENT_TIMESTAMP)
RETURNING id, alias;

-- name: GetAliasByName :one
SELECT id, tag_id, alias FROM tag_aliases WHERE alias = ?;

-- name: DeleteAlias :exec
DELETE FROM tag_aliases WHERE id = ?;

-- name: DeleteAliasByTagId :exec
DELETE FROM tag_aliases WHERE tag_id = ?;

-- name: ListParentTags :many
SELECT parent_id FROM tag_relations WHERE child_id = ? ORDER BY parent_id ASC;

-- name: ListChildTags :many
SELECT child_id FROM tag_relations WHERE parent_id = ? ORDER BY child_id ASC;

-- name: CreateTagRelation :exec
INSERT INTO tag_relations (parent_id, child_id)
VALUES (?, ?);

-- name: DeleteTagRelation :exec
DELETE FROM tag_relations WHERE parent_id = ? AND child_id = ?;

-- name: DeleteTagRelationsByParent :exec
DELETE FROM tag_relations WHERE parent_id = ?;

-- name: DeleteTagRelationsByChild :exec
DELETE FROM tag_relations WHERE child_id = ?;

-- name: CheckCycleWouldExist :one
-- Check if adding (parent_id as parent of child_id) would create a cycle.
-- Cycle exists if child_id is already an ancestor of parent_id, or parent_id = child_id.
WITH RECURSIVE ancestors(id) AS (
  SELECT parent_id FROM tag_relations WHERE child_id = ?
  UNION ALL
  SELECT tr.parent_id FROM tag_relations tr
  INNER JOIN ancestors a ON tr.child_id = a.id
)
SELECT CASE WHEN EXISTS(SELECT 1 FROM ancestors WHERE id = ?) OR ? = ? THEN 1 ELSE 0 END as would_create_cycle;

-- name: GetTagTree :many
-- Get all root tags (tags with no parents) and build tree structure
-- This query returns all root tags; tree building is done in application code
SELECT
  t.id,
  t.name,
  t.color,
  COUNT(DISTINCT rt.resource_id) as resource_count
FROM tags t
LEFT JOIN resource_tags rt ON t.id = rt.tag_id
LEFT JOIN tag_relations tr ON t.id = tr.child_id
WHERE tr.child_id IS NULL
GROUP BY t.id
ORDER BY t.name ASC;

-- name: GetTagTreeNode :many
-- Get direct children of a tag
SELECT
  t.id,
  t.name,
  t.color,
  COUNT(DISTINCT rt.resource_id) as resource_count
FROM tags t
LEFT JOIN resource_tags rt ON t.id = rt.tag_id
WHERE t.id IN (SELECT child_id FROM tag_relations WHERE parent_id = ?)
GROUP BY t.id
ORDER BY t.name ASC;

-- name: SearchTags :many
-- Search tags by name or alias (prefix match)
SELECT DISTINCT
  t.id,
  t.name,
  t.color,
  CASE
    WHEN t.name LIKE ? THEN NULL
    ELSE ta.alias
  END as matched_alias
FROM tags t
LEFT JOIN tag_aliases ta ON t.id = ta.tag_id
WHERE t.name LIKE ? OR ta.alias LIKE ?
ORDER BY t.name ASC;

-- name: CountResourcesForTag :one
-- Count resources with a tag (direct only, recursion handled in app)
SELECT COUNT(DISTINCT rt.resource_id) as count
FROM resource_tags rt
WHERE rt.tag_id = ?;

-- name: GetTagsForResource :many
SELECT t.id, t.name, t.color FROM tags t
WHERE t.id IN (SELECT tag_id FROM resource_tags WHERE resource_id = ?)
ORDER BY t.name ASC;

-- name: ListAllTagAliases :many
SELECT tag_id, alias FROM tag_aliases ORDER BY tag_id ASC, alias ASC;

-- name: ListAllTagRelations :many
SELECT parent_id, child_id FROM tag_relations;
