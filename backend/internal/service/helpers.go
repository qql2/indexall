package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/construct/indexall/internal/db/gen"
)

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
