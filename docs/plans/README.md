# Execution Plans

Plans are versioned work artifacts for agent execution. Use them when a change spans multiple files, workflows, or validation surfaces.

## Structure

- `active/`: currently relevant plans.
- `completed/`: finished plans with evidence.

## Plan Template

```markdown
# Plan: <name>

## Goal

## Context

## In Scope

## Out Of Scope

## Steps

- [ ] Step

## Validation

## Risks And Rollback

## Decision Log
```

Small one-file fixes can skip a plan, but product or architecture changes should update one.
