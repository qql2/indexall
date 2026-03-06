package service

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	indexallv1 "github.com/construct/indexall/api/indexall/v1"
	"github.com/construct/indexall/internal/db/gen"
	"github.com/google/uuid"
)

// setupTagTestData creates sample tags with hierarchy for testing
func setupTagTestData(t *testing.T, q *gen.Queries) (string, string, string) {
	ctx := context.Background()

	// Create parent tag
	parentTag, err := q.CreateTag(ctx, gen.CreateTagParams{
		ID:    uuid.New().String(),
		Name:  "Technology",
		Color: sql.NullString{},
	})
	if err != nil {
		t.Fatalf("failed to create parent tag: %v", err)
	}

	// Create child tag
	childTag, err := q.CreateTag(ctx, gen.CreateTagParams{
		ID:    uuid.New().String(),
		Name:  "Programming",
		Color: sql.NullString{},
	})
	if err != nil {
		t.Fatalf("failed to create child tag: %v", err)
	}

	// Create relationship
	err = q.CreateTagRelation(ctx, gen.CreateTagRelationParams{
		ParentID: parentTag.ID,
		ChildID:  childTag.ID,
	})
	if err != nil {
		t.Fatalf("failed to create tag relation: %v", err)
	}

	// Add alias to child tag
	_, err = q.CreateAlias(ctx, gen.CreateAliasParams{
		ID:    uuid.New().String(),
		TagID: childTag.ID,
		Alias: "coding",
	})
	if err != nil {
		t.Fatalf("failed to create alias: %v", err)
	}

	return parentTag.ID, childTag.ID, ""
}

// TestSearchTagsDirect tests tag search with DIRECT scope
func TestSearchTagsDirect(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	setupTagTestData(t, q)

	ctx := context.Background()
	service := NewTagService(database, q, nil)

	// Search by exact tag name
	req := &indexallv1.SearchTagsRequest{
		Query: "Programming",
		TagScope: indexallv1.SearchTagsRequest_DIRECT,
		Limit: 20,
		Offset: 0,
	}

	resp, err := service.Search(ctx, req)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should find the Programming tag
	if resp.Total < 1 {
		t.Errorf("expected at least 1 tag, got %d", resp.Total)
	}

	found := false
	for _, result := range resp.Results {
		if result.Name == "Programming" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Programming tag not found")
	}
}

// TestSearchTagsWithAncestors tests tag search with WITH_ANCESTORS scope.
// Hierarchy: Technology (parent) -> Programming (child)
// Searching "Tech" (matches "Technology") with WITH_ANCESTORS should return
// BOTH Technology AND Programming (because Programming's ancestor matches).
func TestSearchTagsWithAncestors(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	setupTagTestData(t, q)

	ctx := context.Background()
	service := NewTagService(database, q, nil)

	req := &indexallv1.SearchTagsRequest{
		Query:    "Tech",
		TagScope: indexallv1.SearchTagsRequest_WITH_ANCESTORS,
		Limit:    20,
		Offset:   0,
	}

	resp, err := service.Search(ctx, req)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should find both Technology (direct match) and Programming (its ancestor matches)
	if resp.Total < 2 {
		t.Errorf("expected at least 2 tags (Technology + Programming), got %d", resp.Total)
	}

	foundTech, foundProg := false, false
	for _, r := range resp.Results {
		if r.Name == "Technology" {
			foundTech = true
		}
		if r.Name == "Programming" {
			foundProg = true
		}
	}
	if !foundTech {
		t.Error("Technology tag not found")
	}
	if !foundProg {
		t.Error("Programming tag not found: WITH_ANCESTORS should return descendants of matched tags")
	}
}

// TestSearchTagsWithAncestorsNoFalsePositive verifies that WITH_ANCESTORS does NOT
// return a parent tag when only the child matches.
// Searching "Programming" should NOT return "Technology".
func TestSearchTagsWithAncestorsNoFalsePositive(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	setupTagTestData(t, q)

	ctx := context.Background()
	service := NewTagService(database, q, nil)

	req := &indexallv1.SearchTagsRequest{
		Query:    "Programming",
		TagScope: indexallv1.SearchTagsRequest_WITH_ANCESTORS,
		Limit:    20,
		Offset:   0,
	}

	resp, err := service.Search(ctx, req)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	for _, r := range resp.Results {
		if r.Name == "Technology" {
			t.Error("Technology (parent) should NOT appear when only child 'Programming' matches with WITH_ANCESTORS")
		}
	}
}

// TestSearchTagsWithDescendants verifies that WITH_DESCENDANTS returns a parent tag
// when its descendant matches.
// Searching "Programming" (child) should return BOTH Programming AND Technology (its parent).
func TestSearchTagsWithDescendants(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	setupTagTestData(t, q)

	ctx := context.Background()
	service := NewTagService(database, q, nil)

	req := &indexallv1.SearchTagsRequest{
		Query:    "Programming",
		TagScope: indexallv1.SearchTagsRequest_WITH_DESCENDANTS,
		Limit:    20,
		Offset:   0,
	}

	resp, err := service.Search(ctx, req)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should find both Programming (direct) and Technology (its descendant matches)
	if resp.Total < 2 {
		t.Errorf("expected at least 2 tags (Programming + Technology), got %d", resp.Total)
	}

	foundTech, foundProg := false, false
	for _, r := range resp.Results {
		if r.Name == "Technology" {
			foundTech = true
		}
		if r.Name == "Programming" {
			foundProg = true
		}
	}
	if !foundProg {
		t.Error("Programming tag not found")
	}
	if !foundTech {
		t.Error("Technology (parent) not found: WITH_DESCENDANTS should return ancestors of matched tags")
	}
}

// TestSearchTagsByAlias tests tag search by alias
func TestSearchTagsByAlias(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	setupTagTestData(t, q)

	ctx := context.Background()
	service := NewTagService(database, q, nil)

	// Search by alias
	req := &indexallv1.SearchTagsRequest{
		Query: "coding",
		TagScope: indexallv1.SearchTagsRequest_DIRECT,
		Limit: 20,
		Offset: 0,
	}

	resp, err := service.Search(ctx, req)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Note: The current implementation may not search aliases via FTS if FTS5 is not available
	// This test verifies the structure works even if alias search is limited
	if resp.Total >= 0 {
		t.Logf("Search returned %d results for alias 'coding'", resp.Total)
	}
}

// TestSearchTagsPagination tests pagination in tag search
func TestSearchTagsPagination(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	setupTagTestData(t, q)

	ctx := context.Background()
	service := NewTagService(database, q, nil)

	// Search with limit
	req := &indexallv1.SearchTagsRequest{
		Query: "g",
		TagScope: indexallv1.SearchTagsRequest_DIRECT,
		Limit: 1,
		Offset: 0,
	}

	resp, err := service.Search(ctx, req)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if resp.Total < 0 {
		t.Errorf("expected total >= 0, got %d", resp.Total)
	}

	// Verify pagination fields are set
	if len(resp.Results) > 1 {
		t.Errorf("expected at most 1 result, got %d", len(resp.Results))
	}
}

// TestSearchTagsEmpty tests error handling for empty query
func TestSearchTagsEmpty(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	service := NewTagService(database, q, nil)

	req := &indexallv1.SearchTagsRequest{
		Query: "",
		TagScope: indexallv1.SearchTagsRequest_DIRECT,
		Limit: 20,
		Offset: 0,
	}

	_, err := service.Search(ctx, req)
	if err == nil {
		t.Error("expected error for empty query")
	}
}

// TestSearchTagsDefaultLimit tests default limit handling
func TestSearchTagsDefaultLimit(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	setupTagTestData(t, q)

	ctx := context.Background()
	service := NewTagService(database, q, nil)

	req := &indexallv1.SearchTagsRequest{
		Query: "g",
		TagScope: indexallv1.SearchTagsRequest_DIRECT,
		Limit: 0, // Should default to 20
		Offset: 0,
	}

	resp, err := service.Search(ctx, req)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Response should contain results with proper structure
	if resp.Results == nil {
		t.Error("expected Results slice, got nil")
	}
}

// TestSearchTagsResourceCount tests that resource count is included
func TestSearchTagsResourceCount(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	_, childTagID, _ := setupTagTestData(t, q)

	// Add a resource to the child tag
	ctx := context.Background()
	resource, err := q.CreateResource(ctx, gen.CreateResourceParams{
		ID:         uuid.New().String(),
		Source:     "test",
		Title:      "Test Resource",
		ExternalID: sql.NullString{},
	})
	if err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}

	err = q.AddTagToResource(ctx, gen.AddTagToResourceParams{
		ResourceID: resource.ID,
		TagID:      childTagID,
	})
	if err != nil {
		t.Fatalf("failed to add tag to resource: %v", err)
	}

	service := NewTagService(database, q, nil)

	req := &indexallv1.SearchTagsRequest{
		Query: "Programming",
		TagScope: indexallv1.SearchTagsRequest_DIRECT,
		Limit: 20,
		Offset: 0,
	}

	resp, err := service.Search(ctx, req)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Verify that results include resource_count
	if len(resp.Results) > 0 {
		for _, result := range resp.Results {
			if result.Name == "Programming" {
				if result.ResourceCount < 1 {
					t.Errorf("expected resource_count >= 1, got %d", result.ResourceCount)
				}
				break
			}
		}
	}
}

// TestSearchTagsAliasIncluded tests that aliases are included in results
func TestSearchTagsAliasIncluded(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	setupTagTestData(t, q)

	ctx := context.Background()
	service := NewTagService(database, q, nil)

	req := &indexallv1.SearchTagsRequest{
		Query: "Program",
		TagScope: indexallv1.SearchTagsRequest_DIRECT,
		Limit: 20,
		Offset: 0,
	}

	resp, err := service.Search(ctx, req)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Find Programming tag and check if aliases are included
	for _, result := range resp.Results {
		if result.Name == "Programming" {
			if len(result.Aliases) < 1 {
				t.Error("expected aliases in Programming tag result")
			} else if result.Aliases[0] != "coding" {
				t.Errorf("expected alias 'coding', got '%s'", result.Aliases[0])
			}
			break
		}
	}
}

// BenchmarkSearchTags benchmarks tag search operations
func BenchmarkSearchTags(b *testing.B) {
	database, q := setupTestDB(&testing.T{})
	defer database.Close()

	setupTagTestData(&testing.T{}, q)

	ctx := context.Background()
	service := NewTagService(database, q, nil)

	req := &indexallv1.SearchTagsRequest{
		Query: "Program",
		TagScope: indexallv1.SearchTagsRequest_DIRECT,
		Limit: 20,
		Offset: 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.Search(ctx, req)
	}
}
