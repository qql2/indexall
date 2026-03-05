---
name: IndexAll Assistant
description: >
  Manage and organize your resource collection using IndexAll. Use when: (1) user wants to know
  what resources they have saved or bookmarked, (2) user needs to save, record, or index a new
  resource (URL, article, tool, etc.), (3) organizing tags, searching resources, or maintaining
  your knowledge base. Provides guided workflows for resource discovery, tag management, bulk
  operations, and taxonomy planning. Works with the IndexAll MCP to query and structure your resources.
---

# IndexAll Assistant

Structured workflows for managing your resource library through IndexAll.

## Quick Start

### Add a New Resource
1. Describe the resource (title, URL, topic)
2. I'll create it and assign appropriate tags
3. Verify the resource is discoverable via search

### Find Resources
1. Tell me what you're looking for (topic, keyword, tag)
2. I'll search using your existing taxonomy
3. Show you results and refine as needed

### Organize Your Tags
1. Review your current tag structure
2. Plan new tags or reorganize existing ones
3. Create parent-child relationships to build your taxonomy

## Core Workflows

### Workflow 1: Index New Resources

**Goal**: Add a batch of new resources and categorize them properly

**Steps**:
1. **List existing resources** in the topic area using `query_resources` with tag scope
2. **Create resource** (see tool selection below)
3. **Assign tags** by calling `add_tag_to_resource`
4. **Verify** by searching to confirm discoverability

**Tool selection by resource type**:
- **Local file** (`/Users/...`, `/home/...`): use `index_local_file(path)` — NOT `create_resource`. The filesystem connector (daemon) manages these with a specific format; using `create_resource` directly breaks move/deletion tracking.
- **Web URL, GitHub repo, Notion page, etc.**: use `create_resource` with title, url, source, description.

**Best for**: Importing from bookmarks, adding research materials, building collections

### Workflow 2: Refactor Tag Taxonomy

**Goal**: Improve your tag hierarchy (flatten, restructure, consolidate)

**Steps**:
1. **Review structure** using `get_tag_tree` to see current organization
2. **Identify gaps** - missing tags or unclear relationships
3. **Plan changes** - what should be created, merged, or reorganized?
4. **Execute** - create new tags, update parent-child relationships via `add_tag_parent`/`remove_tag_parent`
5. **Verify** - confirm resources are still discoverable under new structure

**Best for**: Periodic taxonomy reviews, handling growth, clarifying unclear hierarchies

### Workflow 3: Search and Discover

**Goal**: Find resources relevant to a topic or keyword

**Steps**:
1. **Keyword search** using `query_resources` with keyword
2. **Tag-based search** to find all resources under a category (use `WITH_DESCENDANTS` for subtree)
3. **Narrow results** by scope (parents, descendants, or direct children only)
4. **Examine details** via `get_resource` for metadata

**Best for**: Research, finding inspiration, inventory audits

### Workflow 4: Bulk Organization

**Goal**: Retag or reorganize multiple resources at once

**Steps**:
1. **Query resources** to get the batch (by tag or keyword)
2. **Analyze results** - what's missing or miscategorized?
3. **Update tags** - add new tags or remove incorrect ones
4. **Verify** - confirm resources appear under correct categories

**Best for**: Cleaning up imports, fixing bulk miscategorization

## Understanding Tag Scope

When querying by tag, scope controls the search area:

### DIRECT
Only this specific tag. Use when you want resources tagged *exactly* with this tag, nothing broader or narrower.

**Example**: Resources tagged "Python Basics" only (not "Python" or "Python Advanced")

### WITH_DESCENDANTS
This tag + all children. Use when you want everything in a category *and* its subtopics.

**Example**: "Python" + "Python Basics" + "Python Advanced" = all Python learning materials

### WITH_ANCESTORS
This tag + all parents. Use when you want a resource *and* its broader categories.

**Example**: "Python Basics" + "Python" + "Programming" = context and hierarchy

## Tips

- **Use aliases** for tags with multiple names: `add_tag_alias` if "React" should also match "ReactJS"
- **Build hierarchy intentionally**: Parent-child relationships should be meaningful, not just broad-to-narrow
- **Regular audits**: Periodically `query_resources` to find orphaned or miscategorized items
- **Tag sparingly**: More tags = harder to find things. Tag by function, not description

## Customization

Your taxonomy is documented in [`references/taxonomy.md`](references/taxonomy.md). Review and update it to:
- Document your tag structure
- Explain why tags are organized as they are
- Define what each branch of the hierarchy is for
- Store examples of resources in each category

This helps me understand your organization and make better suggestions.
