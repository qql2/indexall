package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	indexallv1 "github.com/construct/indexall/internal/gen/pb/indexall/v1"
	"github.com/construct/indexall/internal/db/gen"
)

type ResourceService struct {
	db *sql.DB
	q  *gen.Queries
}

func NewResourceService(db *sql.DB, q *gen.Queries) *ResourceService {
	return &ResourceService{db: db, q: q}
}

func (s *ResourceService) Create(ctx context.Context, req *connect.Request[indexallv1.CreateResourceRequest]) (*connect.Response[indexallv1.CreateResourceResponse], error) {
	if req.Msg.Title == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("title is required"))
	}

	// Dereference pointer fields with defaults
	source := "manual"
	if req.Msg.Source != nil && *req.Msg.Source != "" {
		source = *req.Msg.Source
	}

	var externalID sql.NullString
	if req.Msg.ExternalId != nil && *req.Msg.ExternalId != "" {
		externalID = sql.NullString{String: *req.Msg.ExternalId, Valid: true}
	}

	// Check uniqueness of source + externalId if both provided
	if source != "manual" && externalID.Valid {
		_, err := s.q.GetResourceBySourceAndExternalId(ctx, gen.GetResourceBySourceAndExternalIdParams{
			Source:     source,
			ExternalID: externalID,
		})
		if err == nil {
			return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("resource with source %q and external_id %q already exists", source, externalID.String))
		} else if err != sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	resourceID := uuid.New().String()
	resource, err := s.q.CreateResource(ctx, gen.CreateResourceParams{
		ID:          resourceID,
		Source:      source,
		ExternalID:  externalID,
		Title:       req.Msg.Title,
		Description: pointerToNullString(req.Msg.Description),
		Url:         pointerToNullString(req.Msg.Url),
		OpenWith:    pointerToNullString(req.Msg.OpenWith),
		Metadata:    sql.NullString{Valid: false},
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create resource: %w", err))
	}

	// Add tags
	for _, tagID := range req.Msg.TagIds {
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

	return connect.NewResponse(resp), nil
}

func (s *ResourceService) Update(ctx context.Context, req *connect.Request[indexallv1.UpdateResourceRequest]) (*connect.Response[indexallv1.UpdateResourceResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("resource id is required"))
	}

	// Verify resource exists
	existingResource, err := s.q.GetResource(ctx, req.Msg.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("resource not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Use existing values if not provided
	updateTitle := existingResource.Title
	if req.Msg.Title != nil && *req.Msg.Title != "" {
		updateTitle = *req.Msg.Title
	}

	err = s.q.UpdateResource(ctx, gen.UpdateResourceParams{
		ID:          req.Msg.Id,
		Title:       updateTitle,
		Description: pointerToNullString(req.Msg.Description),
		Url:         pointerToNullString(req.Msg.Url),
		OpenWith:    pointerToNullString(req.Msg.OpenWith),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update resource: %w", err))
	}

	return connect.NewResponse(&indexallv1.UpdateResourceResponse{
		Success: true,
	}), nil
}

func (s *ResourceService) Delete(ctx context.Context, req *connect.Request[indexallv1.DeleteResourceRequest]) (*connect.Response[indexallv1.DeleteResourceResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("resource id is required"))
	}

	// Verify resource exists
	if _, err := s.q.GetResource(ctx, req.Msg.Id); err != nil {
		if err == sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("resource not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Remove all tags
	_ = s.q.RemoveAllTagsFromResource(ctx, req.Msg.Id)

	// Delete the resource
	err := s.q.DeleteResource(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete resource: %w", err))
	}

	return connect.NewResponse(&indexallv1.DeleteResourceResponse{
		Success: true,
	}), nil
}

func (s *ResourceService) Get(ctx context.Context, req *connect.Request[indexallv1.GetResourceRequest]) (*connect.Response[indexallv1.GetResourceResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("resource id is required"))
	}

	resource, err := s.q.GetResource(ctx, req.Msg.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("resource not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Get tags
	tags := make([]*indexallv1.ResourceTag, 0)
	tagRows, err := s.q.GetResourceTags(ctx, req.Msg.Id)
	if err == nil {
		for _, t := range tagRows {
			tags = append(tags, &indexallv1.ResourceTag{
				Id:    t.TagID,
				Name:  t.Name,
				Color: nullStringToPointer(t.Color),
			})
		}
	}

	syncedAtStr := nullTimeToString(resource.SyncedAt)
	var syncedAtPtr *string
	if syncedAtStr != "" {
		syncedAtPtr = &syncedAtStr
	}

	resp := &indexallv1.GetResourceResponse{
		Id:          resource.ID,
		Source:      resource.Source,
		ExternalId:  nullStringToPointer(resource.ExternalID),
		Title:       resource.Title,
		Description: nullStringToPointer(resource.Description),
		Url:         nullStringToPointer(resource.Url),
		OpenWith:    nullStringToPointer(resource.OpenWith),
		Metadata:    nullStringToPointer(resource.Metadata),
		Status:      statusToProto(resource.Status.String),
		SyncedAt:    syncedAtPtr,
		CreatedAt:   nullTimeToString(resource.CreatedAt),
		UpdatedAt:   nullTimeToString(resource.UpdatedAt),
		Tags:        tags,
	}

	return connect.NewResponse(resp), nil
}

func (s *ResourceService) List(ctx context.Context, req *connect.Request[indexallv1.ListResourcesRequest]) (*connect.Response[indexallv1.ListResourcesResponse], error) {
	page := req.Msg.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.Msg.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var resources []gen.Resource
	var total int64
	var err error

	status := "active"
	if req.Msg.Status != nil && *req.Msg.Status != indexallv1.ResourceStatus_RESOURCE_STATUS_UNSPECIFIED {
		status = statusToString(*req.Msg.Status)
	}

	tagID := ""
	if req.Msg.TagId != nil {
		tagID = *req.Msg.TagId
	}

	if tagID != "" {
		// Get by tag
		resources, err = s.q.ListResourcesByTag(ctx, gen.ListResourcesByTagParams{
			TagID:  tagID,
			Status: stringToNullString(status),
			Limit:  int64(pageSize),
			Offset: int64((page - 1) * pageSize),
		})
		if err != nil {
			resources = []gen.Resource{}
		}

		// Count
		countRes, err := s.q.CountResourcesByTag(ctx, gen.CountResourcesByTagParams{
			TagID:  tagID,
			Status: stringToNullString(status),
		})
		if err == nil {
			total = countRes
		}
	} else {
		// List all
		resources, err = s.q.ListResources(ctx, gen.ListResourcesParams{
			Status: stringToNullString(status),
			Limit:  int64(pageSize),
			Offset: int64((page - 1) * pageSize),
		})
		if err != nil {
			resources = []gen.Resource{}
		}

		// Count
		countRes, err := s.q.CountResources(ctx, stringToNullString(status))
		if err == nil {
			total = countRes
		}
	}

	// Build response
	items := make([]*indexallv1.ResourceListItem, 0)
	for _, r := range resources {
		tags := make([]*indexallv1.ResourceTag, 0)
		tagRows, _ := s.q.GetResourceTags(ctx, r.ID)
		for _, t := range tagRows {
			tags = append(tags, &indexallv1.ResourceTag{
				Id:    t.TagID,
				Name:  t.Name,
				Color: nullStringToPointer(t.Color),
			})
		}

		items = append(items, &indexallv1.ResourceListItem{
			Id:          r.ID,
			Source:      r.Source,
			Title:       r.Title,
			Description: nullStringToPointer(r.Description),
			Url:         nullStringToPointer(r.Url),
			Status:      statusToProto(r.Status.String),
			CreatedAt:   nullTimeToString(r.CreatedAt),
			Tags:        tags,
		})
	}

	return connect.NewResponse(&indexallv1.ListResourcesResponse{
		Items:    items,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}), nil
}

func (s *ResourceService) Search(ctx context.Context, req *connect.Request[indexallv1.SearchResourcesRequest]) (*connect.Response[indexallv1.SearchResourcesResponse], error) {
	if req.Msg.Query == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("search query is required"))
	}

	page := req.Msg.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.Msg.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := "%" + req.Msg.Query + "%"
	resources, err := s.q.SearchResources(ctx, gen.SearchResourcesParams{
		Title: query,
		Description: stringToNullString(query),
		Limit:   int64(pageSize),
		Offset:  int64((page - 1) * pageSize),
	})
	if err != nil {
		resources = []gen.SearchResourcesRow{}
	}

	// Count
	countRes, err := s.q.CountSearchResults(ctx, gen.CountSearchResultsParams{
		Title: query,
		Description: stringToNullString(query),
	})
	var total int64
	if err == nil {
		total = countRes
	}

	// Build response
	items := make([]*indexallv1.ResourceSearchResult, 0)
	for _, r := range resources {
		tags := make([]*indexallv1.ResourceTag, 0)
		tagRows, _ := s.q.GetResourceTags(ctx, r.ID)
		for _, t := range tagRows {
			tags = append(tags, &indexallv1.ResourceTag{
				Id:    t.TagID,
				Name:  t.Name,
				Color: nullStringToPointer(t.Color),
			})
		}

		items = append(items, &indexallv1.ResourceSearchResult{
			Id:          r.ID,
			Source:      r.Source,
			Title:       r.Title,
			Description: nullStringToPointer(r.Description),
			Url:         nullStringToPointer(r.Url),
			CreatedAt:   nullTimeToString(r.CreatedAt),
			Tags:        tags,
			MatchSource: indexallv1.MatchSource_MATCH_SOURCE_TITLE, // Always title for now
		})
	}

	return connect.NewResponse(&indexallv1.SearchResourcesResponse{
		Items:    items,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}), nil
}

func (s *ResourceService) GetByUrl(ctx context.Context, req *connect.Request[indexallv1.GetByUrlRequest]) (*connect.Response[indexallv1.GetByUrlResponse], error) {
	if req.Msg.Url == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("url is required"))
	}

	// Normalize URL
	normalizedURL := normalizeURL(req.Msg.Url)

	resource, err := s.q.GetResourceByUrl(ctx, stringToNullString(normalizedURL))
	if err != nil {
		if err == sql.ErrNoRows {
			return connect.NewResponse(&indexallv1.GetByUrlResponse{}), nil
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Get tags
	tags := make([]*indexallv1.ResourceTag, 0)
	tagRows, err := s.q.GetResourceTags(ctx, resource.ID)
	if err == nil {
		for _, t := range tagRows {
			tags = append(tags, &indexallv1.ResourceTag{
				Id:    t.TagID,
				Name:  t.Name,
				Color: nullStringToPointer(t.Color),
			})
		}
	}

	return connect.NewResponse(&indexallv1.GetByUrlResponse{
		Resource: &indexallv1.GetByUrlResponse_Resource{
			Id:    resource.ID,
			Title: resource.Title,
			Tags:  tags,
		},
	}), nil
}

func (s *ResourceService) AddTag(ctx context.Context, req *connect.Request[indexallv1.AddTagRequest]) (*connect.Response[indexallv1.AddTagResponse], error) {
	if req.Msg.ResourceId == "" || req.Msg.TagId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("resource id and tag id are required"))
	}

	err := s.q.AddTagToResource(ctx, gen.AddTagToResourceParams{
		ResourceID: req.Msg.ResourceId,
		TagID:      req.Msg.TagId,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to add tag: %w", err))
	}

	return connect.NewResponse(&indexallv1.AddTagResponse{
		Success: true,
	}), nil
}

func (s *ResourceService) RemoveTag(ctx context.Context, req *connect.Request[indexallv1.RemoveTagRequest]) (*connect.Response[indexallv1.RemoveTagResponse], error) {
	if req.Msg.ResourceId == "" || req.Msg.TagId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("resource id and tag id are required"))
	}

	err := s.q.RemoveTagFromResource(ctx, gen.RemoveTagFromResourceParams{
		ResourceID: req.Msg.ResourceId,
		TagID:      req.Msg.TagId,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to remove tag: %w", err))
	}

	return connect.NewResponse(&indexallv1.RemoveTagResponse{
		Success: true,
	}), nil
}

// Helper functions
func statusToProto(status string) indexallv1.ResourceStatus {
	switch status {
	case "active":
		return indexallv1.ResourceStatus_RESOURCE_STATUS_ACTIVE
	case "stale":
		return indexallv1.ResourceStatus_RESOURCE_STATUS_STALE
	case "deleted":
		return indexallv1.ResourceStatus_RESOURCE_STATUS_DELETED
	default:
		return indexallv1.ResourceStatus_RESOURCE_STATUS_UNSPECIFIED
	}
}

func statusToString(status indexallv1.ResourceStatus) string {
	switch status {
	case indexallv1.ResourceStatus_RESOURCE_STATUS_ACTIVE:
		return "active"
	case indexallv1.ResourceStatus_RESOURCE_STATUS_STALE:
		return "stale"
	case indexallv1.ResourceStatus_RESOURCE_STATUS_DELETED:
		return "deleted"
	default:
		return "active"
	}
}

func normalizeURL(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return u
	}
	return parsed.String()
}
