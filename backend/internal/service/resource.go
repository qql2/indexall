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

func (s *ResourceService) List(ctx context.Context, req *indexallv1.ListResourcesRequest) (*indexallv1.ListResourcesResponse, error) {
	var statusFilter sql.NullString
	if req.Status != nil && *req.Status != indexallv1.ResourceStatus_RESOURCE_STATUS_UNSPECIFIED {
		// Map proto enum to database string value
		statusFilter = sql.NullString{String: req.Status.String(), Valid: true}
	}

	resources, err := s.q.ListResources(ctx, gen.ListResourcesParams{
		Status: statusFilter,
		Limit:  int64(req.PageSize),
		Offset: int64((req.Page - 1) * req.PageSize),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list resources: %v", err)
	}

	// Get total count
	count, _ := s.q.CountResources(ctx, statusFilter)

	items := make([]*indexallv1.ResourceListItem, len(resources))
	for i, resource := range resources {
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

		items[i] = &indexallv1.ResourceListItem{
			Id:          resource.ID,
			Source:      resource.Source,
			Title:       resource.Title,
			Description: nullStringToPointer(resource.Description),
			Url:         nullStringToPointer(resource.Url),
			Status:      indexallv1.ResourceStatus_RESOURCE_STATUS_ACTIVE,
			CreatedAt:   nullTimeToString(resource.CreatedAt),
			Tags:        tags,
		}
	}

	return &indexallv1.ListResourcesResponse{
		Items:    items,
		Total:    int32(count),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *ResourceService) Search(ctx context.Context, req *indexallv1.SearchResourcesRequest) (*indexallv1.SearchResourcesResponse, error) {
	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}

	// TODO: Implement FTS5 search
	// For now, use simple LIKE search
	query := "%" + req.Query + "%"

	resources, err := s.q.ListResources(ctx, gen.ListResourcesParams{
		Status: sql.NullString{Valid: false},
		Limit:  int64(req.PageSize),
		Offset: int64((req.Page - 1) * req.PageSize),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to search resources: %v", err)
	}

	_ = query // TODO: use query in search

	items := make([]*indexallv1.ResourceSearchResult, len(resources))
	for i, resource := range resources {
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

		items[i] = &indexallv1.ResourceSearchResult{
			Id:          resource.ID,
			Source:      resource.Source,
			Title:       resource.Title,
			Description: nullStringToPointer(resource.Description),
			Url:         nullStringToPointer(resource.Url),
			CreatedAt:   nullTimeToString(resource.CreatedAt),
			Tags:        tags,
			MatchSource: indexallv1.MatchSource_MATCH_SOURCE_UNSPECIFIED,
		}
	}

	count, _ := s.q.CountResources(ctx, sql.NullString{Valid: false})

	return &indexallv1.SearchResourcesResponse{
		Items:    items,
		Total:    int32(count),
		Page:     req.Page,
		PageSize: req.PageSize,
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
