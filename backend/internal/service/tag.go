package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	indexallv1 "github.com/construct/indexall/internal/gen/pb/indexall/v1"
	"github.com/construct/indexall/internal/db/gen" // Import directly without alias
)

type TagService struct {
	db *sql.DB
	q  *gen.Queries
}

func NewTagService(db *sql.DB, q *gen.Queries) *TagService {
	return &TagService{db: db, q: q}
}

func (s *TagService) Create(ctx context.Context, req *connect.Request[indexallv1.CreateTagRequest]) (*connect.Response[indexallv1.CreateTagResponse], error) {
	if req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("tag name is required"))
	}

	// Check if name already exists
	_, err := s.q.GetTagByName(ctx, req.Msg.Name)
	if err == nil {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("tag name %q already exists", req.Msg.Name))
	} else if err != sql.ErrNoRows {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Check aliases uniqueness
	for _, alias := range req.Msg.Aliases {
		_, err := s.q.GetAliasByName(ctx, alias)
		if err == nil {
			return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("alias %q already exists", alias))
		} else if err != sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	tagID := uuid.New().String()
	tag, err := s.q.CreateTag(ctx, gen.CreateTagParams{
		ID:    tagID,
		Name:  req.Msg.Name,
		Color: nilIfEmpty(req.Msg.Color),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create tag: %w", err))
	}

	// Create aliases
	for _, alias := range req.Msg.Aliases {
		_, err := s.q.CreateAlias(ctx, gen.CreateAliasParams{
			ID:    uuid.New().String(),
			TagID: tagID,
			Alias: alias,
		})
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create alias: %w", err))
		}
	}

	// Add parent relations
	for _, parentID := range req.Msg.ParentIds {
		err := s.q.CreateTagRelation(ctx, gen.CreateTagRelationParams{
			ParentID: parentID,
			ChildID:  tagID,
		})
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to add parent relation: %w", err))
		}
	}

	resp := &indexallv1.CreateTagResponse{
		Id:        tag.ID,
		Name:      tag.Name,
		Color:     nullStringToPointer(tag.Color),
		Aliases:   req.Msg.Aliases,
		ParentIds: req.Msg.ParentIds,
		CreatedAt: nullTimeToString(tag.CreatedAt),
	}

	return connect.NewResponse(resp), nil
}

func (s *TagService) Update(ctx context.Context, req *connect.Request[indexallv1.UpdateTagRequest]) (*connect.Response[indexallv1.UpdateTagResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("tag id is required"))
	}

	// Verify tag exists
	if _, err := s.q.GetTag(ctx, req.Msg.Id); err != nil {
		if err == sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("tag not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Check if new name is unique (if provided)
	if req.Msg.Name != nil && *req.Msg.Name != "" {
		existing, err := s.q.GetTagByName(ctx, *req.Msg.Name)
		if err == nil && existing.ID != req.Msg.Id {
			return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("tag name %q already exists", *req.Msg.Name))
		} else if err != nil && err != sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	// Get existing tag to preserve unchanged fields
	existingTag, _ := s.q.GetTag(ctx, req.Msg.Id)

	updateName := existingTag.Name
	if req.Msg.Name != nil {
		updateName = *req.Msg.Name
	}

	err := s.q.UpdateTag(ctx, gen.UpdateTagParams{
		ID:    req.Msg.Id,
		Name:  updateName,
		Color: nilIfEmpty(req.Msg.Color),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update tag: %w", err))
	}

	return connect.NewResponse(&indexallv1.UpdateTagResponse{
		Success: true,
	}), nil
}

func (s *TagService) Delete(ctx context.Context, req *connect.Request[indexallv1.DeleteTagRequest]) (*connect.Response[indexallv1.DeleteTagResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("tag id is required"))
	}

	// Verify tag exists
	if _, err := s.q.GetTag(ctx, req.Msg.Id); err != nil {
		if err == sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("tag not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Delete aliases
	_ = s.q.DeleteAliasByTagId(ctx, req.Msg.Id)

	// Delete relations
	_ = s.q.DeleteTagRelationsByParent(ctx, req.Msg.Id)
	_ = s.q.DeleteTagRelationsByChild(ctx, req.Msg.Id)

	// Delete the tag (cascades to resource_tags)
	err := s.q.DeleteTag(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete tag: %w", err))
	}

	return connect.NewResponse(&indexallv1.DeleteTagResponse{
		Success: true,
	}), nil
}

func (s *TagService) List(ctx context.Context, req *connect.Request[indexallv1.ListTagsRequest]) (*connect.Response[indexallv1.ListTagsResponse], error) {
	tags, err := s.q.GetTagsWithCounts(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list tags: %w", err))
	}

	var items []*indexallv1.TagListItem
	for _, tag := range tags {
		aliases, _ := s.q.ListAliasesByTag(ctx, tag.ID)
		parents, _ := s.q.ListParentTags(ctx, tag.ID)

		aliasStrs := make([]string, len(aliases))
		for i, a := range aliases {
			aliasStrs[i] = a.Alias
		}

		var parentStrs []string
		for _, p := range parents {
			parentStrs = append(parentStrs, p)
		}

		items = append(items, &indexallv1.TagListItem{
			Id:            tag.ID,
			Name:          tag.Name,
			Color:         nullStringToPointer(tag.Color),
			Aliases:       aliasStrs,
			ParentIds:     parentStrs,
			ResourceCount: int32(tag.ResourceCount),
		})
	}

	return connect.NewResponse(&indexallv1.ListTagsResponse{
		Tags: items,
	}), nil
}

func (s *TagService) GetTree(ctx context.Context, req *connect.Request[indexallv1.GetTreeRequest]) (*connect.Response[indexallv1.GetTreeResponse], error) {
	// Get root tags
	roots, err := s.q.GetTagTree(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get tag tree: %w", err))
	}

	nodes := make([]*indexallv1.TagTreeNode, 0)
	for _, root := range roots {
		node, err := s.buildTagTreeNode(ctx, root.ID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		nodes = append(nodes, node)
	}

	return connect.NewResponse(&indexallv1.GetTreeResponse{
		Roots: nodes,
	}), nil
}

func (s *TagService) buildTagTreeNode(ctx context.Context, tagID string) (*indexallv1.TagTreeNode, error) {
	tag, err := s.q.GetTag(ctx, tagID)
	if err != nil {
		return nil, err
	}

	count, err := s.q.CountResourcesForTag(ctx, tagID)
	if err != nil {
		count = 0
	}

	children, err := s.q.GetTagTreeNode(ctx, tagID)
	if err != nil {
		children = []gen.GetTagTreeNodeRow{}
	}

	childNodes := make([]*indexallv1.TagTreeNode, 0)
	for _, child := range children {
		childNode, err := s.buildTagTreeNode(ctx, child.ID)
		if err == nil {
			childNodes = append(childNodes, childNode)
		}
	}

	resCount := int32(0)
	if count > 0 {
		resCount = int32(count)
	}

	return &indexallv1.TagTreeNode{
		Id:            tag.ID,
		Name:          tag.Name,
		Color:         nullStringToPointer(tag.Color),
		ResourceCount: resCount,
		Children:      childNodes,
	}, nil
}

func (s *TagService) Search(ctx context.Context, req *connect.Request[indexallv1.SearchTagsRequest]) (*connect.Response[indexallv1.SearchTagsResponse], error) {
	if req.Msg.Query == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("search query is required"))
	}

	query := "%" + req.Msg.Query + "%"
	results, err := s.q.SearchTags(ctx, gen.SearchTagsParams{
		Name:   query,
		Name_2: query,
		Alias:  query,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to search tags: %w", err))
	}

	items := make([]*indexallv1.TagSearchResult, 0)
	for _, result := range results {
		var matchedAlias *string
		if result.MatchedAlias != nil {
			if alias, ok := result.MatchedAlias.(string); ok {
				matchedAlias = &alias
			}
		}
		items = append(items, &indexallv1.TagSearchResult{
			Id:            result.ID,
			Name:          result.Name,
			Color:         nullStringToPointer(result.Color),
			MatchedAlias:  matchedAlias,
		})
	}

	return connect.NewResponse(&indexallv1.SearchTagsResponse{
		Results: items,
	}), nil
}

func (s *TagService) AddAlias(ctx context.Context, req *connect.Request[indexallv1.AddAliasRequest]) (*connect.Response[indexallv1.AddAliasResponse], error) {
	if req.Msg.TagId == "" || req.Msg.Alias == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("tag id and alias are required"))
	}

	// Check alias uniqueness
	_, err := s.q.GetAliasByName(ctx, req.Msg.Alias)
	if err == nil {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("alias %q already exists", req.Msg.Alias))
	} else if err != sql.ErrNoRows {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	aliasID := uuid.New().String()
	alias, err := s.q.CreateAlias(ctx, gen.CreateAliasParams{
		ID:    aliasID,
		TagID: req.Msg.TagId,
		Alias: req.Msg.Alias,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to add alias: %w", err))
	}

	return connect.NewResponse(&indexallv1.AddAliasResponse{
		Id:    alias.ID,
		Alias: alias.Alias,
	}), nil
}

func (s *TagService) RemoveAlias(ctx context.Context, req *connect.Request[indexallv1.RemoveAliasRequest]) (*connect.Response[indexallv1.RemoveAliasResponse], error) {
	if req.Msg.AliasId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("alias id is required"))
	}

	err := s.q.DeleteAlias(ctx, req.Msg.AliasId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to remove alias: %w", err))
	}

	return connect.NewResponse(&indexallv1.RemoveAliasResponse{
		Success: true,
	}), nil
}

func (s *TagService) AddParent(ctx context.Context, req *connect.Request[indexallv1.AddParentRequest]) (*connect.Response[indexallv1.AddParentResponse], error) {
	if req.Msg.ChildId == "" || req.Msg.ParentId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("child id and parent id are required"))
	}

	if req.Msg.ChildId == req.Msg.ParentId {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("cannot add self-reference"))
	}

	// Check for cycles
	wouldCreateCycle, err := s.q.CheckCycleWouldExist(ctx, gen.CheckCycleWouldExistParams{
		ParentID: req.Msg.ParentId,
		ChildID:  req.Msg.ChildId,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if wouldCreateCycle > 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("adding this relation would create a cycle"))
	}

	err = s.q.CreateTagRelation(ctx, gen.CreateTagRelationParams{
		ParentID: req.Msg.ParentId,
		ChildID:  req.Msg.ChildId,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to add parent: %w", err))
	}

	return connect.NewResponse(&indexallv1.AddParentResponse{
		Success: true,
	}), nil
}

func (s *TagService) RemoveParent(ctx context.Context, req *connect.Request[indexallv1.RemoveParentRequest]) (*connect.Response[indexallv1.RemoveParentResponse], error) {
	if req.Msg.ChildId == "" || req.Msg.ParentId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("child id and parent id are required"))
	}

	err := s.q.DeleteTagRelation(ctx, gen.DeleteTagRelationParams{
		ParentID: req.Msg.ParentId,
		ChildID:  req.Msg.ChildId,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to remove parent: %w", err))
	}

	return connect.NewResponse(&indexallv1.RemoveParentResponse{
		Success: true,
	}), nil
}

// Helper function to convert pointer string to sql.NullString
func nilIfEmpty(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}
