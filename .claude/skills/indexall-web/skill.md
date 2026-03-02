---
name: indexall-web
description: IndexAll Web frontend development standards. Use when editing files in packages/web/** for UI implementation, React components, TailwindCSS styling, shadcn/ui integration, page layouts, interactions, TagPicker, tag tree rendering, or responsive design.
---

# IndexAll Web Frontend Development Standards

## Pre-Development Checklist

- UI implementation → Read `UI_DESIGN.md` first (layout, components, interaction details)
- API calls needed → Read `API_DESIGN.md` first (see "Client Call Scenarios → Web UI Interfaces" section)

## UI Framework

- React + TailwindCSS + shadcn/ui
- **Single Page Application, no routing**, all operations complete within same view

## Layout Structure

- **Top bar**: Logo + global search bar (centered) + add resource button (right)
- **Left panel**: Tag tree panel (collapsible), top has "All Resources" and "Uncategorized" virtual nodes
- **Right panel**: Resource list (default descending by time), pagination at bottom
- **Dialogs**: Add/edit resource dialog, edit tag dialog

## Key Interactions

- **Tag tree DAG display**: Same tag appears **equally** under all parent tags, no primary/secondary distinction, hover tooltip shows multi-parent relationship
- **TagPicker**: Globally reused component, search on input for tag names + aliases, shows "Create new tag" option when no exact match
- **Search**: debounce 300ms, clears tag selection when searching, restores previous tag filter when search bar is cleared
- **Resource cards**: Title clickable to jump, tag pills clickable to filter, hover shows edit/delete buttons

## Responsive Design

| Viewport | Layout |
|---|---|
| ≥1024px | Left-right split, tag tree always visible |
| 768-1023px | Tag tree collapses to side drawer |
| <768px | Tag tree in fullscreen drawer, resource list single column |

## Post-Development Maintenance

- Add/modify UI components or interaction patterns → Update `UI_DESIGN.md`
