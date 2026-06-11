# CSV Import: Duplicate Handling, Upsert, Column Mapping & Parenting

**Date:** 2026-06-12
**Issue:** kopexa-grc/kopexa#878 ("CSV Import duplicate check")
**Repos:** kopexa (backend), kopexa-frontend

## Problem

Importing an asset CSV into a space that already contains assets fails without a
helpful error. Root causes:

1. `bulkCreateAsset` batches all rows into a single `CreateBulk().Save()` inside
   one transaction. The first duplicate row rolls back the entire import; the
   client only receives a generic `ALREADY_EXISTS`.
2. `internal_id` / `external_id` are `Optional()` but not normalized: empty CSV
   cells are stored as `""` instead of NULL. The partial unique indexes
   `(space_id, internal_id)` and `(space_id, external_id)` treat `""` as a real
   value, so two rows with empty IDs collide — the likely trigger of the
   customer report.
3. The result payload carries no counts, no row-level issues, no partial
   success. The frontend can only show a generic toast.

The customer expects an upsert; behavior should be configurable.

## Decisions (validated with Julian)

| Topic | Decision |
| --- | --- |
| Strategies | `CREATE` (strict, all-or-nothing), `SKIP` (skip duplicates), `UPSERT` |
| API default | `CREATE` — callers without `options` keep today's semantics (plus a populated report) |
| Frontend default | `SKIP` |
| Empty cells on upsert | Configurable: `IGNORE` (default — non-empty CSV values only) or `CLEAR` (CSV is the full truth) |
| Result UI | Result step inside the import drawer (counts + expandable row issues); toast only for hard errors |
| Parenting | New CSV column `parent_external_id` (comma-separated for multiple), resolved against existing assets and rows of the same file; unresolvable → warning, asset still created |
| Column mapping | `columnMapping` overrides in the options + mapping step in the frontend wizard |
| Architecture | Generic import pipeline with per-entity adapters (approach B); assets first |
| Async imports | Out of scope — follow-up (River job can reuse the same pipeline later) |

## GraphQL API

```graphql
createBulkCSVAsset(input: Upload!, options: BulkImportOptions): AssetBulkCreatePayload!

input BulkImportOptions {
  strategy: ImportStrategy = CREATE        # CREATE | SKIP | UPSERT
  emptyCells: ImportEmptyCellMode = IGNORE # IGNORE | CLEAR (UPSERT only)
  columnMapping: [ImportColumnMapping!]    # [{ column: "Hostname", field: "name" }]
}

input ImportColumnMapping {
  column: String!  # CSV header as found in the file
  field: String!   # target input field name
}

# AssetBulkCreatePayload, extended (non-breaking):
type AssetBulkCreatePayload {
  assets: [Asset!]
  createdCount: Int!
  updatedCount: Int!
  skippedCount: Int!
  rowIssues: [ImportRowIssue!]!
}

type ImportRowIssue {
  row: Int!            # 1-based CSV line (header = line 1)
  identifier: String   # internal_id/external_id/name of the row if available
  code: ImportRowIssueCode!
  message: String!
}

enum ImportRowIssueCode {
  DUPLICATE          # CREATE/SKIP: row matches an existing asset
  CONFLICTING_MATCH  # internal_id and external_id match two different assets
  PARENT_NOT_FOUND   # parent_external_id could not be resolved
  INVALID_ROW        # row failed parsing/validation
}
```

Enums live in `pkg/enums` (check existing naming conventions before adding —
Source/Status/Type enums are domain-prefixed).

Backward compatibility: calling the mutation without `options` behaves exactly
like today (`CREATE`, all-or-nothing) — only the report fields are now
populated so the client can show which rows collided.

## Backend Pipeline (approach B)

New file `internal/channels/graphapi/bulk_import.go`. The generic runner owns
the flow; entity-specific bits live in a small adapter interface:

```
parse      csvutil; columnMapping applied as a header rewrite before unmarshal
normalize  "" → unset for internal_id / external_id (fixes root cause #2)
prefetch   one internalIDIn(...) + one externalIDIn(...) query in the space
           → two lookup maps; no N+1
classify   per row: create | update | skip | issue
           match internal_id first, then external_id;
           both match different assets → CONFLICTING_MATCH;
           rows without either ID always create (name is NOT a match key)
execute    creates via CreateBulk; updates via Update().SetInput()
           (hooks + privacy run unchanged — no SQL-level OnConflict);
           UPSERT + IGNORE: only set non-empty CSV values
post       parent resolution: parent_external_id (comma-separated) resolved
           against existing assets + freshly created rows of this import;
           unresolved → PARENT_NOT_FOUND warning, asset kept
report     counts + rowIssues; row number = slice index + 2
```

Adapter surface (assets implement first): extract match keys, lookup query,
create builder, update builder, optional post-phase. The 35+ structurally
identical `bulkCreate*` handlers get a migration path; other entities are
follow-ups.

Transactions: everything runs inside the existing entgql Transactioner
transaction. A race on the unique index between prefetch and insert rolls back
with an error — accepted (rare; pre-check covers the normal case).

### Empty-string normalization beyond the import

- Asset hook: normalize `""` → clear for `internal_id` / `external_id` on all
  create/update mutations so the bug cannot return through other paths.
- Data migration for existing `""` values — one consolidated migration at the
  end of the feature branch (per project convention, via `./kopexa-gen migrate`).

## Frontend: Import Wizard

The current single-step upload drawer
(`src/modules/inventory/assets/components/import-asset-drawer.tsx`) becomes a
four-step wizard:

1. **File** — pick CSV (as today).
2. **Mapping** — parse the header row client-side; one select per CSV column
   mapping to an asset field or "ignore". Exact/known header names are
   auto-matched. Required-field check: `name` must be mapped.
3. **Options** — strategy radio (default `SKIP`); selecting `UPSERT` reveals
   the empty-cells choice (default "ignore empty cells").
4. **Result** — "12 created · 3 updated · 2 skipped" plus expandable row
   issues ("row 7: internal_id 'SRV-001' already exists"). Toast remains only
   for hard errors (parse failure, transaction rollback).

GraphQL document gains `options` + the new payload fields; regenerate types
(backend must run on :8080 for codegen).

## Testing

- Unit tests on the pipeline logic: classification (create/update/skip),
  conflicting match, empty-cell IGNORE vs CLEAR, parent resolution including
  same-file references, `""` normalization, row numbering.
- Early spike: verify how csvutil populates pointer fields for empty cells —
  this determines the exact normalization point.
- No tests for generated ent/gqlgen output (project convention).

## Risks

- **csvutil empty-cell semantics** — verify first; the normalization step
  depends on it.
- **Large files** stay synchronous in one transaction — same as today, no
  regression; async (approach C, River job + FileImport) is a follow-up that
  can wrap this pipeline.
- **Mapping step adds a click** to the happy path — mitigated by auto-matching
  known headers.

## Out of Scope

- Async imports / progress UI (approach C)
- Rolling the pipeline out to other entities (People, Vendors are next
  candidates — separate issues)
- Matching by name (not unique by design)
