package vault

// OpType represents the type of operation
type OpType string

const (
	OpCreate OpType = "create"
	OpUpdate OpType = "update"
	OpDelete OpType = "delete"
)

// EntityType represents the type of entity being operated on
type EntityType string

const (
	EntityTag         EntityType = "tag"
	EntityResource    EntityType = "resource"
	EntityTagAlias    EntityType = "tag_alias"
	EntityTagRelation EntityType = "tag_relation"
	EntityResourceTag EntityType = "resource_tag"
)

// Entry is a single operation recorded in the vault log
type Entry struct {
	ID       string     `json:"id"`        // unique op ID (UUID)
	Op       OpType     `json:"op"`        // create | update | delete
	Type     EntityType `json:"type"`      // entity type
	EntityID string     `json:"entity_id"` // ID of the affected entity
	Data     any        `json:"data,omitempty"`
	TS       string     `json:"ts"` // RFC3339Nano timestamp
}

// TagData is the data payload for tag create/update ops
type TagData struct {
	ID        string   `json:"id"`
	Name      string   `json:"name,omitempty"`
	Color     string   `json:"color,omitempty"`
	Aliases   []string `json:"aliases,omitempty"` // only in create op
	Parents   []string `json:"parents,omitempty"` // only in create op
	CreatedAt string   `json:"created_at,omitempty"`
}

// ResourceData is the data payload for resource create/update ops
type ResourceData struct {
	ID          string   `json:"id"`
	Source      string   `json:"source,omitempty"`
	ExternalID  string   `json:"external_id,omitempty"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	URL         string   `json:"url,omitempty"`
	OpenWith    string   `json:"open_with,omitempty"`
	Status      string   `json:"status,omitempty"`
	Tags        []string `json:"tags,omitempty"` // only in create op
	CreatedAt   string   `json:"created_at,omitempty"`
}

// TagAliasData is the data payload for tag_alias ops
type TagAliasData struct {
	ID    string `json:"id"`
	TagID string `json:"tag_id"`
	Alias string `json:"alias"`
}

// TagRelationData is the data payload for tag_relation ops
type TagRelationData struct {
	ParentID string `json:"parent_id"`
	ChildID  string `json:"child_id"`
}

// ResourceTagData is the data payload for resource_tag ops
type ResourceTagData struct {
	ResourceID string `json:"resource_id"`
	TagID      string `json:"tag_id"`
}
