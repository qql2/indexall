package service

import (
	"context"
	"database/sql"
	"net/url"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	indexallv1 "github.com/construct/indexall/api/indexall/v1"
	"github.com/construct/indexall/internal/db/gen"
)

// Ensure ResourceService implements ResourceServiceServer interface
var _ indexallv1.ResourceServiceServer = (*ResourceService)(nil)

type ResourceService struct {
	indexallv1.UnimplementedResourceServiceServer
	db *sql.DB
	q  *gen.Queries
}

func NewResourceService(db *sql.DB, q *gen.Queries) *ResourceService {
	return &ResourceService{db: db, q: q}
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

	resourceID := uuid.New().String()
	resource, err := s.q.CreateResource(ctx, gen.CreateResourceParams{
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
		_ = s.q.AddTagToResource(ctx, gen.AddTagToResourceParams{
			ResourceID: resourceID,
			TagID:      tagID,
		})
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
		Status:      indexallv1.ResourceStatus(indexallv1.ResourceStatus_RESOURCE_STATUS_ACTIVE), // TODO: parse from resource.Status
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

	err := s.q.RemoveTagFromResource(ctx, gen.RemoveTagFromResourceParams{
		ResourceID: req.ResourceId,
		TagID:      req.TagId,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove tag: %v", err)
	}

	return &indexallv1.RemoveTagResponse{
		Success: true,
	}, nil
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

	// Build response
	items := make([]*indexallv1.ResourceSearchResult, len(resources))
	for i, resource := range resources {
		// Get tags
		tags := make([]*indexallv1.TagInfo, 0)
		tagRows, err := s.q.GetResourceTags(ctx, resource.ID)
		if err == nil {
			for _, t := range tagRows {
				// Get tag aliases
				aliases := make([]string, 0)
				aliasRows, aliasErr := s.q.ListAliasesByTag(ctx, t.TagID)
				if aliasErr == nil {
					for _, a := range aliasRows {
						aliases = append(aliases, a.Alias)
					}
				}

				tags = append(tags, &indexallv1.TagInfo{
					Id:       t.TagID,
					Name:     t.Name,
					Aliases:  aliases,
				})
			}
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

// queryByKeyword queries resources by keyword using LIKE (FTS5 fallback)
func (s *ResourceService) queryByKeyword(ctx context.Context, kq *indexallv1.KeywordQuery, offset, limit int32) ([]gen.Resource, int64, error) {
	if kq.Keyword == "" {
		return nil, 0, status.Error(codes.InvalidArgument, "keyword is required")
	}

	// Use LIKE for keyword matching (FTS5 fallback)
	likeKeyword := "%" + kq.Keyword + "%"

	// Build WHERE clause based on field scope
	var whereClause string
	switch kq.FieldScope {
	case indexallv1.KeywordQuery_ALL:
		whereClause = "(r.title LIKE ? OR r.description LIKE ?)"
	case indexallv1.KeywordQuery_TITLE:
		whereClause = "r.title LIKE ?"
	case indexallv1.KeywordQuery_DESCRIPTION:
		whereClause = "r.description LIKE ?"
	default:
		return nil, 0, status.Error(codes.InvalidArgument, "invalid field_scope")
	}

	// Build tag scope recursion (without WHERE keyword)
	var tagScopeClause string
	switch kq.TagScope {
	case indexallv1.KeywordQuery_DIRECT:
		tagScopeClause = `rt.tag_id IN (
			SELECT DISTINCT rt2.tag_id FROM resource_tags rt2
		)`
	case indexallv1.KeywordQuery_WITH_ANCESTORS:
		tagScopeClause = `rt.tag_id IN (
			WITH RECURSIVE ancestors AS (
				SELECT t.id FROM tags t
				JOIN resource_tags rt2 ON t.id = rt2.tag_id
				WHERE rt2.resource_id IN (
					SELECT r.id FROM resources r WHERE ` + whereClause + `
				)
				UNION ALL
				SELECT tr.parent_id FROM tag_relations tr
				JOIN ancestors a ON tr.child_id = a.id
			)
			SELECT id FROM ancestors
		)`
	case indexallv1.KeywordQuery_WITH_DESCENDANTS:
		tagScopeClause = `rt.tag_id IN (
			WITH RECURSIVE descendants AS (
				SELECT t.id FROM tags t
				JOIN resource_tags rt2 ON t.id = rt2.tag_id
				WHERE rt2.resource_id IN (
					SELECT r.id FROM resources r WHERE ` + whereClause + `
				)
				UNION ALL
				SELECT tr.child_id FROM tag_relations tr
				JOIN descendants d ON tr.parent_id = d.id
			)
			SELECT id FROM descendants
		)`
	default:
		return nil, 0, status.Error(codes.InvalidArgument, "invalid tag_scope")
	}

	// Query count
	countQuery := `SELECT COUNT(DISTINCT r.id) FROM resources r
		JOIN resource_tags rt ON r.id = rt.resource_id
		WHERE ` + whereClause + ` AND ` + tagScopeClause

	var total int64
	var countErr error
	if kq.FieldScope == indexallv1.KeywordQuery_ALL {
		// ALL: need two parameters
		if kq.TagScope == indexallv1.KeywordQuery_DIRECT {
			countErr = s.db.QueryRowContext(ctx, countQuery, likeKeyword, likeKeyword).Scan(&total)
		} else {
			countErr = s.db.QueryRowContext(ctx, countQuery, likeKeyword, likeKeyword, likeKeyword, likeKeyword).Scan(&total)
		}
	} else {
		// TITLE or DESCRIPTION: single parameter
		if kq.TagScope == indexallv1.KeywordQuery_DIRECT {
			countErr = s.db.QueryRowContext(ctx, countQuery, likeKeyword).Scan(&total)
		} else {
			countErr = s.db.QueryRowContext(ctx, countQuery, likeKeyword, likeKeyword).Scan(&total)
		}
	}
	if countErr != nil && countErr != sql.ErrNoRows {
		return nil, 0, status.Errorf(codes.Internal, "failed to count resources: %v", countErr)
	}

	// Query resources
	dataQuery := `SELECT DISTINCT r.id, r.source, r.external_id, r.title, r.description, r.url, r.open_with, r.metadata, r.status, r.synced_at, r.created_at, r.updated_at
		FROM resources r
		JOIN resource_tags rt ON r.id = rt.resource_id
		WHERE ` + whereClause + ` AND ` + tagScopeClause + `
		ORDER BY r.created_at DESC
		LIMIT ? OFFSET ?`

	// Debug: uncomment to see generated SQL
	// fmt.Printf("CountQuery: %s\n", countQuery)
	// fmt.Printf("DataQuery: %s\n", dataQuery)

	var rows *sql.Rows
	var queryErr error
	if kq.FieldScope == indexallv1.KeywordQuery_ALL {
		// ALL: need two parameters + LIMIT/OFFSET
		if kq.TagScope == indexallv1.KeywordQuery_DIRECT {
			rows, queryErr = s.db.QueryContext(ctx, dataQuery, likeKeyword, likeKeyword, limit, offset)
		} else {
			rows, queryErr = s.db.QueryContext(ctx, dataQuery, likeKeyword, likeKeyword, likeKeyword, likeKeyword, limit, offset)
		}
	} else {
		// TITLE or DESCRIPTION: single parameter + LIMIT/OFFSET
		if kq.TagScope == indexallv1.KeywordQuery_DIRECT {
			rows, queryErr = s.db.QueryContext(ctx, dataQuery, likeKeyword, limit, offset)
		} else {
			rows, queryErr = s.db.QueryContext(ctx, dataQuery, likeKeyword, likeKeyword, limit, offset)
		}
	}
	if queryErr != nil {
		return nil, 0, status.Errorf(codes.Internal, "failed to query resources: %v", queryErr)
	}
	defer rows.Close()

	var resources []gen.Resource
	for rows.Next() {
		var r gen.Resource
		scanErr := rows.Scan(&r.ID, &r.Source, &r.ExternalID, &r.Title, &r.Description, &r.Url, &r.OpenWith, &r.Metadata, &r.Status, &r.SyncedAt, &r.CreatedAt, &r.UpdatedAt)
		if scanErr != nil {
			return nil, 0, status.Errorf(codes.Internal, "failed to scan resource: %v", scanErr)
		}
		resources = append(resources, r)
	}

	return resources, total, nil
}
