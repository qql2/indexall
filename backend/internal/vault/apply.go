package vault

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/construct/indexall/internal/db/gen"
)

// ApplyEntry applies a vault entry to the SQLite database idempotently.
// All create operations use INSERT OR IGNORE so replaying is always safe.
func ApplyEntry(ctx context.Context, database *sql.DB, q *gen.Queries, entry Entry) error {
	switch entry.Type {
	case EntityTag:
		return applyTagOp(ctx, database, entry)
	case EntityResource:
		return applyResourceOp(ctx, database, entry)
	case EntityTagAlias:
		return applyTagAliasOp(ctx, database, entry)
	case EntityTagRelation:
		return applyTagRelationOp(ctx, database, entry)
	case EntityResourceTag:
		return applyResourceTagOp(ctx, database, entry)
	default:
		return fmt.Errorf("vault: unknown entity type: %s", entry.Type)
	}
}

func applyTagOp(ctx context.Context, db *sql.DB, entry Entry) error {
	switch entry.Op {
	case OpCreate:
		var d TagData
		if err := remarshal(entry.Data, &d); err != nil {
			return err
		}
		_, err := db.ExecContext(ctx,
			`INSERT OR IGNORE INTO tags (id, name, color, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
			d.ID, d.Name, nullableStr(d.Color), d.CreatedAt, d.CreatedAt)
		if err != nil {
			return err
		}
		// Apply initial aliases
		for _, alias := range d.Aliases {
			_, _ = db.ExecContext(ctx,
				`INSERT OR IGNORE INTO tag_aliases (id, tag_id, alias) VALUES (lower(hex(randomblob(16))), ?, ?)`,
				d.ID, alias)
		}
		// Apply initial parent relations
		for _, parentID := range d.Parents {
			_, _ = db.ExecContext(ctx,
				`INSERT OR IGNORE INTO tag_relations (parent_id, child_id) VALUES (?, ?)`,
				parentID, d.ID)
		}
		return nil

	case OpUpdate:
		var d TagData
		if err := remarshal(entry.Data, &d); err != nil {
			return err
		}
		_, err := db.ExecContext(ctx,
			`UPDATE tags SET
				name = CASE WHEN ? != '' THEN ? ELSE name END,
				color = CASE WHEN ? != '' THEN ? ELSE color END,
				updated_at = ?
			WHERE id = ?`,
			d.Name, d.Name, d.Color, d.Color, entry.TS, d.ID)
		return err

	case OpDelete:
		_, err := db.ExecContext(ctx, `DELETE FROM tags WHERE id = ?`, entry.EntityID)
		return err
	}
	return nil
}

func applyResourceOp(ctx context.Context, db *sql.DB, entry Entry) error {
	switch entry.Op {
	case OpCreate:
		var d ResourceData
		if err := remarshal(entry.Data, &d); err != nil {
			return err
		}
		_, err := db.ExecContext(ctx,
			`INSERT OR IGNORE INTO resources (id, source, external_id, title, description, url, open_with, status, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			d.ID, d.Source, nullableStr(d.ExternalID), d.Title,
			nullableStr(d.Description), nullableStr(d.URL), nullableStr(d.OpenWith),
			coalesceStr(d.Status, "active"), d.CreatedAt, d.CreatedAt)
		if err != nil {
			return err
		}
		// Apply initial tags
		for _, tagID := range d.Tags {
			_, _ = db.ExecContext(ctx,
				`INSERT OR IGNORE INTO resource_tags (resource_id, tag_id) VALUES (?, ?)`,
				d.ID, tagID)
		}
		return nil

	case OpUpdate:
		var d ResourceData
		if err := remarshal(entry.Data, &d); err != nil {
			return err
		}
		_, err := db.ExecContext(ctx,
			`UPDATE resources SET
				title = CASE WHEN ? != '' THEN ? ELSE title END,
				description = CASE WHEN ? != '' THEN ? ELSE description END,
				url = CASE WHEN ? != '' THEN ? ELSE url END,
				open_with = CASE WHEN ? != '' THEN ? ELSE open_with END,
				updated_at = ?
			WHERE id = ?`,
			d.Title, d.Title, d.Description, d.Description,
			d.URL, d.URL, d.OpenWith, d.OpenWith, entry.TS, d.ID)
		return err

	case OpDelete:
		_, err := db.ExecContext(ctx, `DELETE FROM resources WHERE id = ?`, entry.EntityID)
		return err
	}
	return nil
}

func applyTagAliasOp(ctx context.Context, db *sql.DB, entry Entry) error {
	switch entry.Op {
	case OpCreate:
		var d TagAliasData
		if err := remarshal(entry.Data, &d); err != nil {
			return err
		}
		_, err := db.ExecContext(ctx,
			`INSERT OR IGNORE INTO tag_aliases (id, tag_id, alias) VALUES (?, ?, ?)`,
			d.ID, d.TagID, d.Alias)
		return err

	case OpDelete:
		var d TagAliasData
		if err := remarshal(entry.Data, &d); err != nil {
			return err
		}
		if d.ID != "" {
			_, err := db.ExecContext(ctx, `DELETE FROM tag_aliases WHERE id = ?`, d.ID)
			return err
		}
		_, err := db.ExecContext(ctx, `DELETE FROM tag_aliases WHERE tag_id = ? AND alias = ?`, d.TagID, d.Alias)
		return err
	}
	return nil
}

func applyTagRelationOp(ctx context.Context, db *sql.DB, entry Entry) error {
	var d TagRelationData
	if err := remarshal(entry.Data, &d); err != nil {
		return err
	}
	switch entry.Op {
	case OpCreate:
		_, err := db.ExecContext(ctx,
			`INSERT OR IGNORE INTO tag_relations (parent_id, child_id) VALUES (?, ?)`,
			d.ParentID, d.ChildID)
		return err
	case OpDelete:
		_, err := db.ExecContext(ctx,
			`DELETE FROM tag_relations WHERE parent_id = ? AND child_id = ?`,
			d.ParentID, d.ChildID)
		return err
	}
	return nil
}

func applyResourceTagOp(ctx context.Context, db *sql.DB, entry Entry) error {
	var d ResourceTagData
	if err := remarshal(entry.Data, &d); err != nil {
		return err
	}
	switch entry.Op {
	case OpCreate:
		_, err := db.ExecContext(ctx,
			`INSERT OR IGNORE INTO resource_tags (resource_id, tag_id) VALUES (?, ?)`,
			d.ResourceID, d.TagID)
		return err
	case OpDelete:
		_, err := db.ExecContext(ctx,
			`DELETE FROM resource_tags WHERE resource_id = ? AND tag_id = ?`,
			d.ResourceID, d.TagID)
		return err
	}
	return nil
}

// remarshal round-trips through JSON to convert map[string]any → typed struct.
func remarshal(data any, target any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("vault: remarshal marshal: %w", err)
	}
	if err := json.Unmarshal(b, target); err != nil {
		return fmt.Errorf("vault: remarshal unmarshal: %w", err)
	}
	return nil
}

func nullableStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func coalesceStr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
