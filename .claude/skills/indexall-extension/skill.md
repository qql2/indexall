---
name: indexall-extension
description: IndexAll browser extension development standards. Use when editing files in packages/extension/** for WXT framework, popup UI, content scripts, background scripts, page capture, tag selection, or bookmark import functionality.
---

# IndexAll Browser Extension Development Standards

## Pre-Development Checklist

- Extension UI → Read `UI_DESIGN.md` first (see "Browser Extension" section)
- API calls needed → Read `API_DESIGN.md` first (see "Browser Extension Interfaces" section)

## Framework

- WXT (modern browser extension framework supporting React + TypeScript + HMR)
- Popup window is the main interaction interface

## Interaction Flow

1. Open popup → Auto-fill current page title and URL
2. Call `resource.getByUrl` to check if already bookmarked
3. Not bookmarked → Select tags → `resource.create` to save
4. Already bookmarked → Show existing tags, button changes to "Update", can add more tags → `resource.addTag`

## Key Conventions

- Tag selector **identical** to Web UI (reuse from shared or keep behavior consistent)
- Show success message after saving
- Bottom hint shows number of bookmarks from this domain

## Post-Development Maintenance

- Modify extension UI or interactions → Update `UI_DESIGN.md` browser extension section
