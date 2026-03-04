package service

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	indexallv1 "github.com/construct/indexall/api/indexall/v1"
	"github.com/construct/indexall/internal/db"
	"github.com/construct/indexall/internal/db/gen"
	"github.com/google/uuid"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) (*sql.DB, *gen.Queries) {
	// Use in-memory database for testing
	dbPath := ":memory:"
	database, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init test database: %v", err)
	}

	q := gen.New(database)
	return database, q
}

// setupTestData creates sample tags and resources for testing
func setupTestData(t *testing.T, q *gen.Queries) (string, string, string, string) {
	ctx := context.Background()

	// Create tags
	learningTag, err := q.CreateTag(ctx, gen.CreateTagParams{
		ID:    uuid.New().String(),
		Name:  "Learning",
		Color: sql.NullString{},
	})
	if err != nil {
		t.Fatalf("failed to create Learning tag: %v", err)
	}

	pythonTag, err := q.CreateTag(ctx, gen.CreateTagParams{
		ID:    uuid.New().String(),
		Name:  "Python",
		Color: sql.NullString{},
	})
	if err != nil {
		t.Fatalf("failed to create Python tag: %v", err)
	}

	// Create parent-child relationship: Learning -> Python
	err = q.CreateTagRelation(ctx, gen.CreateTagRelationParams{
		ParentID: learningTag.ID,
		ChildID:  pythonTag.ID,
	})
	if err != nil {
		t.Fatalf("failed to create tag relation: %v", err)
	}

	// Create resources
	resource1, err := q.CreateResource(ctx, gen.CreateResourceParams{
		ID:          uuid.New().String(),
		Source:      "manual",
		ExternalID:  sql.NullString{},
		Title:       "Python Basics Tutorial",
		Description: sql.NullString{String: "Learn Python fundamentals", Valid: true},
		Url:         sql.NullString{String: "https://example.com/python-basics", Valid: true},
		OpenWith:    sql.NullString{},
		Metadata:    sql.NullString{},
	})
	if err != nil {
		t.Fatalf("failed to create resource1: %v", err)
	}

	resource2, err := q.CreateResource(ctx, gen.CreateResourceParams{
		ID:          uuid.New().String(),
		Source:      "manual",
		ExternalID:  sql.NullString{},
		Title:       "Advanced Python Patterns",
		Description: sql.NullString{String: "Master Python design patterns", Valid: true},
		Url:         sql.NullString{String: "https://example.com/python-patterns", Valid: true},
		OpenWith:    sql.NullString{},
		Metadata:    sql.NullString{},
	})
	if err != nil {
		t.Fatalf("failed to create resource2: %v", err)
	}

	// Assign tags to resources
	err = q.AddTagToResource(ctx, gen.AddTagToResourceParams{
		ResourceID: resource1.ID,
		TagID:      pythonTag.ID,
	})
	if err != nil {
		t.Fatalf("failed to add tag to resource1: %v", err)
	}

	err = q.AddTagToResource(ctx, gen.AddTagToResourceParams{
		ResourceID: resource2.ID,
		TagID:      pythonTag.ID,
	})
	if err != nil {
		t.Fatalf("failed to add tag to resource2: %v", err)
	}

	return learningTag.ID, pythonTag.ID, resource1.ID, resource2.ID
}

// TestQueryByTagDirect tests TagQuery with DIRECT scope
func TestQueryByTagDirect(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	_, pythonTagID, _, _ := setupTestData(t, q)

	ctx := context.Background()
	service := NewResourceService(database, q)

	// Query by tag with DIRECT scope
	req := &indexallv1.ResourceQueryRequest{
		Query: &indexallv1.ResourceQueryRequest_TagQuery{
			TagQuery: &indexallv1.TagQuery{
				TagId:    pythonTagID,
				TagScope: indexallv1.TagQuery_DIRECT,
			},
		},
		Page:     1,
		PageSize: 20,
	}

	resp, err := service.Query(ctx, req)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if resp.Total != 2 {
		t.Errorf("expected 2 resources, got %d", resp.Total)
	}

	if len(resp.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Items))
	}

	// Check resources
	titles := make(map[string]bool)
	for _, item := range resp.Items {
		titles[item.Title] = true
	}

	if !titles["Python Basics Tutorial"] {
		t.Error("Python Basics Tutorial not found")
	}
	if !titles["Advanced Python Patterns"] {
		t.Error("Advanced Python Patterns not found")
	}
}

// TestQueryByTagWithDescendants tests TagQuery with WITH_DESCENDANTS scope
func TestQueryByTagWithDescendants(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	learningTagID, _, _, _ := setupTestData(t, q)

	ctx := context.Background()
	service := NewResourceService(database, q)

	// Query by tag with WITH_DESCENDANTS scope
	req := &indexallv1.ResourceQueryRequest{
		Query: &indexallv1.ResourceQueryRequest_TagQuery{
			TagQuery: &indexallv1.TagQuery{
				TagId:    learningTagID,
				TagScope: indexallv1.TagQuery_WITH_DESCENDANTS,
			},
		},
		Page:     1,
		PageSize: 20,
	}

	resp, err := service.Query(ctx, req)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Should include Python resources (descendants)
	if resp.Total != 2 {
		t.Errorf("expected 2 resources, got %d", resp.Total)
	}

	if len(resp.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Items))
	}
}

// TestQueryByKeywordTitle tests KeywordQuery with TITLE field scope
func TestQueryByKeywordTitle(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	setupTestData(t, q)

	ctx := context.Background()
	service := NewResourceService(database, q)

	// Query by keyword in title
	req := &indexallv1.ResourceQueryRequest{
		Query: &indexallv1.ResourceQueryRequest_KeywordQuery{
			KeywordQuery: &indexallv1.KeywordQuery{
				Keyword:    "Basics",
				FieldScope: indexallv1.KeywordQuery_TITLE,
				TagScope:   indexallv1.KeywordQuery_DIRECT,
			},
		},
		Page:     1,
		PageSize: 20,
	}

	resp, err := service.Query(ctx, req)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if resp.Total != 1 {
		t.Errorf("expected 1 resource, got %d", resp.Total)
	}

	if len(resp.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.Items))
	}

	if resp.Items[0].Title != "Python Basics Tutorial" {
		t.Errorf("expected 'Python Basics Tutorial', got '%s'", resp.Items[0].Title)
	}
}

// TestQueryByKeywordDescription tests KeywordQuery with DESCRIPTION field scope
func TestQueryByKeywordDescription(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	setupTestData(t, q)

	ctx := context.Background()
	service := NewResourceService(database, q)

	// Query by keyword in description
	req := &indexallv1.ResourceQueryRequest{
		Query: &indexallv1.ResourceQueryRequest_KeywordQuery{
			KeywordQuery: &indexallv1.KeywordQuery{
				Keyword:    "design patterns",
				FieldScope: indexallv1.KeywordQuery_DESCRIPTION,
				TagScope:   indexallv1.KeywordQuery_DIRECT,
			},
		},
		Page:     1,
		PageSize: 20,
	}

	resp, err := service.Query(ctx, req)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if resp.Total != 1 {
		t.Errorf("expected 1 resource, got %d", resp.Total)
	}

	if resp.Items[0].Title != "Advanced Python Patterns" {
		t.Errorf("expected 'Advanced Python Patterns', got '%s'", resp.Items[0].Title)
	}
}

// TestQueryByKeywordAll tests KeywordQuery with ALL field scope
func TestQueryByKeywordAll(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	setupTestData(t, q)

	ctx := context.Background()
	service := NewResourceService(database, q)

	// Query by keyword in all fields
	req := &indexallv1.ResourceQueryRequest{
		Query: &indexallv1.ResourceQueryRequest_KeywordQuery{
			KeywordQuery: &indexallv1.KeywordQuery{
				Keyword:    "Python",
				FieldScope: indexallv1.KeywordQuery_ALL,
				TagScope:   indexallv1.KeywordQuery_DIRECT,
			},
		},
		Page:     1,
		PageSize: 20,
	}

	resp, err := service.Query(ctx, req)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Should find both resources with "Python" in title
	if resp.Total != 2 {
		t.Errorf("expected 2 resources, got %d", resp.Total)
	}
}

// TestQueryPagination tests pagination functionality
func TestQueryPagination(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	_, pythonTagID, _, _ := setupTestData(t, q)

	ctx := context.Background()
	service := NewResourceService(database, q)

	// Query with page size 1
	req := &indexallv1.ResourceQueryRequest{
		Query: &indexallv1.ResourceQueryRequest_TagQuery{
			TagQuery: &indexallv1.TagQuery{
				TagId:    pythonTagID,
				TagScope: indexallv1.TagQuery_DIRECT,
			},
		},
		Page:     1,
		PageSize: 1,
	}

	resp, err := service.Query(ctx, req)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if resp.Total != 2 {
		t.Errorf("expected total 2, got %d", resp.Total)
	}

	if resp.PageSize != 1 {
		t.Errorf("expected page_size 1, got %d", resp.PageSize)
	}

	if len(resp.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.Items))
	}

	// Query page 2
	req.Page = 2
	resp, err = service.Query(ctx, req)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(resp.Items) != 1 {
		t.Errorf("expected 1 item on page 2, got %d", len(resp.Items))
	}
}

// TestQueryTagInfo tests that response includes tag information
func TestQueryTagInfo(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	_, pythonTagID, _, _ := setupTestData(t, q)

	ctx := context.Background()
	service := NewResourceService(database, q)

	req := &indexallv1.ResourceQueryRequest{
		Query: &indexallv1.ResourceQueryRequest_TagQuery{
			TagQuery: &indexallv1.TagQuery{
				TagId:    pythonTagID,
				TagScope: indexallv1.TagQuery_DIRECT,
			},
		},
		Page:     1,
		PageSize: 20,
	}

	resp, err := service.Query(ctx, req)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(resp.Items) == 0 {
		t.Fatal("no items returned")
	}

	item := resp.Items[0]
	if len(item.Tags) == 0 {
		t.Error("expected tags in response")
	}

	if item.Tags[0].Name != "Python" {
		t.Errorf("expected tag name 'Python', got '%s'", item.Tags[0].Name)
	}
}

// TestQueryInvalidTagID tests error handling for invalid tag ID
func TestQueryInvalidTagID(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	service := NewResourceService(database, q)

	req := &indexallv1.ResourceQueryRequest{
		Query: &indexallv1.ResourceQueryRequest_TagQuery{
			TagQuery: &indexallv1.TagQuery{
				TagId:    "nonexistent-tag",
				TagScope: indexallv1.TagQuery_DIRECT,
			},
		},
		Page:     1,
		PageSize: 20,
	}

	_, err := service.Query(ctx, req)
	if err == nil {
		t.Error("expected error for invalid tag ID")
	}
}

// TestQueryEmptyKeyword tests error handling for empty keyword
func TestQueryEmptyKeyword(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	service := NewResourceService(database, q)

	req := &indexallv1.ResourceQueryRequest{
		Query: &indexallv1.ResourceQueryRequest_KeywordQuery{
			KeywordQuery: &indexallv1.KeywordQuery{
				Keyword:    "",
				FieldScope: indexallv1.KeywordQuery_ALL,
				TagScope:   indexallv1.KeywordQuery_DIRECT,
			},
		},
		Page:     1,
		PageSize: 20,
	}

	_, err := service.Query(ctx, req)
	if err == nil {
		t.Error("expected error for empty keyword")
	}
}

// TestQueryNoQuery tests error handling when no query is provided
func TestQueryNoQuery(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	service := NewResourceService(database, q)

	req := &indexallv1.ResourceQueryRequest{
		Query:    nil,
		Page:     1,
		PageSize: 20,
	}

	_, err := service.Query(ctx, req)
	if err == nil {
		t.Error("expected error when no query is provided")
	}
}

// TestQueryDefaultPageSize tests default page size
func TestQueryDefaultPageSize(t *testing.T) {
	database, q := setupTestDB(t)
	defer database.Close()

	_, pythonTagID, _, _ := setupTestData(t, q)

	ctx := context.Background()
	service := NewResourceService(database, q)

	req := &indexallv1.ResourceQueryRequest{
		Query: &indexallv1.ResourceQueryRequest_TagQuery{
			TagQuery: &indexallv1.TagQuery{
				TagId:    pythonTagID,
				TagScope: indexallv1.TagQuery_DIRECT,
			},
		},
		Page:     1,
		PageSize: 0, // Should default to 20
	}

	resp, err := service.Query(ctx, req)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if resp.PageSize != 20 {
		t.Errorf("expected default page_size 20, got %d", resp.PageSize)
	}
}

// BenchmarkQueryByTag benchmarks tag-based queries
func BenchmarkQueryByTag(b *testing.B) {
	database, q := setupTestDB(&testing.T{})
	defer database.Close()

	_, pythonTagID, _, _ := setupTestData(&testing.T{}, q)

	ctx := context.Background()
	service := NewResourceService(database, q)

	req := &indexallv1.ResourceQueryRequest{
		Query: &indexallv1.ResourceQueryRequest_TagQuery{
			TagQuery: &indexallv1.TagQuery{
				TagId:    pythonTagID,
				TagScope: indexallv1.TagQuery_DIRECT,
			},
		},
		Page:     1,
		PageSize: 20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.Query(ctx, req)
	}
}

// BenchmarkQueryByKeyword benchmarks keyword-based queries
func BenchmarkQueryByKeyword(b *testing.B) {
	database, q := setupTestDB(&testing.T{})
	defer database.Close()

	setupTestData(&testing.T{}, q)

	ctx := context.Background()
	service := NewResourceService(database, q)

	req := &indexallv1.ResourceQueryRequest{
		Query: &indexallv1.ResourceQueryRequest_KeywordQuery{
			KeywordQuery: &indexallv1.KeywordQuery{
				Keyword:    "Python",
				FieldScope: indexallv1.KeywordQuery_TITLE,
				TagScope:   indexallv1.KeywordQuery_DIRECT,
			},
		},
		Page:     1,
		PageSize: 20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.Query(ctx, req)
	}
}
