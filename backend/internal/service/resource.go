package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	indexallv1 "github.com/construct/indexall/api/indexall/v1"
	"github.com/construct/indexall/internal/db/gen"
	"github.com/construct/indexall/internal/vault"
)

// Ensure ResourceService implements ResourceServiceServer interface
var _ indexallv1.ResourceServiceServer = (*ResourceService)(nil)

type ResourceService struct {
	indexallv1.UnimplementedResourceServiceServer
	db *sql.DB
	q  *gen.Queries
	v  *vault.Vault
}

func NewResourceService(db *sql.DB, q *gen.Queries, v *vault.Vault) *ResourceService {
	return &ResourceService{db: db, q: q, v: v}
}

func (s *ResourceService) logVault(op vault.OpType, entityType vault.EntityType, entityID string, data any) {
	if s.v == nil {
		return
	}
	if err := s.v.Append(op, entityType, entityID, data); err != nil {
		fmt.Printf("vault append error: %v\n", err)
	}
}

func (s *ResourceService) Create(ctx context.Context, req *indexallv1.CreateResourceRequest) (*indexallv1.CreateResourceResponse, error) {
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	// Dereference pointer fields with defaults
	source := "manual"
	if req.Source != nil && *req.Source != "" {
		source = *req.Source
	}

	var externalID sql.NullString
	if req.ExternalId != nil && *req.ExternalId != "" {
		externalID = sql.NullString{String: *req.ExternalId, Valid: true}
	}

	// Check uniqueness of source + externalId if both provided
	if source != "manual" && externalID.Valid {
		_, err := s.q.GetResourceBySourceAndExternalId(ctx, gen.GetResourceBySourceAndExternalIdParams{
			Source:     source,
			ExternalID: externalID,
		})
		if err == nil {
			return nil, status.Errorf(codes.AlreadyExists, "resource with source %q and external_id %q already exists", source, externalID.String)
		} else if err != sql.ErrNoRows {
			return nil, status.Errorf(codes.Internal, "failed to check resource: %v", err)
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to begin transaction: %v", err)
	}
	defer tx.Rollback()
	qtx := s.q.WithTx(tx)

	resourceID := uuid.New().String()
	resource, err := qtx.CreateResource(ctx, gen.CreateResourceParams{
		ID:          resourceID,
		Source:      source,
		ExternalID:  externalID,
		Title:       req.Title,
		Description: pointerToNullString(req.Description),
		Url:         pointerToNullString(req.Url),
		OpenWith:    pointerToNullString(req.OpenWith),
		Metadata:    sql.NullString{Valid: false},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create resource: %v", err)
	}

	// Add tags
	for _, tagID := range req.TagIds {
		if err := qtx.AddTagToResource(ctx, gen.AddTagToResourceParams{
			ResourceID: resourceID,
			TagID:      tagID,
		}); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to add tag %q: %v", tagID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
	}

	// Get tags for response
	tags := make([]*indexallv1.ResourceTag, 0)
	tagRows, err := s.q.GetResourceTags(ctx, resourceID)
	if err == nil {
		for _, t := range tagRows {
			tags = append(tags, &indexallv1.ResourceTag{
				Id:    t.TagID,
				Name:  t.Name,
				Color: nullStringToPointer(t.Color),
			})
		}
	}

	s.logVault(vault.OpCreate, vault.EntityResource, resourceID, vault.ResourceData{
		ID:         resourceID,
		Source:     source,
		ExternalID: externalID.String,
		Title:      resource.Title,
		URL:        derefString(nullStringToPointer(resource.Url), ""),
		Status:     "active",
		Tags:       req.TagIds,
		CreatedAt:  nullTimeToString(resource.CreatedAt),
	})

	resp := &indexallv1.CreateResourceResponse{
		Id:        resource.ID,
		Title:     resource.Title,
		Url:       nullStringToPointer(resource.Url),
		Source:    resource.Source,
		Tags:      tags,
		CreatedAt: nullTimeToString(resource.CreatedAt),
	}

	return resp, nil
}

func (s *ResourceService) Update(ctx context.Context, req *indexallv1.UpdateResourceRequest) (*indexallv1.UpdateResourceResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Verify resource exists
	existing, err := s.q.GetResource(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get resource: %v", err)
	}

	// Use existing values if not provided
	title := existing.Title
	if req.Title != nil && *req.Title != "" {
		title = *req.Title
	}

	// If external_id is being updated (file moved/renamed), use specialized query
	if req.ExternalId != nil && *req.ExternalId != "" {
		err = s.q.UpdateResourceExternalId(ctx, gen.UpdateResourceExternalIdParams{
			ID:         req.Id,
			ExternalID: sql.NullString{String: *req.ExternalId, Valid: true},
			Url:        pointerToNullString(req.Url),
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update resource external_id: %v", err)
		}
	}

	err = s.q.UpdateResource(ctx, gen.UpdateResourceParams{
		ID:          req.Id,
		Title:       title,
		Description: pointerToNullString(req.Description),
		Url:         pointerToNullString(req.Url),
		OpenWith:    pointerToNullString(req.OpenWith),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update resource: %v", err)
	}

	s.logVault(vault.OpUpdate, vault.EntityResource, req.Id, vault.ResourceData{
		ID:          req.Id,
		Title:       title,
		Description: derefString(req.Description, ""),
		URL:         derefString(req.Url, ""),
		OpenWith:    derefString(req.OpenWith, ""),
	})

	return &indexallv1.UpdateResourceResponse{
		Success: true,
	}, nil
}

func (s *ResourceService) Delete(ctx context.Context, req *indexallv1.DeleteResourceRequest) (*indexallv1.DeleteResourceResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Verify resource exists
	if _, err := s.q.GetResource(ctx, req.Id); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get resource: %v", err)
	}

	err := s.q.DeleteResource(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete resource: %v", err)
	}

	s.logVault(vault.OpDelete, vault.EntityResource, req.Id, nil)

	return &indexallv1.DeleteResourceResponse{
		Success: true,
	}, nil
}

func (s *ResourceService) Get(ctx context.Context, req *indexallv1.GetResourceRequest) (*indexallv1.GetResourceResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	resource, err := s.q.GetResource(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get resource: %v", err)
	}

	// Get tags
	tags := make([]*indexallv1.ResourceTag, 0)
	tagRows, err := s.q.GetResourceTags(ctx, req.Id)
	if err == nil {
		for _, t := range tagRows {
			tags = append(tags, &indexallv1.ResourceTag{
				Id:    t.TagID,
				Name:  t.Name,
				Color: nullStringToPointer(t.Color),
			})
		}
	}

	return &indexallv1.GetResourceResponse{
		Id:          resource.ID,
		Source:      resource.Source,
		ExternalId:  nullStringToPointer(resource.ExternalID),
		Title:       resource.Title,
		Description: nullStringToPointer(resource.Description),
		Url:         nullStringToPointer(resource.Url),
		OpenWith:    nullStringToPointer(resource.OpenWith),
		Metadata:    nullStringToPointer(resource.Metadata),
		Status:      parseResourceStatus(resource.Status),
		CreatedAt:   nullTimeToString(resource.CreatedAt),
		UpdatedAt:   nullTimeToString(resource.UpdatedAt),
		Tags:        tags,
	}, nil
}

func (s *ResourceService) GetByUrl(ctx context.Context, req *indexallv1.GetByUrlRequest) (*indexallv1.GetByUrlResponse, error) {
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	// Validate URL
	if _, err := url.Parse(req.Url); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid url: %v", err)
	}

	resourceRow, err := s.q.GetResourceByUrl(ctx, sql.NullString{String: req.Url, Valid: true})
	if err != nil {
		if err == sql.ErrNoRows {
			return &indexallv1.GetByUrlResponse{}, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to get resource: %v", err)
	}

	// Get tags
	tags := make([]*indexallv1.ResourceTag, 0)
	tagRows, err := s.q.GetResourceTags(ctx, resourceRow.ID)
	if err == nil {
		for _, t := range tagRows {
			tags = append(tags, &indexallv1.ResourceTag{
				Id:    t.TagID,
				Name:  t.Name,
				Color: nullStringToPointer(t.Color),
			})
		}
	}

	return &indexallv1.GetByUrlResponse{
		Resource: &indexallv1.GetByUrlResponse_Resource{
			Id:   resourceRow.ID,
			Title: resourceRow.Title,
			Tags: tags,
		},
	}, nil
}

func (s *ResourceService) GetByExternalId(ctx context.Context, req *indexallv1.GetByExternalIdRequest) (*indexallv1.GetByExternalIdResponse, error) {
	if req.Source == "" || req.ExternalId == "" {
		return nil, status.Error(codes.InvalidArgument, "source and external_id are required")
	}

	resource, err := s.q.GetResourceBySourceAndExternalId(ctx, gen.GetResourceBySourceAndExternalIdParams{
		Source:     req.Source,
		ExternalID: sql.NullString{String: req.ExternalId, Valid: true},
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return &indexallv1.GetByExternalIdResponse{}, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to get resource: %v", err)
	}

	return &indexallv1.GetByExternalIdResponse{
		Id:    &resource.ID,
		Title: &resource.Title,
	}, nil
}

func (s *ResourceService) AddTag(ctx context.Context, req *indexallv1.AddTagRequest) (*indexallv1.AddTagResponse, error) {
	if req.ResourceId == "" || req.TagId == "" {
		return nil, status.Error(codes.InvalidArgument, "resource_id and tag_id are required")
	}

	// Verify resource exists
	if _, err := s.q.GetResource(ctx, req.ResourceId); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get resource: %v", err)
	}

	// Verify tag exists
	if _, err := s.q.GetTag(ctx, req.TagId); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "tag not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get tag: %v", err)
	}

	err := s.q.AddTagToResource(ctx, gen.AddTagToResourceParams{
		ResourceID: req.ResourceId,
		TagID:      req.TagId,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add tag: %v", err)
	}

	s.logVault(vault.OpCreate, vault.EntityResourceTag, req.ResourceId, vault.ResourceTagData{
		ResourceID: req.ResourceId,
		TagID:      req.TagId,
	})

	return &indexallv1.AddTagResponse{
		Success: true,
	}, nil
}

func (s *ResourceService) RemoveTag(ctx context.Context, req *indexallv1.RemoveTagRequest) (*indexallv1.RemoveTagResponse, error) {
	if req.ResourceId == "" || req.TagId == "" {
		return nil, status.Error(codes.InvalidArgument, "resource_id and tag_id are required")
	}

	// Verify resource exists
	if _, err := s.q.GetResource(ctx, req.ResourceId); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get resource: %v", err)
	}

	// Verify tag exists
	if _, err := s.q.GetTag(ctx, req.TagId); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "tag %q not found", req.TagId)
		}
		return nil, status.Errorf(codes.Internal, "failed to get tag: %v", err)
	}

	err := s.q.RemoveTagFromResource(ctx, gen.RemoveTagFromResourceParams{
		ResourceID: req.ResourceId,
		TagID:      req.TagId,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove tag: %v", err)
	}

	s.logVault(vault.OpDelete, vault.EntityResourceTag, req.ResourceId, vault.ResourceTagData{
		ResourceID: req.ResourceId,
		TagID:      req.TagId,
	})

	return &indexallv1.RemoveTagResponse{
		Success: true,
	}, nil
}

// listAllResources returns all resources with pagination
func (s *ResourceService) listAllResources(ctx context.Context, offset, limit int32) ([]gen.Resource, int64, error) {
	// Count all resources directly (CountResources uses "WHERE status = ?" which fails for NULL)
	var total int64
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM resources`).Scan(&total)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "failed to count resources: %v", err)
	}

	// List resources with pagination
	resources, err := s.q.ListResources(ctx, gen.ListResourcesParams{
		Status: sql.NullString{Valid: false}, // Get all statuses
		Offset: int64(offset),
		Limit:  int64(limit),
	})
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "failed to list resources: %v", err)
	}

	return resources, total, nil
}

func (s *ResourceService) Query(ctx context.Context, req *indexallv1.ResourceQueryRequest) (*indexallv1.ResourceQueryResponse, error) {
	// Set defaults
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var resources []gen.Resource
	var total int64
	var err error

	switch queryType := req.Query.(type) {
	case *indexallv1.ResourceQueryRequest_TagQuery:
		resources, total, err = s.queryByTag(ctx, queryType.TagQuery, offset, pageSize)
	case *indexallv1.ResourceQueryRequest_KeywordQuery:
		resources, total, err = s.queryByKeyword(ctx, queryType.KeywordQuery, offset, pageSize)
	default:
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}

	if err != nil {
		return nil, err
	}

	// Batch-fetch all tags+aliases for the page in a single query
	// resourceID -> tagID -> TagInfo
	resourceTagMap := make(map[string]map[string]*indexallv1.TagInfo)
	if len(resources) > 0 {
		ids := make([]any, len(resources))
		for i, r := range resources {
			ids[i] = r.ID
		}
		placeholders := strings.Repeat("?,", len(ids))
		placeholders = placeholders[:len(placeholders)-1]
		batchQuery := fmt.Sprintf(`
			SELECT rt.resource_id, rt.tag_id, t.name, ta.alias
			FROM resource_tags rt
			JOIN tags t ON rt.tag_id = t.id
			LEFT JOIN tag_aliases ta ON t.id = ta.tag_id
			WHERE rt.resource_id IN (%s)
			ORDER BY rt.resource_id, t.name, ta.alias`, placeholders)
		rows, qErr := s.db.QueryContext(ctx, batchQuery, ids...)
		if qErr == nil {
			defer rows.Close()
			for rows.Next() {
				var resourceID, tagID, name string
				var alias sql.NullString
				if err := rows.Scan(&resourceID, &tagID, &name, &alias); err != nil {
					continue
				}
				if resourceTagMap[resourceID] == nil {
					resourceTagMap[resourceID] = make(map[string]*indexallv1.TagInfo)
				}
				if _, ok := resourceTagMap[resourceID][tagID]; !ok {
					resourceTagMap[resourceID][tagID] = &indexallv1.TagInfo{
						Id:      tagID,
						Name:    name,
						Aliases: []string{},
					}
				}
				if alias.Valid {
					resourceTagMap[resourceID][tagID].Aliases = append(resourceTagMap[resourceID][tagID].Aliases, alias.String)
				}
			}
		}
	}

	items := make([]*indexallv1.ResourceSearchResult, len(resources))
	for i, resource := range resources {
		tags := make([]*indexallv1.TagInfo, 0)
		for _, t := range resourceTagMap[resource.ID] {
			tags = append(tags, t)
		}
		items[i] = &indexallv1.ResourceSearchResult{
			Id:          resource.ID,
			Source:      resource.Source,
			Title:       resource.Title,
			Description: nullStringToPointer(resource.Description),
			Url:         nullStringToPointer(resource.Url),
			CreatedAt:   nullTimeToString(resource.CreatedAt),
			UpdatedAt:   nullTimeToString(resource.UpdatedAt),
			Tags:        tags,
			MatchSource: indexallv1.MatchSource_MATCH_SOURCE_UNSPECIFIED,
		}
	}

	return &indexallv1.ResourceQueryResponse{
		Items:    items,
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// queryByTag queries resources by tag with different scopes
func (s *ResourceService) queryByTag(ctx context.Context, tq *indexallv1.TagQuery, offset, limit int32) ([]gen.Resource, int64, error) {
	if tq.TagId == "" {
		return nil, 0, status.Error(codes.InvalidArgument, "tag_id is required")
	}

	// Verify tag exists
	if _, err := s.q.GetTag(ctx, tq.TagId); err != nil {
		if err == sql.ErrNoRows {
			return nil, 0, status.Error(codes.NotFound, "tag not found")
		}
		return nil, 0, status.Errorf(codes.Internal, "failed to get tag: %v", err)
	}

	// Build recursive CTE based on scope
	var withClause string
	switch tq.TagScope {
	case indexallv1.TagQuery_DIRECT:
		withClause = "WITH tag_ids AS (SELECT ? AS id)"
	case indexallv1.TagQuery_WITH_ANCESTORS:
		withClause = `WITH RECURSIVE ancestors AS (
			SELECT id FROM tags WHERE id = ?
			UNION ALL
			SELECT tr.parent_id FROM tag_relations tr JOIN ancestors a ON tr.child_id = a.id
		), tag_ids AS (SELECT id FROM ancestors)`
	case indexallv1.TagQuery_WITH_DESCENDANTS:
		withClause = `WITH RECURSIVE descendants AS (
			SELECT id FROM tags WHERE id = ?
			UNION ALL
			SELECT tr.child_id FROM tag_relations tr JOIN descendants d ON tr.parent_id = d.id
		), tag_ids AS (SELECT id FROM descendants)`
	default:
		return nil, 0, status.Error(codes.InvalidArgument, "invalid tag_scope")
	}

	// Query total count
	countQuery := withClause + `
		SELECT COUNT(DISTINCT r.id) FROM resources r
		JOIN resource_tags rt ON r.id = rt.resource_id
		WHERE rt.tag_id IN (SELECT id FROM tag_ids)`

	var total int64
	err := s.db.QueryRowContext(ctx, countQuery, tq.TagId).Scan(&total)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "failed to count resources: %v", err)
	}

	// Query resources
	dataQuery := withClause + `
		SELECT DISTINCT r.id, r.source, r.external_id, r.title, r.description, r.url, r.open_with, r.metadata, r.status, r.synced_at, r.created_at, r.updated_at
		FROM resources r
		JOIN resource_tags rt ON r.id = rt.resource_id
		WHERE rt.tag_id IN (SELECT id FROM tag_ids)
		ORDER BY r.created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := s.db.QueryContext(ctx, dataQuery, tq.TagId, limit, offset)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "failed to query resources: %v", err)
	}
	defer rows.Close()

	var resources []gen.Resource
	for rows.Next() {
		var r gen.Resource
		err := rows.Scan(&r.ID, &r.Source, &r.ExternalID, &r.Title, &r.Description, &r.Url, &r.OpenWith, &r.Metadata, &r.Status, &r.SyncedAt, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return nil, 0, status.Errorf(codes.Internal, "failed to scan resource: %v", err)
		}
		resources = append(resources, r)
	}

	return resources, total, nil
}

// queryByKeywordDirect handles keyword search with DIRECT tag scope.
// Searches resource title/description AND tag names/aliases.
func (s *ResourceService) queryByKeywordDirect(ctx context.Context, kq *indexallv1.KeywordQuery, likeKeyword string, offset, limit int32) ([]gen.Resource, int64, error) {
	tagMatchSub := `r.id IN (
		SELECT rt.resource_id FROM resource_tags rt
		JOIN tags t ON rt.tag_id = t.id
		LEFT JOIN tag_aliases ta ON t.id = ta.tag_id
		WHERE t.name LIKE ? OR ta.alias LIKE ?
	)`

	var matchCond string
	var params []any
	switch kq.FieldScope {
	case indexallv1.KeywordQuery_ALL:
		matchCond = "(r.title LIKE ? OR r.description LIKE ? OR " + tagMatchSub + ")"
		params = []any{likeKeyword, likeKeyword, likeKeyword, likeKeyword}
	case indexallv1.KeywordQuery_TITLE:
		matchCond = "(r.title LIKE ? OR " + tagMatchSub + ")"
		params = []any{likeKeyword, likeKeyword, likeKeyword}
	case indexallv1.KeywordQuery_DESCRIPTION:
		matchCond = "(r.description LIKE ? OR " + tagMatchSub + ")"
		params = []any{likeKeyword, likeKeyword, likeKeyword}
	default:
		return nil, 0, status.Error(codes.InvalidArgument, "invalid field_scope")
	}

	var total int64
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT r.id) FROM resources r WHERE "+matchCond, params...).Scan(&total); err != nil && err != sql.ErrNoRows {
		return nil, 0, status.Errorf(codes.Internal, "failed to count resources: %v", err)
	}

	dataParams := make([]any, 0, len(params)+2)
	dataParams = append(dataParams, params...)
	dataParams = append(dataParams, limit, offset)
	rows, err := s.db.QueryContext(ctx,
		`SELECT DISTINCT r.id, r.source, r.external_id, r.title, r.description, r.url, r.open_with, r.metadata, r.status, r.synced_at, r.created_at, r.updated_at
		FROM resources r WHERE `+matchCond+` ORDER BY r.created_at DESC LIMIT ? OFFSET ?`,
		dataParams...)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "failed to query resources: %v", err)
	}
	defer rows.Close()

	var resources []gen.Resource
	for rows.Next() {
		var r gen.Resource
		if err := rows.Scan(&r.ID, &r.Source, &r.ExternalID, &r.Title, &r.Description, &r.Url, &r.OpenWith, &r.Metadata, &r.Status, &r.SyncedAt, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, 0, status.Errorf(codes.Internal, "failed to scan resource: %v", err)
		}
		resources = append(resources, r)
	}
	return resources, total, nil
}

// queryByKeyword queries resources by keyword using LIKE (FTS5 fallback).
// Empty keyword returns all resources.
func (s *ResourceService) queryByKeyword(ctx context.Context, kq *indexallv1.KeywordQuery, offset, limit int32) ([]gen.Resource, int64, error) {
	if kq.Keyword == "" {
		return s.listAllResources(ctx, offset, limit)
	}

	// Use LIKE for keyword matching (FTS5 fallback)
	likeKeyword := "%" + kq.Keyword + "%"

	// DIRECT scope: use dedicated handler that also searches tag names/aliases
	if kq.TagScope == indexallv1.KeywordQuery_DIRECT {
		return s.queryByKeywordDirect(ctx, kq, likeKeyword, offset, limit)
	}

	// Field scope: which resource fields to search directly
	var fieldMatchClause string
	fieldParamCount := 1
	switch kq.FieldScope {
	case indexallv1.KeywordQuery_ALL:
		fieldMatchClause = "(r.title LIKE ? OR r.description LIKE ?)"
		fieldParamCount = 2
	case indexallv1.KeywordQuery_TITLE:
		fieldMatchClause = "r.title LIKE ?"
	case indexallv1.KeywordQuery_DESCRIPTION:
		fieldMatchClause = "r.description LIKE ?"
	default:
		return nil, 0, status.Error(codes.InvalidArgument, "invalid field_scope")
	}

	// Tag scope CTE: start from tags matching keyword by name/alias, recurse in correct direction.
	//
	// WITH_ANCESTORS: "resource's tag has an ancestor matching keyword"
	//   = resource's tag is a DESCENDANT of a keyword-matching tag → recurse downward (child_id)
	//
	// WITH_DESCENDANTS: "resource's tag has a descendant matching keyword"
	//   = resource's tag is an ANCESTOR of a keyword-matching tag → recurse upward (parent_id)
	var recursiveSelect string
	switch kq.TagScope {
	case indexallv1.KeywordQuery_WITH_ANCESTORS:
		recursiveSelect = "SELECT tr.child_id FROM tag_relations tr JOIN tag_scope ts ON tr.parent_id = ts.id"
	case indexallv1.KeywordQuery_WITH_DESCENDANTS:
		recursiveSelect = "SELECT tr.parent_id FROM tag_relations tr JOIN tag_scope ts ON tr.child_id = ts.id"
	default:
		return nil, 0, status.Error(codes.InvalidArgument, "invalid tag_scope")
	}

	// Two-CTE approach: separate non-recursive seed from recursive expansion.
	// seed_tags: tags directly matching keyword in name or alias (2 params: name, alias).
	// tag_scope: seed + recursive expansion in chosen direction (no extra params).
	cte := `WITH RECURSIVE
	seed_tags(id) AS (
		SELECT t.id FROM tags t WHERE t.name LIKE ?
		UNION
		SELECT ta.tag_id FROM tag_aliases ta WHERE ta.alias LIKE ?
	),
	tag_scope(id) AS (
		SELECT id FROM seed_tags
		UNION ALL
		` + recursiveSelect + `
	)`

	// Unified WHERE: resource matches field directly OR is tagged with a tag in tag_scope
	tagMatchSub := `r.id IN (
		SELECT rt.resource_id FROM resource_tags rt
		WHERE rt.tag_id IN (SELECT id FROM tag_scope)
	)`
	whereClause := fieldMatchClause + " OR " + tagMatchSub

	// params: 2 (seed CTE: name, alias) + fieldParamCount (title/description match)
	params := make([]any, 0, 2+fieldParamCount)
	params = append(params, likeKeyword, likeKeyword)
	for i := 0; i < fieldParamCount; i++ {
		params = append(params, likeKeyword)
	}

	countQuery := cte + "\n\tSELECT COUNT(DISTINCT r.id) FROM resources r WHERE " + whereClause
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, params...).Scan(&total); err != nil && err != sql.ErrNoRows {
		return nil, 0, status.Errorf(codes.Internal, "failed to count resources: %v", err)
	}

	dataQuery := cte + `
	SELECT DISTINCT r.id, r.source, r.external_id, r.title, r.description, r.url, r.open_with, r.metadata, r.status, r.synced_at, r.created_at, r.updated_at
	FROM resources r WHERE ` + whereClause + `
	ORDER BY r.created_at DESC LIMIT ? OFFSET ?`

	dataParams := make([]any, 0, len(params)+2)
	dataParams = append(dataParams, params...)
	dataParams = append(dataParams, limit, offset)

	rows, err := s.db.QueryContext(ctx, dataQuery, dataParams...)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "failed to query resources: %v", err)
	}
	defer rows.Close()

	var resources []gen.Resource
	for rows.Next() {
		var r gen.Resource
		if scanErr := rows.Scan(&r.ID, &r.Source, &r.ExternalID, &r.Title, &r.Description, &r.Url, &r.OpenWith, &r.Metadata, &r.Status, &r.SyncedAt, &r.CreatedAt, &r.UpdatedAt); scanErr != nil {
			return nil, 0, status.Errorf(codes.Internal, "failed to scan resource: %v", scanErr)
		}
		resources = append(resources, r)
	}

	return resources, total, nil
}
