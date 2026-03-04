package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	indexallv1 "github.com/construct/indexall/api/indexall/v1"
	"github.com/construct/indexall/internal/db/gen"
)

// Ensure TagService implements TagServiceServer interface
var _ indexallv1.TagServiceServer = (*TagService)(nil)

type TagService struct {
	indexallv1.UnimplementedTagServiceServer
	db *sql.DB
	q  *gen.Queries
}

func NewTagService(db *sql.DB, q *gen.Queries) *TagService {
	return &TagService{db: db, q: q}
}

func (s *TagService) Create(ctx context.Context, req *indexallv1.CreateTagRequest) (*indexallv1.CreateTagResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "tag name is required")
	}

	// Check if name already exists
	_, err := s.q.GetTagByName(ctx, req.Name)
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "tag name %q already exists", req.Name)
	} else if err != sql.ErrNoRows {
		return nil, status.Errorf(codes.Internal, "failed to check tag name: %v", err)
	}

	// Check aliases uniqueness
	for _, alias := range req.Aliases {
		_, err := s.q.GetAliasByName(ctx, alias)
		if err == nil {
			return nil, status.Errorf(codes.AlreadyExists, "alias %q already exists", alias)
		} else if err != sql.ErrNoRows {
			return nil, status.Errorf(codes.Internal, "failed to check alias: %v", err)
		}
	}

	tagID := uuid.New().String()
	tag, err := s.q.CreateTag(ctx, gen.CreateTagParams{
		ID:    tagID,
		Name:  req.Name,
		Color: nilIfEmpty(req.Color),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create tag: %v", err)
	}

	// Create aliases
	for _, alias := range req.Aliases {
		_, err := s.q.CreateAlias(ctx, gen.CreateAliasParams{
			ID:    uuid.New().String(),
			TagID: tagID,
			Alias: alias,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create alias: %v", err)
		}
	}

	// Add parent relations
	for _, parentID := range req.ParentIds {
		err := s.q.CreateTagRelation(ctx, gen.CreateTagRelationParams{
			ParentID: parentID,
			ChildID:  tagID,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to add parent relation: %v", err)
		}
	}

	resp := &indexallv1.CreateTagResponse{
		Id:        tag.ID,
		Name:      tag.Name,
		Color:     nullStringToPointer(tag.Color),
		Aliases:   req.Aliases,
		ParentIds: req.ParentIds,
		CreatedAt: nullTimeToString(tag.CreatedAt),
	}

	return resp, nil
}

func (s *TagService) Update(ctx context.Context, req *indexallv1.UpdateTagRequest) (*indexallv1.UpdateTagResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "tag id is required")
	}

	// Verify tag exists
	if _, err := s.q.GetTag(ctx, req.Id); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "tag not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get tag: %v", err)
	}

	// Check if new name is unique (if provided)
	if req.Name != nil && *req.Name != "" {
		existing, err := s.q.GetTagByName(ctx, *req.Name)
		if err == nil && existing.ID != req.Id {
			return nil, status.Errorf(codes.AlreadyExists, "tag name %q already exists", *req.Name)
		} else if err != nil && err != sql.ErrNoRows {
			return nil, status.Errorf(codes.Internal, "failed to check tag name: %v", err)
		}
	}

	// Get existing tag to preserve unchanged fields
	existingTag, _ := s.q.GetTag(ctx, req.Id)

	updateName := existingTag.Name
	if req.Name != nil {
		updateName = *req.Name
	}

	err := s.q.UpdateTag(ctx, gen.UpdateTagParams{
		ID:    req.Id,
		Name:  updateName,
		Color: nilIfEmpty(req.Color),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update tag: %v", err)
	}

	return &indexallv1.UpdateTagResponse{
		Success: true,
	}, nil
}

func (s *TagService) Delete(ctx context.Context, req *indexallv1.DeleteTagRequest) (*indexallv1.DeleteTagResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "tag id is required")
	}

	// Verify tag exists
	if _, err := s.q.GetTag(ctx, req.Id); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "tag not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get tag: %v", err)
	}

	err := s.q.DeleteTag(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete tag: %v", err)
	}

	return &indexallv1.DeleteTagResponse{
		Success: true,
	}, nil
}

func (s *TagService) List(ctx context.Context, req *indexallv1.ListTagsRequest) (*indexallv1.ListTagsResponse, error) {
	tags, err := s.q.ListTags(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tags: %v", err)
	}

	items := make([]*indexallv1.TagListItem, len(tags))
	for i, tag := range tags {
		items[i] = &indexallv1.TagListItem{
			Id:            tag.ID,
			Name:          tag.Name,
			Color:         nullStringToPointer(tag.Color),
			Aliases:       getTagAliases(ctx, s.q, tag.ID),
			ParentIds:     getTagParents(ctx, s.q, tag.ID),
			ResourceCount: getTagResourceCount(ctx, s.q, tag.ID),
		}
	}

	return &indexallv1.ListTagsResponse{
		Tags: items,
	}, nil
}

func (s *TagService) GetTree(ctx context.Context, req *indexallv1.GetTreeRequest) (*indexallv1.GetTreeResponse, error) {
	tags, err := s.q.ListTags(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get tags: %v", err)
	}

	// Find root tags (tags without parents)
	roots := make([]*indexallv1.TagTreeNode, 0)
	tagMap := make(map[string]*indexallv1.TagTreeNode)

	for _, tag := range tags {
		node := &indexallv1.TagTreeNode{
			Id:              tag.ID,
			Name:            tag.Name,
			Color:           nullStringToPointer(tag.Color),
			ResourceCount:   getTagResourceCount(ctx, s.q, tag.ID),
			Children:        make([]*indexallv1.TagTreeNode, 0),
		}
		tagMap[tag.ID] = node
	}

	// Build parent-child relationships
	for _, tag := range tags {
		parents := getTagParents(ctx, s.q, tag.ID)
		if len(parents) == 0 {
			roots = append(roots, tagMap[tag.ID])
		} else {
			for _, parentID := range parents {
				if parent, ok := tagMap[parentID]; ok {
					parent.Children = append(parent.Children, tagMap[tag.ID])
				}
			}
		}
	}

	return &indexallv1.GetTreeResponse{
		Roots: roots,
	}, nil
}

func (s *TagService) Search(ctx context.Context, req *indexallv1.SearchTagsRequest) (*indexallv1.SearchTagsResponse, error) {
	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}

	// Set defaults
	limit := req.Limit
	if limit < 1 {
		limit = 20
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	// Use LIKE for keyword matching (FTS5 fallback)
	likeQuery := "%" + req.Query + "%"

	// Build tag scope clause (search in tag name and aliases)
	var tagScopeClause string
	switch req.TagScope {
	case indexallv1.SearchTagsRequest_DIRECT:
		tagScopeClause = ""
	case indexallv1.SearchTagsRequest_WITH_ANCESTORS:
		tagScopeClause = `WITH RECURSIVE ancestors AS (
			SELECT t.id FROM tags t
			WHERE t.name LIKE ? OR t.id IN (
				SELECT tag_id FROM tag_aliases WHERE alias LIKE ?
			)
			UNION ALL
			SELECT tr.parent_id FROM tag_relations tr
			JOIN ancestors a ON tr.child_id = a.id
		) SELECT DISTINCT id FROM ancestors`
	case indexallv1.SearchTagsRequest_WITH_DESCENDANTS:
		tagScopeClause = `WITH RECURSIVE descendants AS (
			SELECT t.id FROM tags t
			WHERE t.name LIKE ? OR t.id IN (
				SELECT tag_id FROM tag_aliases WHERE alias LIKE ?
			)
			UNION ALL
			SELECT tr.child_id FROM tag_relations tr
			JOIN descendants d ON tr.parent_id = d.id
		) SELECT DISTINCT id FROM descendants`
	default:
		tagScopeClause = ""
	}

	// Query count
	var countQuery string
	var countErr error
	var total int64

	if req.TagScope == indexallv1.SearchTagsRequest_DIRECT {
		countQuery = `SELECT COUNT(DISTINCT t.id) FROM tags t
			LEFT JOIN tag_aliases ta ON t.id = ta.tag_id
			WHERE t.name LIKE ? OR ta.alias LIKE ?`
		countErr = s.db.QueryRowContext(ctx, countQuery, likeQuery, likeQuery).Scan(&total)
	} else {
		countQuery = `SELECT COUNT(*) FROM (` + tagScopeClause + `)`
		countErr = s.db.QueryRowContext(ctx, countQuery, likeQuery, likeQuery).Scan(&total)
	}

	if countErr != nil && countErr != sql.ErrNoRows {
		return nil, status.Errorf(codes.Internal, "failed to count tags: %v", countErr)
	}

	// Query tags
	var dataQuery string
	var rows *sql.Rows
	var queryErr error

	if req.TagScope == indexallv1.SearchTagsRequest_DIRECT {
		dataQuery = `SELECT DISTINCT t.id, t.name, t.color, t.created_at
			FROM tags t
			LEFT JOIN tag_aliases ta ON t.id = ta.tag_id
			WHERE t.name LIKE ? OR ta.alias LIKE ?
			ORDER BY t.name ASC
			LIMIT ? OFFSET ?`
		rows, queryErr = s.db.QueryContext(ctx, dataQuery, likeQuery, likeQuery, limit, offset)
	} else {
		dataQuery = `SELECT DISTINCT t.id, t.name, t.color, t.created_at FROM tags t
			WHERE t.id IN (` + tagScopeClause + `)
			ORDER BY t.name ASC
			LIMIT ? OFFSET ?`
		rows, queryErr = s.db.QueryContext(ctx, dataQuery, likeQuery, likeQuery, limit, offset)
	}

	if queryErr != nil {
		return nil, status.Errorf(codes.Internal, "failed to query tags: %v", queryErr)
	}
	defer rows.Close()

	results := make([]*indexallv1.TagSearchResult, 0)
	for rows.Next() {
		var tagID, name string
		var color sql.NullString
		var createdAt sql.NullTime

		if err := rows.Scan(&tagID, &name, &color, &createdAt); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan tag: %v", err)
		}

		// Get aliases
		aliases := make([]string, 0)
		aliasRows, err := s.q.ListAliasesByTag(ctx, tagID)
		if err == nil {
			for _, a := range aliasRows {
				aliases = append(aliases, a.Alias)
			}
		}

		// Get resource count
		resourceCount, _ := s.q.CountResourcesForTag(ctx, tagID)

		results = append(results, &indexallv1.TagSearchResult{
			Id:            tagID,
			Name:          name,
			Color:         nullStringToPointer(color),
			Description:   nil, // TODO: Add description field to tags table
			Aliases:       aliases,
			ResourceCount: int32(resourceCount),
		})
	}

	return &indexallv1.SearchTagsResponse{
		Results: results,
		Total:   int32(total),
	}, nil
}

func (s *TagService) AddAlias(ctx context.Context, req *indexallv1.AddAliasRequest) (*indexallv1.AddAliasResponse, error) {
	if req.TagId == "" || req.Alias == "" {
		return nil, status.Error(codes.InvalidArgument, "tag_id and alias are required")
	}

	// Verify tag exists
	if _, err := s.q.GetTag(ctx, req.TagId); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "tag not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get tag: %v", err)
	}

	// Check alias uniqueness
	_, err := s.q.GetAliasByName(ctx, req.Alias)
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "alias %q already exists", req.Alias)
	} else if err != sql.ErrNoRows {
		return nil, status.Errorf(codes.Internal, "failed to check alias: %v", err)
	}

	aliasID := uuid.New().String()
	_, err = s.q.CreateAlias(ctx, gen.CreateAliasParams{
		ID:    aliasID,
		TagID: req.TagId,
		Alias: req.Alias,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create alias: %v", err)
	}

	return &indexallv1.AddAliasResponse{
		Id:    aliasID,
		Alias: req.Alias,
	}, nil
}

func (s *TagService) RemoveAlias(ctx context.Context, req *indexallv1.RemoveAliasRequest) (*indexallv1.RemoveAliasResponse, error) {
	if req.AliasId == "" {
		return nil, status.Error(codes.InvalidArgument, "alias_id is required")
	}

	err := s.q.DeleteAlias(ctx, req.AliasId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete alias: %v", err)
	}

	return &indexallv1.RemoveAliasResponse{
		Success: true,
	}, nil
}

func (s *TagService) AddParent(ctx context.Context, req *indexallv1.AddParentRequest) (*indexallv1.AddParentResponse, error) {
	if req.ChildId == "" || req.ParentId == "" {
		return nil, status.Error(codes.InvalidArgument, "child_id and parent_id are required")
	}

	err := s.q.CreateTagRelation(ctx, gen.CreateTagRelationParams{
		ParentID: req.ParentId,
		ChildID:  req.ChildId,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add parent relation: %v", err)
	}

	return &indexallv1.AddParentResponse{
		Success: true,
	}, nil
}

func (s *TagService) RemoveParent(ctx context.Context, req *indexallv1.RemoveParentRequest) (*indexallv1.RemoveParentResponse, error) {
	if req.ChildId == "" || req.ParentId == "" {
		return nil, status.Error(codes.InvalidArgument, "child_id and parent_id are required")
	}

	err := s.q.DeleteTagRelation(ctx, gen.DeleteTagRelationParams{
		ParentID: req.ParentId,
		ChildID:  req.ChildId,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove parent relation: %v", err)
	}

	return &indexallv1.RemoveParentResponse{
		Success: true,
	}, nil
}
