package service

import (
	"context"
	"database/sql"
	"time"

	indexallv1 "github.com/construct/indexall/api/indexall/v1"
	"github.com/construct/indexall/internal/db/gen"
)

// getAncestorNames returns the ancestor tag names from root to immediate parent.
// Used for displaying match path in search results (e.g. ["Learning", "Python"]).
func getAncestorNames(ctx context.Context, db *sql.DB, tagID string) []string {
	// Walk up the DAG and collect ancestor ids with their depth
	rows, err := db.QueryContext(ctx, `
		WITH RECURSIVE ancestors AS (
			SELECT tr.parent_id AS id, 1 AS depth
			FROM tag_relations tr WHERE tr.child_id = ?
			UNION ALL
			SELECT tr.parent_id, a.depth + 1
			FROM tag_relations tr JOIN ancestors a ON tr.child_id = a.id
		)
		SELECT t.name, MAX(a.depth) AS depth
		FROM ancestors a JOIN tags t ON t.id = a.id
		GROUP BY a.id, t.name
		ORDER BY depth DESC`, tagID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		var depth int
		if err := rows.Scan(&name, &depth); err == nil {
			names = append(names, name)
		}
	}
	return names
}

// Helper functions for type conversions

func stringToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullStringToPointer(ns sql.NullString) *string {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	return &ns.String
}

func nullTimeToString(nt sql.NullTime) string {
	if !nt.Valid {
		return ""
	}
	return nt.Time.Format(time.RFC3339)
}

func pointerToNullString(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func derefString(s *string, defaultVal string) string {
	if s == nil {
		return defaultVal
	}
	return *s
}

func nilIfEmpty(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func parseResourceStatus(ns sql.NullString) indexallv1.ResourceStatus {
	if !ns.Valid {
		return indexallv1.ResourceStatus_RESOURCE_STATUS_ACTIVE
	}
	switch ns.String {
	case "stale":
		return indexallv1.ResourceStatus_RESOURCE_STATUS_STALE
	case "deleted":
		return indexallv1.ResourceStatus_RESOURCE_STATUS_DELETED
	default:
		return indexallv1.ResourceStatus_RESOURCE_STATUS_ACTIVE
	}
}

func getTagAliases(ctx context.Context, q *gen.Queries, tagID string) []string {
	aliases, err := q.ListAliasesByTag(ctx, tagID)
	if err != nil {
		return []string{}
	}
	result := make([]string, len(aliases))
	for i, alias := range aliases {
		result[i] = alias.Alias
	}
	return result
}

func getTagParents(ctx context.Context, q *gen.Queries, tagID string) []string {
	parents, err := q.ListParentTags(ctx, tagID)
	if err != nil {
		return []string{}
	}
	return parents
}

func getTagResourceCount(ctx context.Context, q *gen.Queries, tagID string) int32 {
	count, err := q.CountResourcesForTag(ctx, tagID)
	if err != nil {
		return 0
	}
	return int32(count)
}
