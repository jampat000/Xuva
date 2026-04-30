# Plan: Issue 16 Migration Tooling

## Goal

Ship a real migration path for watched state, resume position, and core metadata identifiers so Vyrden is easier to adopt from Plex, Emby, or Jellyfin.

## Context

- Issue: `#16 P3.3 Migration and Adoption Tooling`
- Dependencies are closed in issue `#1`.
- Vyrden already has stable catalog, playback-state, metadata, auth, and route-policy foundations.

## In Scope

- supported import format definition
- dry-run parser and conflict report
- selective import execution
- rollback storage and rollback endpoint
- verification report
- admin UI in Settings
- sample fixtures and tests
- documentation

## Out Of Scope

- plugin migration
- cloud-account migration
- direct database scraping from Plex, Emby, or Jellyfin

## Steps

- [ ] Add migration persistence tables and service.
- [ ] Add dry-run/import/rollback API routes and tests.
- [ ] Add Settings UI for preview/import/rollback.
- [ ] Add docs and fixture coverage.
- [ ] Run full verification sweep.

## Validation

- `go test ./internal/migration ./internal/api`
- `go test ./...`
- `node --test server/internal/webapp/frontend_tests/*.test.cjs`
- `node --check server/internal/webapp/static/app.js`
- `git diff --check`

## Risks And Rollback

- Playback state is version-specific, so imports must avoid ambiguous multi-version matches.
- Rollback must restore both playback state and metadata IDs.
- If the feature misbehaves, revert the migration routes/service/UI and leave the new DB tables in place; older builds ignore them.

## Decision Log

- Use one normalized JSON bundle format with source tags instead of pretending Vyrden can safely ingest every vendor's private export format directly.
- Prefer dry-run plus selective import over best-effort silent skips.
