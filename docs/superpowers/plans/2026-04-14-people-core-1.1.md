# Plan 1.1 — People Core Schema + External Workforce

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the existing `people` entity to support external workforce (contractors, freelancers, interns, temps) as first-class citizens alongside employees, via a `worker_type` discriminator plus vendor linkage and contract-end tracking. Schema fields for the later HR sync feature are added in the same migration so that Plan 1.2 is a pure code delta. **No HR sync framework, no cron sweeps, no mapping UI in this plan** — those are explicit deferrals to Plan 1.2 and Block 2.

**Architecture:** Additive schema change on `people`; one Ent migration; automatic GraphQL field exposure via entgql; three new enums in `pkg/enums/`; unique FK edge to the existing `vendor` schema; admin UI updated to show a `worker_type` selector with conditional vendor picker + contract-end-date field, plus a roster filter. All workflows in later blocks read `worker_type` uniformly — there are no parallel codepaths.

**Tech Stack:** Go 1.23, Ent v0.14, entgql, gqlgen, Atlas migrations, PostgreSQL. Frontend: Next.js 15, React, react-hook-form, zod, next-intl, Tanstack Query via generated hooks. Task runner: `task` (go-task).

**Branches (already created):**
- Backend: `kopexa` repo, branch `feat/people-platform`
- Frontend: `kopexa-frontend` repo, branch `feat/people-platform`
- Docs: `comms` repo, branch `feat/people-platform-roadmap` (this plan lives here)

**Spec reference:** `docs/superpowers/specs/2026-04-14-people-platform-roadmap.md` — Block 1 "People Core + HR Import (incl. External Workforce)".

**What is explicitly OUT of scope for this plan (deferred):**
- HR connector framework (`internal/services/hr_sync/`) — Plan 1.2
- Personio / CSV connectors — Plan 1.2 & 1.3
- `hr_field_mapping` schema — Plan 1.2
- River sync job + audit-trail per run — Plan 1.2
- River cron sweep "contract_end → terminated event" — needs `people_employment_event`, deferred to Block 2
- Admin UI for connector setup, mapping config, sync history — Plan 1.2
- Populating `sync_state` / `last_synced_at` / `sync_error` — Plan 1.2 (fields exist from this plan but stay at default values)

---

## File Structure

| Path | Action | Responsibility |
|---|---|---|
| `pkg/enums/worker_type.go` | Create | `WorkerType` enum + GQL marshalers (`EMPLOYEE`, `CONTRACTOR`, `FREELANCER`, `INTERN`, `TEMP`) |
| `pkg/enums/worker_type_test.go` | Create | Unit tests for `WorkerType` |
| `pkg/enums/people_source.go` | Create | `PeopleSource` enum + GQL marshalers (`PERSONIO`, `CSV`, `MANUAL`, `VENDOR_ONBOARDING`) |
| `pkg/enums/people_source_test.go` | Create | Unit tests for `PeopleSource` |
| `pkg/enums/sync_state.go` | Create | `SyncState` enum + GQL marshalers (`NOT_SYNCED`, `SYNCED`, `DRIFTED`, `PENDING`, `ERROR`) |
| `pkg/enums/sync_state_test.go` | Create | Unit tests for `SyncState` |
| `internal/store/schema/people.go` | Modify | Add 8 new fields + unique edge to `Vendor` |
| `internal/store/schema/vendor.go` | Modify | Add inverse pagination edge to `People` (for vendor → people listing later) |
| `internal/store/ent/**` | Regenerate | `task gen:ent` |
| `db/migrations/**` | Create | Atlas migration via `./kopexa-gen migrate --name add_people_worker_type` |
| `internal/channels/graphapi/**` | Regenerate | `task gen:gql` |
| `internal/channels/graphapi/people_test.go` | Create or extend | Integration tests for create employee + create contractor |
| `src/modules/people/validation/index.ts` | Modify | Add `workerType`, `vendorId`, `contractEndDate` to zod schema |
| `src/modules/people/components/create-form/people-create-form.tsx` | Modify | Add `WorkerTypeField` + conditional vendor picker + conditional `contractEndDate` |
| `src/modules/people/components/edit-form/people-edit-form.tsx` | Modify | Same fields, editable |
| `src/modules/people/components/form/worker-type-field.tsx` | Create | Reusable `<WorkerTypeField />` (Select with enum values + i18n labels) |
| `src/modules/people/components/table/table-config.tsx` | Modify | Add `workerType` column with badge rendering |
| `src/modules/people/components/table/people-table-toolbar.tsx` | Modify | Add `workerType` filter dropdown |
| `messages/en/people.json` | Modify | i18n keys for worker types + new form labels |
| `messages/de/people.json` | Modify | Same, German (du-form) |

---

## Task 1: Create `WorkerType` enum

**Why:** The enum is a standalone primitive with no codebase dependencies — TDD it in isolation before touching the schema so the enum is battle-tested before Ent generates code against it.

**Files:**
- Create: `pkg/enums/worker_type.go`
- Create: `pkg/enums/worker_type_test.go`

- [ ] **Step 1: Write the failing test**

Create `pkg/enums/worker_type_test.go`:

```go
// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package enums_test

import (
	"testing"

	"github.com/kopexa-grc/kopexa/pkg/enums"
	"github.com/stretchr/testify/assert"
)

func TestWorkerType_String(t *testing.T) {
	tests := []struct {
		name  string
		input enums.WorkerType
		want  string
	}{
		{"employee", enums.WorkerTypeEmployee, "EMPLOYEE"},
		{"contractor", enums.WorkerTypeContractor, "CONTRACTOR"},
		{"freelancer", enums.WorkerTypeFreelancer, "FREELANCER"},
		{"intern", enums.WorkerTypeIntern, "INTERN"},
		{"temp", enums.WorkerTypeTemp, "TEMP"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.input.String())
		})
	}
}

func TestWorkerType_Values(t *testing.T) {
	got := enums.WorkerType("").Values()
	assert.ElementsMatch(t, []string{"EMPLOYEE", "CONTRACTOR", "FREELANCER", "INTERN", "TEMP"}, got)
}

func TestToWorkerType(t *testing.T) {
	assert.Equal(t, enums.WorkerTypeContractor, enums.ToWorkerType("contractor"))
	assert.Equal(t, enums.WorkerTypeContractor, enums.ToWorkerType("CONTRACTOR"))
	assert.Equal(t, enums.WorkerTypeInvalid, enums.ToWorkerType("xyz"))
}

func TestWorkerType_IsExternal(t *testing.T) {
	assert.False(t, enums.WorkerTypeEmployee.IsExternal())
	assert.True(t, enums.WorkerTypeContractor.IsExternal())
	assert.True(t, enums.WorkerTypeFreelancer.IsExternal())
	assert.True(t, enums.WorkerTypeIntern.IsExternal())
	assert.True(t, enums.WorkerTypeTemp.IsExternal())
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/enums/ -run TestWorkerType -v`
Expected: FAIL with `undefined: enums.WorkerType`

- [ ] **Step 3: Implement the enum**

Create `pkg/enums/worker_type.go`:

```go
// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package enums

import (
	"fmt"
	"io"
	"strings"
)

// WorkerType classifies a People row as employee or a kind of external workforce.
// The discriminator drives workflow-template selection, report filters and
// HR-sync scoping — all compliance workflows are uniform across worker types,
// only configuration differs.
type WorkerType string

const (
	WorkerTypeEmployee   WorkerType = "EMPLOYEE"
	WorkerTypeContractor WorkerType = "CONTRACTOR"
	WorkerTypeFreelancer WorkerType = "FREELANCER"
	WorkerTypeIntern     WorkerType = "INTERN"
	WorkerTypeTemp       WorkerType = "TEMP"
	WorkerTypeInvalid    WorkerType = "INVALID"
)

// Values returns all valid WorkerType values, as required by Ent for enum fields.
func (WorkerType) Values() []string {
	return []string{
		string(WorkerTypeEmployee),
		string(WorkerTypeContractor),
		string(WorkerTypeFreelancer),
		string(WorkerTypeIntern),
		string(WorkerTypeTemp),
	}
}

func (w WorkerType) String() string { return string(w) }

// IsExternal reports whether the worker type is non-employee (contractor,
// freelancer, intern, temp). Used by the admin UI to decide whether to show
// vendor + contract-end-date fields.
func (w WorkerType) IsExternal() bool {
	return w != WorkerTypeEmployee && w != WorkerTypeInvalid && w != ""
}

// ToWorkerType parses a string into a WorkerType, returning Invalid for
// unknown values.
func ToWorkerType(s string) WorkerType {
	switch strings.ToUpper(s) {
	case WorkerTypeEmployee.String():
		return WorkerTypeEmployee
	case WorkerTypeContractor.String():
		return WorkerTypeContractor
	case WorkerTypeFreelancer.String():
		return WorkerTypeFreelancer
	case WorkerTypeIntern.String():
		return WorkerTypeIntern
	case WorkerTypeTemp.String():
		return WorkerTypeTemp
	default:
		return WorkerTypeInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (w WorkerType) MarshalGQL(out io.Writer) {
	_, _ = out.Write([]byte(`"` + w.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (w *WorkerType) UnmarshalGQL(v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for WorkerType, got: %T", v) //nolint:err113
	}
	*w = WorkerType(s)
	return nil
}

// ToPtr returns a pointer to the WorkerType value.
func (w WorkerType) ToPtr() *WorkerType { return &w }
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/enums/ -run TestWorkerType -v`
Expected: PASS for all five subtests.

- [ ] **Step 5: Commit**

```bash
cd /Users/juliankoehn/workbox/github.com/kopexa-grc/kopexa
git add pkg/enums/worker_type.go pkg/enums/worker_type_test.go
git commit -m "feat(enums): add WorkerType for external workforce classification"
```

---

## Task 2: Create `PeopleSource` enum

**Why:** Tracks where a `People` row originated so HR sync can scope its writes correctly later. Mirrors the Task 1 pattern.

**Files:**
- Create: `pkg/enums/people_source.go`
- Create: `pkg/enums/people_source_test.go`

- [ ] **Step 1: Write the failing test**

Create `pkg/enums/people_source_test.go`:

```go
// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package enums_test

import (
	"testing"

	"github.com/kopexa-grc/kopexa/pkg/enums"
	"github.com/stretchr/testify/assert"
)

func TestPeopleSource_Values(t *testing.T) {
	got := enums.PeopleSource("").Values()
	assert.ElementsMatch(t, []string{"PERSONIO", "CSV", "MANUAL", "VENDOR_ONBOARDING"}, got)
}

func TestToPeopleSource(t *testing.T) {
	assert.Equal(t, enums.PeopleSourcePersonio, enums.ToPeopleSource("personio"))
	assert.Equal(t, enums.PeopleSourceManual, enums.ToPeopleSource("MANUAL"))
	assert.Equal(t, enums.PeopleSourceInvalid, enums.ToPeopleSource("bogus"))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/enums/ -run TestPeopleSource -v`
Expected: FAIL with `undefined: enums.PeopleSource`

- [ ] **Step 3: Implement the enum**

Create `pkg/enums/people_source.go`:

```go
// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package enums

import (
	"fmt"
	"io"
	"strings"
)

// PeopleSource identifies where a People row originated. HR sync scopes its
// writes by source (e.g. Personio sync never touches manual rows by default).
type PeopleSource string

const (
	PeopleSourcePersonio         PeopleSource = "PERSONIO"
	PeopleSourceCSV              PeopleSource = "CSV"
	PeopleSourceManual           PeopleSource = "MANUAL"
	PeopleSourceVendorOnboarding PeopleSource = "VENDOR_ONBOARDING"
	PeopleSourceInvalid          PeopleSource = "INVALID"
)

func (PeopleSource) Values() []string {
	return []string{
		string(PeopleSourcePersonio),
		string(PeopleSourceCSV),
		string(PeopleSourceManual),
		string(PeopleSourceVendorOnboarding),
	}
}

func (s PeopleSource) String() string { return string(s) }

func ToPeopleSource(s string) PeopleSource {
	switch strings.ToUpper(s) {
	case PeopleSourcePersonio.String():
		return PeopleSourcePersonio
	case PeopleSourceCSV.String():
		return PeopleSourceCSV
	case PeopleSourceManual.String():
		return PeopleSourceManual
	case PeopleSourceVendorOnboarding.String():
		return PeopleSourceVendorOnboarding
	default:
		return PeopleSourceInvalid
	}
}

func (s PeopleSource) MarshalGQL(out io.Writer) {
	_, _ = out.Write([]byte(`"` + s.String() + `"`))
}

func (s *PeopleSource) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for PeopleSource, got: %T", v) //nolint:err113
	}
	*s = PeopleSource(str)
	return nil
}

func (s PeopleSource) ToPtr() *PeopleSource { return &s }
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/enums/ -run TestPeopleSource -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/enums/people_source.go pkg/enums/people_source_test.go
git commit -m "feat(enums): add PeopleSource for provenance tracking"
```

---

## Task 3: Create `SyncState` enum

**Files:**
- Create: `pkg/enums/sync_state.go`
- Create: `pkg/enums/sync_state_test.go`

- [ ] **Step 1: Write the failing test**

Create `pkg/enums/sync_state_test.go`:

```go
// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package enums_test

import (
	"testing"

	"github.com/kopexa-grc/kopexa/pkg/enums"
	"github.com/stretchr/testify/assert"
)

func TestSyncState_Values(t *testing.T) {
	got := enums.SyncState("").Values()
	assert.ElementsMatch(t, []string{"NOT_SYNCED", "SYNCED", "DRIFTED", "PENDING", "ERROR"}, got)
}

func TestToSyncState(t *testing.T) {
	assert.Equal(t, enums.SyncStateSynced, enums.ToSyncState("synced"))
	assert.Equal(t, enums.SyncStateNotSynced, enums.ToSyncState("not_synced"))
	assert.Equal(t, enums.SyncStateInvalid, enums.ToSyncState("nope"))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/enums/ -run TestSyncState -v`
Expected: FAIL with `undefined: enums.SyncState`.

- [ ] **Step 3: Implement the enum**

Create `pkg/enums/sync_state.go`:

```go
// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package enums

import (
	"fmt"
	"io"
	"strings"
)

// SyncState captures the relationship of a People row to its HR source.
// NOT_SYNCED is the default for manually-created rows; SYNCED/DRIFTED/PENDING/ERROR
// are set by the HR sync runner in Plan 1.2.
type SyncState string

const (
	SyncStateNotSynced SyncState = "NOT_SYNCED"
	SyncStateSynced    SyncState = "SYNCED"
	SyncStateDrifted   SyncState = "DRIFTED"
	SyncStatePending   SyncState = "PENDING"
	SyncStateError     SyncState = "ERROR"
	SyncStateInvalid   SyncState = "INVALID"
)

func (SyncState) Values() []string {
	return []string{
		string(SyncStateNotSynced),
		string(SyncStateSynced),
		string(SyncStateDrifted),
		string(SyncStatePending),
		string(SyncStateError),
	}
}

func (s SyncState) String() string { return string(s) }

func ToSyncState(s string) SyncState {
	switch strings.ToUpper(s) {
	case SyncStateNotSynced.String():
		return SyncStateNotSynced
	case SyncStateSynced.String():
		return SyncStateSynced
	case SyncStateDrifted.String():
		return SyncStateDrifted
	case SyncStatePending.String():
		return SyncStatePending
	case SyncStateError.String():
		return SyncStateError
	default:
		return SyncStateInvalid
	}
}

func (s SyncState) MarshalGQL(out io.Writer) {
	_, _ = out.Write([]byte(`"` + s.String() + `"`))
}

func (s *SyncState) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for SyncState, got: %T", v) //nolint:err113
	}
	*s = SyncState(str)
	return nil
}

func (s SyncState) ToPtr() *SyncState { return &s }
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/enums/ -run TestSyncState -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/enums/sync_state.go pkg/enums/sync_state_test.go
git commit -m "feat(enums): add SyncState for HR sync status tracking"
```

---

## Task 4: Extend `People` schema with new fields + vendor edge

**Context:** This is the core schema change. It adds 8 new fields (`worker_type`, `source`, `external_id`, `vendor_id`, `contract_end_date`, `last_synced_at`, `sync_state`, `sync_error`) and one unique FK edge to the existing `Vendor` schema. Follow the existing edge pattern used by `account.go` for its `people_id` FK (search for `uniqueEdgeFrom` in that file).

**Important:** Do NOT add any hooks that reference generated types in this task — that triggers the bootstrap circular-dep (see project memory "Code Generation Bootstrap Problem"). Pure field additions + an edge are safe.

**Files:**
- Modify: `internal/store/schema/people.go`
- Modify: `internal/store/schema/vendor.go` (add inverse edge)

- [ ] **Step 1: Add the fields to `People.Fields()`**

Open `internal/store/schema/people.go`. In the `Fields()` method, add the following fields after the existing `employment_status` field:

```go
field.Enum("worker_type").
	GoType(enums.WorkerType("")).
	Default(enums.WorkerTypeEmployee.String()).
	Annotations(entgql.OrderField("WORKER_TYPE")).
	Comment("Classification of the workforce entry: employee, contractor, freelancer, intern, or temp. Drives workflow template selection and HR sync scoping."),
field.Enum("source").
	GoType(enums.PeopleSource("")).
	Default(enums.PeopleSourceManual.String()).
	Annotations(entgql.OrderField("SOURCE")).
	Comment("Where this People row was created: Personio, CSV upload, manual entry, or vendor onboarding."),
field.String("external_id").
	Optional().
	Comment("ID of this person in the source HR system, unique per (source, space_id). Nil for manual entries."),
field.String("vendor_id").
	Optional().
	Comment("For external workers: the vendor company they were hired through. Nil for employees."),
field.Time("contract_end_date").
	Optional().
	Comment("For external workers: the planned end date of the engagement. Used to auto-trigger offboarding in Block 2."),
field.Time("last_synced_at").
	Optional().
	Comment("Timestamp of the most recent HR sync that touched this row. Populated by Plan 1.2."),
field.Enum("sync_state").
	GoType(enums.SyncState("")).
	Default(enums.SyncStateNotSynced.String()).
	Annotations(entgql.OrderField("SYNC_STATE")).
	Comment("Relationship of this row to its HR source. NOT_SYNCED for manual rows."),
field.Text("sync_error").
	Optional().
	Comment("Last sync error message, if any. Populated by Plan 1.2."),
```

- [ ] **Step 2: Update the unique index on `external_id`**

Add a new unique index scoped by `(space_id, source, external_id)` so HR sync can reliably upsert. Replace the existing `Indexes()` method in `people.go`:

```go
func (People) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("space_id", "email").
			Unique(),
		index.Fields("space_id", "source", "external_id").
			Unique().
			Annotations(entsql.IndexWhere("external_id IS NOT NULL AND deleted_at IS NULL")),
	}
}
```

Add the import for `entsql` at the top:

```go
"entgo.io/ent/dialect/entsql"
```

- [ ] **Step 3: Add the `vendor` edge**

In `People.Edges()`, add a new edge right after the existing ones:

```go
uniqueEdgeFrom(&edgeDefinition{
	fromSchema: s,
	edgeSchema: Vendor{},
	field:      "vendor_id",
}),
```

The existing edge helpers are in `internal/store/schema/zzz_edge.go`. The pattern matches what `Account{}` does with `People{}` via `people_id` (see `account.go:121-126` for reference).

- [ ] **Step 4: Add the inverse edge on `Vendor`**

Open `internal/store/schema/vendor.go`. In `Vendor.Edges()`, add the inverse edge (scan the existing file to find the right location among the current edges):

```go
defaultEdgeToWithPagination(schema, People{}),
```

This mirrors how `BusinessUnit` exposes its People edge in `business_unit.go:89`.

- [ ] **Step 5: Regenerate Ent code**

Run from the `kopexa` repo root:

```bash
task gen:ent
```

Expected: ent code regenerates without errors. If you see "undefined: enums.WorkerType" errors it means Task 1-3 weren't pushed into the same module — re-check the enum files.

**If the generator fails** with a bootstrap issue (hooks referencing generated types that don't exist yet), it's the known project memory issue — but for this plan there are no hooks touching the new fields, so this should not happen. If it does, the fix is to temporarily move any affected hook file to `.tmp`, regenerate, then restore.

- [ ] **Step 6: Verify the schema compiled**

Run:

```bash
go build ./internal/store/ent/... && echo OK
```

Expected: `OK`.

- [ ] **Step 7: Commit the schema change**

```bash
git add internal/store/schema/people.go internal/store/schema/vendor.go internal/store/ent/
git commit -m "feat(people): add worker_type, source, vendor link and sync fields"
```

---

## Task 5: Create the database migration

**Files:**
- Create: `db/migrations/**` (Atlas-generated)
- Create: `db/migrations-goose-postgres/**` (generator-created)

- [ ] **Step 1: Generate the migration**

Run:

```bash
./kopexa-gen migrate --name add_people_worker_type
```

Expected: new files appear in `db/migrations/` and `db/migrations-goose-postgres/`.

- [ ] **Step 2: Inspect the generated SQL**

Open the newly created files in `db/migrations/`. Verify that the migration:
- Adds `worker_type` column with default `'EMPLOYEE'`
- Adds `source` column with default `'MANUAL'`
- Adds `external_id`, `vendor_id`, `contract_end_date`, `last_synced_at`, `sync_error` columns as nullable
- Adds `sync_state` column with default `'NOT_SYNCED'`
- Adds a new unique index on `(space_id, source, external_id) WHERE external_id IS NOT NULL AND deleted_at IS NULL`
- Adds a foreign key from `people.vendor_id → vendors.id` (or whatever the vendor table is called; check the existing migration structure)

If something looks wrong, fix the schema and regenerate (`task gen:ent` then `./kopexa-gen migrate --name add_people_worker_type` again — delete the old migration file first).

- [ ] **Step 3: Apply the migration locally**

```bash
docker-compose up -d db
./kopexa serve &
# Server should auto-run pending migrations on boot. Watch logs for
# "migrated to version ..." from the atlas runner.
# Stop the server once confirmed.
```

Alternatively if there is a direct migrate command in the `Taskfile`:

```bash
task db:migrate
```

- [ ] **Step 4: Verify the columns exist**

```bash
docker-compose exec db psql -U kopexa -d kopexa -c "\\d people" | grep -E "worker_type|source|vendor_id|contract_end_date|sync_state"
```

Expected: all five columns listed.

- [ ] **Step 5: Commit the migration**

```bash
git add db/migrations/ db/migrations-goose-postgres/
git commit -m "chore(db): migrate people to support worker_type and vendor link"
```

---

## Task 6: Regenerate GraphQL + verify schema

**Files:**
- Modify: `internal/channels/graphapi/**` (generated)

- [ ] **Step 1: Regenerate GraphQL**

```bash
task gen:gql
```

Expected: `internal/channels/graphapi/generated/` updates. `people.generated.go` now exposes the new fields in `People`, `CreatePeopleInput`, `UpdatePeopleInput`, `PeopleWhereInput`.

**Known expected error:** `graphapi/generate` always shows "main undeclared" during `go build ./...` — this is fine (project memory).

- [ ] **Step 2: Verify the generated model**

```bash
grep -n "WorkerType\|VendorID\|ContractEndDate\|SyncState" internal/channels/graphapi/generated/people.generated.go | head -20
```

Expected: hits for each field in `People`, `CreatePeopleInput`, `UpdatePeopleInput`, and `PeopleWhereInput`.

- [ ] **Step 3: Build the whole backend**

```bash
go build ./... 2>&1 | grep -v "graphapi/generate.*main undeclared"
```

Expected: no other errors.

- [ ] **Step 4: Commit**

```bash
git add internal/channels/graphapi/
git commit -m "chore(gqlgen): regenerate people types for new fields"
```

---

## Task 7: Backend integration test — create employee (default)

**Why:** Prove that creating a `People` row without specifying `worker_type` still works and defaults to `EMPLOYEE`.

**Files:**
- Modify: `internal/channels/graphapi/people_test.go` (add new test; create file if it does not exist yet)

- [ ] **Step 1: Find existing people test patterns**

```bash
grep -rln "CreatePeople" internal/channels/graphapi/*_test.go internal/store/ent/people_test.go 2>/dev/null
```

If `internal/channels/graphapi/people_test.go` exists, extend it. Otherwise find any `_test.go` in that package that sets up the test client and model the test after it (e.g. `business_unit_test.go`).

- [ ] **Step 2: Write the failing test**

Add to `internal/channels/graphapi/people_test.go` (create the file if needed, using the same `package graphapi_test` header and test-setup helpers as the neighboring `*_test.go` files in that directory):

```go
func TestCreatePeople_DefaultsToEmployee(t *testing.T) {
	ctx, cli := setupTestClient(t) // use whatever helper the neighboring tests use

	space := newTestSpace(t, ctx, cli)

	got, err := cli.People.Create().
		SetFirstName("Ada").
		SetLastName("Lovelace").
		SetEmail("ada@example.com").
		SetSpaceID(space.ID).
		Save(ctx)
	require.NoError(t, err)

	assert.Equal(t, enums.WorkerTypeEmployee, got.WorkerType)
	assert.Equal(t, enums.PeopleSourceManual, got.Source)
	assert.Equal(t, enums.SyncStateNotSynced, got.SyncState)
	assert.Empty(t, got.ExternalID)
	assert.Empty(t, got.VendorID)
}
```

*(The `setupTestClient` / `newTestSpace` names are placeholders — use whatever helpers the neighboring tests use. If you're unsure, run `grep -l "setupTestClient\|NewTestClient\|TestSpace" internal/channels/graphapi/*_test.go` to find the pattern.)*

- [ ] **Step 3: Run the test to verify it fails**

```bash
go test ./internal/channels/graphapi/ -run TestCreatePeople_DefaultsToEmployee -v
```

Expected: compile passes (the fields exist from Task 6), test passes — because the defaults in the schema produce exactly these values. *If it fails*, verify Task 4's defaults were correctly set.

- [ ] **Step 4: Commit**

```bash
git add internal/channels/graphapi/people_test.go
git commit -m "test(people): verify employee defaults on creation"
```

---

## Task 8: Backend integration test — create contractor with vendor + contract end date

**Why:** Prove that the external-workforce path works end-to-end: vendor FK resolves, contract_end_date stores, sync defaults stay correct.

**Files:**
- Modify: `internal/channels/graphapi/people_test.go`

- [ ] **Step 1: Write the failing test**

Add to `people_test.go`:

```go
func TestCreatePeople_Contractor_WithVendorAndContractEnd(t *testing.T) {
	ctx, cli := setupTestClient(t)

	space := newTestSpace(t, ctx, cli)

	vendor, err := cli.Vendor.Create().
		SetName("Acme Consulting GmbH").
		SetSpaceID(space.ID).
		Save(ctx)
	require.NoError(t, err)

	endDate := time.Now().AddDate(0, 6, 0)

	got, err := cli.People.Create().
		SetFirstName("Grace").
		SetLastName("Hopper").
		SetEmail("grace@acme-consulting.example").
		SetSpaceID(space.ID).
		SetWorkerType(enums.WorkerTypeContractor).
		SetVendorID(vendor.ID).
		SetContractEndDate(endDate).
		Save(ctx)
	require.NoError(t, err)

	assert.Equal(t, enums.WorkerTypeContractor, got.WorkerType)
	assert.Equal(t, vendor.ID, got.VendorID)
	assert.WithinDuration(t, endDate, got.ContractEndDate, time.Second)

	// The edge resolves back to the vendor.
	linkedVendor, err := got.QueryVendor().Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, vendor.ID, linkedVendor.ID)
}
```

- [ ] **Step 2: Run the test to verify it passes**

```bash
go test ./internal/channels/graphapi/ -run TestCreatePeople_Contractor -v
```

Expected: PASS.

*If the test fails at `.QueryVendor()` with "undefined field", Task 4 Step 3 missed the edge. Fix it, re-run gen:ent, retry.*

- [ ] **Step 3: Commit**

```bash
git add internal/channels/graphapi/people_test.go
git commit -m "test(people): verify contractor creation with vendor link"
```

---

## Task 9: Backend integration test — filter roster by worker_type

**Why:** The admin UI relies on `PeopleWhereInput.workerType` filtering working through the generated GraphQL `where` clause. Prove it.

**Files:**
- Modify: `internal/channels/graphapi/people_test.go`

- [ ] **Step 1: Write the failing test**

Add to `people_test.go`:

```go
func TestPeople_FilterByWorkerType(t *testing.T) {
	ctx, cli := setupTestClient(t)
	space := newTestSpace(t, ctx, cli)

	_, err := cli.People.Create().SetFirstName("Emp").SetLastName("A").SetSpaceID(space.ID).Save(ctx)
	require.NoError(t, err)
	_, err = cli.People.Create().SetFirstName("Con").SetLastName("B").SetSpaceID(space.ID).
		SetWorkerType(enums.WorkerTypeContractor).Save(ctx)
	require.NoError(t, err)
	_, err = cli.People.Create().SetFirstName("Free").SetLastName("C").SetSpaceID(space.ID).
		SetWorkerType(enums.WorkerTypeFreelancer).Save(ctx)
	require.NoError(t, err)

	externals, err := cli.People.Query().
		Where(people.WorkerTypeIn(enums.WorkerTypeContractor, enums.WorkerTypeFreelancer)).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, externals, 2)

	employees, err := cli.People.Query().
		Where(people.WorkerTypeEQ(enums.WorkerTypeEmployee)).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, employees, 1)
}
```

- [ ] **Step 2: Run**

```bash
go test ./internal/channels/graphapi/ -run TestPeople_FilterByWorkerType -v
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/channels/graphapi/people_test.go
git commit -m "test(people): verify worker_type filtering on list query"
```

---

## Task 10: Add German + English i18n keys

**Files:**
- Modify: `messages/en/people.json` (create/extend keys)
- Modify: `messages/de/people.json` (create/extend keys)

The exact file layout follows existing keys. Search `messages/en/people.json` for `employment_status` to see the conventions.

- [ ] **Step 1: Add English keys**

Open `messages/en/people.json` in the **frontend** repo (`kopexa-frontend`). Add inside the top-level `people` object:

```json
"worker_type": "Worker type",
"worker_type_employee": "Employee",
"worker_type_contractor": "Contractor",
"worker_type_freelancer": "Freelancer",
"worker_type_intern": "Intern",
"worker_type_temp": "Temporary worker",
"vendor": "Vendor",
"vendor_placeholder": "Select the vendor this contractor was hired through",
"vendor_description": "Only shown for external workers. Link to the supplier company.",
"contract_end_date": "Contract end date",
"contract_end_date_description": "Planned end of the engagement. Offboarding fires automatically on this date (once Block 2 ships).",
"filter_worker_type": "Worker type",
"filter_worker_type_all": "All",
"filter_worker_type_employees_only": "Employees",
"filter_worker_type_externals_only": "Externals"
```

- [ ] **Step 2: Add German keys (du-form, never Sie)**

Open `messages/de/people.json`. Add:

```json
"worker_type": "Beschäftigungsart",
"worker_type_employee": "Mitarbeiter",
"worker_type_contractor": "Auftragnehmer",
"worker_type_freelancer": "Freelancer",
"worker_type_intern": "Praktikant",
"worker_type_temp": "Aushilfe",
"vendor": "Lieferant",
"vendor_placeholder": "Wähle den Lieferanten, über den der Auftragnehmer beschäftigt ist",
"vendor_description": "Nur für externe Kräfte. Verknüpfung zur beauftragten Firma.",
"contract_end_date": "Vertragsende",
"contract_end_date_description": "Geplantes Ende der Beschäftigung. Das Offboarding startet automatisch an diesem Datum (sobald Block 2 ausgeliefert ist).",
"filter_worker_type": "Beschäftigungsart",
"filter_worker_type_all": "Alle",
"filter_worker_type_employees_only": "Mitarbeiter",
"filter_worker_type_externals_only": "Externe"
```

- [ ] **Step 3: Commit**

```bash
cd /Users/juliankoehn/workbox/github.com/kopexa-grc/kopexa-frontend
git add messages/en/people.json messages/de/people.json
git commit -m "i18n(people): add worker_type, vendor and contract_end keys"
```

---

## Task 11: Regenerate frontend GraphQL types

**Why:** The new backend fields need to flow into generated TypeScript types and hooks before the form can consume them.

**Files:**
- Modify: `src/generated/**` (all regenerated)

- [ ] **Step 1: Make sure the backend is running with the new schema**

Your running backend needs to serve the updated GraphQL schema. If it isn't, start it:

```bash
cd /Users/juliankoehn/workbox/github.com/kopexa-grc/kopexa
./kopexa serve &
```

(Or `docker-compose up -d` + whatever runs the server locally.)

- [ ] **Step 2: Regenerate the frontend types**

```bash
cd /Users/juliankoehn/workbox/github.com/kopexa-grc/kopexa-frontend
pnpm run codegen   # or whatever the project uses — check package.json scripts
```

Expected: `src/generated/graphql.ts` (or equivalent) now exports `PeopleWorkerType`, `PeopleSource`, `PeopleSyncState` enums, plus the new fields on `People`, `CreatePeopleInput`, `UpdatePeopleInput`, `PeopleWhereInput`.

- [ ] **Step 3: Verify the enum is exported**

```bash
grep -n "PeopleWorkerType\|workerType" src/generated/graphql.ts | head
```

Expected: enum + field references.

- [ ] **Step 4: Commit**

```bash
git add src/generated/
git commit -m "chore(codegen): regenerate people types with worker_type"
```

---

## Task 12: Extend the zod validation schema

**Files:**
- Modify: `src/modules/people/validation/index.ts`

- [ ] **Step 1: Update the schema**

Replace the contents of `src/modules/people/validation/index.ts` with:

```ts
import { z } from "zod";
import {
	PeopleEmploymentStatus,
	PeopleWorkerType,
} from "@/generated/graphql";

const externalWorkerTypes = [
	PeopleWorkerType.Contractor,
	PeopleWorkerType.Freelancer,
	PeopleWorkerType.Intern,
	PeopleWorkerType.Temp,
] as const;

export const isExternalWorkerType = (
	value: PeopleWorkerType | null | undefined,
): boolean => {
	if (!value) return false;
	return (externalWorkerTypes as readonly PeopleWorkerType[]).includes(value);
};

export const peopleSchema = z
	.object({
		firstName: z.string().min(1).max(255),
		lastName: z.string().min(1).max(255),
		email: z.string().email().optional(),
		startDate: z.date().nullish(),
		endDate: z.date().nullish(),
		jobTitle: z.string().nullish(),
		location: z.string().nullish(),
		employmentStatus: z.nativeEnum(PeopleEmploymentStatus).nullish(),
		workerType: z
			.nativeEnum(PeopleWorkerType)
			.default(PeopleWorkerType.Employee),
		vendorId: z.string().nullish(),
		contractEndDate: z.date().nullish(),
	})
	.superRefine((data, ctx) => {
		if (isExternalWorkerType(data.workerType) && !data.vendorId) {
			ctx.addIssue({
				code: z.ZodIssueCode.custom,
				message: "vendor_required_for_externals",
				path: ["vendorId"],
			});
		}
	});

export type PeopleFormValues = z.infer<typeof peopleSchema>;

export const peopleUpdateSchema = peopleSchema.innerType().partial();

export type PeopleUpdateFormValues = z.infer<typeof peopleUpdateSchema>;
```

**Note on `innerType().partial()`:** `superRefine` wraps the schema in an effect that breaks `.partial()` directly. Calling `.innerType()` first unwraps it; then `.partial()` is applied to the underlying object. This preserves the previous update-schema behavior.

- [ ] **Step 2: Build to check for type errors**

```bash
pnpm run type-check   # or tsc --noEmit, whatever the project uses
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add src/modules/people/validation/index.ts
git commit -m "feat(people): extend validation schema with worker_type fields"
```

---

## Task 13: Create `<WorkerTypeField />` reusable component

**Why:** Used in both the create and edit forms. Extract to avoid duplication.

**Files:**
- Create: `src/modules/people/components/form/worker-type-field.tsx`

- [ ] **Step 1: Create the component**

```tsx
"use client";

import { useFormContext } from "react-hook-form";
import { useTranslations } from "next-intl";
import { PeopleWorkerType } from "@/generated/graphql";
import { SelectField } from "@/modules/common/forms/select-field";
import type { PeopleFormValues } from "../../validation";

const WORKER_TYPE_ORDER: PeopleWorkerType[] = [
	PeopleWorkerType.Employee,
	PeopleWorkerType.Contractor,
	PeopleWorkerType.Freelancer,
	PeopleWorkerType.Intern,
	PeopleWorkerType.Temp,
];

export function WorkerTypeField() {
	const t = useTranslations("people");
	const { control } = useFormContext<PeopleFormValues>();

	const options = WORKER_TYPE_ORDER.map((value) => ({
		value,
		label: t(`worker_type_${value.toLowerCase()}`),
	}));

	return (
		<SelectField
			control={control}
			name="workerType"
			label={t("worker_type")}
			options={options}
		/>
	);
}
```

*(If the project already has a different `SelectField` abstraction, use whichever form-integrated select the rest of `src/modules/people/` uses. Run `grep -l "SelectField\|<Select" src/modules/people/components/**/*.tsx` to find examples. The important part is the options map + the zod binding.)*

- [ ] **Step 2: Commit**

```bash
git add src/modules/people/components/form/worker-type-field.tsx
git commit -m "feat(people): add reusable WorkerTypeField component"
```

---

## Task 14: Wire `WorkerTypeField` + conditional fields into the create form

**Files:**
- Modify: `src/modules/people/components/create-form/people-create-form.tsx`

- [ ] **Step 1: Update the form**

Modify `people-create-form.tsx` to:
1. Import `WorkerTypeField`, `PeopleWorkerType` from generated, `isExternalWorkerType` from validation, a `VendorPicker` (look for an existing vendor select component — `grep -rln "vendor" src/modules/common/ src/modules/vendor/`), and `DateField`.
2. Add `workerType: PeopleWorkerType.Employee` to `defaultValues`.
3. Watch the `workerType` field using `form.watch("workerType")`.
4. Render conditional fields.

Key changes to the JSX (after the existing `DateField name="startDate"`):

```tsx
<WorkerTypeField />

{isExternalWorkerType(form.watch("workerType")) && (
	<>
		<VendorPicker
			name="vendorId"
			label={common("vendor")}
			description={t("vendor_description")}
		/>
		<DateField
			name="contractEndDate"
			label={t("contract_end_date")}
			description={t("contract_end_date_description")}
			className="w-full"
		/>
	</>
)}
```

And pass the new fields through in the submit handler:

```tsx
const onSubmit = (data: PeopleFormValues) => {
	toast.promise(
		mutateAsync(
			{
				input: {
					...data,
					startDate: data.startDate?.toISOString(),
					endDate: data.endDate?.toISOString(),
					contractEndDate: data.contractEndDate?.toISOString(),
					parentID: spaceId,
					businessUnitIDs: defaultBusinessUnitId ? [defaultBusinessUnitId] : undefined,
				},
			},
			// ... onSuccess/onError unchanged
		),
		{
			loading: common("create_loading"),
			success: common("create_success"),
			error: common("create_error"),
		},
	);
};
```

Also update `defaultValues`:

```tsx
defaultValues: {
	firstName: "",
	lastName: "",
	email: "",
	startDate: new Date(),
	workerType: PeopleWorkerType.Employee,
},
```

And update the `t` translator at the top:

```tsx
const t = useTranslations("people");
```

- [ ] **Step 2: Build to verify types**

```bash
pnpm run type-check
```

Expected: no errors. If `VendorPicker` doesn't exist as named, find the right vendor-select helper and adjust — do not stub it.

- [ ] **Step 3: Run dev server and manually verify**

```bash
pnpm run dev
```

Open the app, navigate to People → Add person, check:
- Default worker type is "Employee"; vendor and contract-end-date are hidden
- Switching to "Contractor" reveals vendor picker + contract-end-date field
- Submitting as employee works
- Submitting as contractor without a vendor shows a validation error on the vendor field
- Submitting as contractor with a vendor + date persists; the new row appears in the list

- [ ] **Step 4: Commit**

```bash
git add src/modules/people/components/create-form/people-create-form.tsx
git commit -m "feat(people): support worker_type + vendor in create form"
```

---

## Task 15: Mirror the changes into the edit form

**Files:**
- Modify: `src/modules/people/components/edit-form/people-edit-form.tsx`

- [ ] **Step 1: Apply the same shape as Task 14**

Repeat the changes from Task 14 in `people-edit-form.tsx`:
- Add `WorkerTypeField`
- Add conditional vendor picker + contract-end-date
- Pass `workerType`, `vendorId`, `contractEndDate` through in the update mutation payload

Hydrate the form from the loaded `People` record so existing contractors show their vendor selection.

- [ ] **Step 2: Type check + manual verify**

```bash
pnpm run type-check && pnpm run dev
```

Open the detail page of a contractor created in Task 14, open Edit, verify fields are pre-filled, change the contract end date, save — the new value appears on the detail page.

- [ ] **Step 3: Commit**

```bash
git add src/modules/people/components/edit-form/people-edit-form.tsx
git commit -m "feat(people): support worker_type + vendor in edit form"
```

---

## Task 16: Add `workerType` column + filter to the roster

**Files:**
- Modify: `src/modules/people/components/table/table-config.tsx`
- Modify: `src/modules/people/components/table/people-table-toolbar.tsx`

- [ ] **Step 1: Add the column**

Open `table-config.tsx`. Find where existing columns like `employmentStatus` or `jobTitle` are declared. Add a new column after `employmentStatus`:

```tsx
{
	id: "workerType",
	header: t("worker_type"),
	accessorFn: (row) => row.workerType,
	cell: ({ row }) => (
		<Chip size="sm" variant="flat">
			{t(`worker_type_${row.original.workerType.toLowerCase()}`)}
		</Chip>
	),
}
```

*(Follow whatever column-config shape the file actually uses — tanstack-table `ColumnDef`, custom wrapper, etc. Match the style of neighboring columns.)*

- [ ] **Step 2: Add the toolbar filter**

Open `people-table-toolbar.tsx`. Find the existing filters (employment status, business unit). Add a worker-type filter that emits a `PeopleWhereInput.workerTypeIn` clause.

Follow the exact pattern of the existing employment-status filter — search for `employmentStatus` in that file. The filter selector should offer:
- **All** (no filter)
- **Employees** (`workerTypeIn: [EMPLOYEE]`)
- **Externals** (`workerTypeIn: [CONTRACTOR, FREELANCER, INTERN, TEMP]`)
- Individual types (one per enum value)

Use the i18n keys added in Task 10.

- [ ] **Step 3: Build + manual verify**

```bash
pnpm run type-check && pnpm run dev
```

Open the people list:
- New "Worker type" column appears with a badge per row
- Filter dropdown lets you narrow to just externals, just employees, or one specific type
- Results update live

- [ ] **Step 4: Commit**

```bash
git add src/modules/people/components/table/
git commit -m "feat(people): add worker_type column and filter to roster"
```

---

## Task 17: End-to-end sanity check + push

- [ ] **Step 1: Full backend test sweep**

```bash
cd /Users/juliankoehn/workbox/github.com/kopexa-grc/kopexa
go test ./pkg/enums/ ./internal/channels/graphapi/ ./internal/store/... -count=1
```

Expected: all green.

- [ ] **Step 2: Lint**

```bash
make lint
```

Expected: no errors.

- [ ] **Step 3: Frontend typecheck + build**

```bash
cd /Users/juliankoehn/workbox/github.com/kopexa-grc/kopexa-frontend
pnpm run type-check
pnpm run build
```

Expected: no errors.

- [ ] **Step 4: Manual golden-path check in browser**

Start both servers. Walk through:
1. Create an employee — defaults to Employee, no extra fields shown. ✓
2. Create a contractor — switches to show vendor + contract end date. ✓
3. Submit without vendor → validation error on the vendor field. ✓
4. Submit with vendor + future date → success, new row appears. ✓
5. Filter roster to "Externals only" → only the contractor shows. ✓
6. Filter to "Employees only" → contractor hidden. ✓
7. Open contractor detail page → worker type + vendor + contract end visible. ✓
8. Edit contractor → vendor change persists; reload shows new vendor. ✓

- [ ] **Step 5: Push both branches**

```bash
cd /Users/juliankoehn/workbox/github.com/kopexa-grc/kopexa
git push -u origin feat/people-platform

cd /Users/juliankoehn/workbox/github.com/kopexa-grc/kopexa-frontend
git push -u origin feat/people-platform
```

- [ ] **Step 6: Open PRs**

For each repo, open a PR titled `feat(people): worker_type + external workforce support` with a body that links to the spec at `comms/docs/superpowers/specs/2026-04-14-people-platform-roadmap.md` (Block 1) and this plan (Plan 1.1).

---

## Plan Self-Review

**Spec coverage check** — each Block 1 spec deliverable mapped to a task or explicit deferral:

| Spec deliverable | Where |
|---|---|
| Schema deltas (`external_id`, `source`, `worker_type`, `vendor_id`, `contract_end_date`, `last_synced_at`, `sync_state`, `sync_error`) | Task 4 |
| HR connector framework | **Deferred to Plan 1.2** (header) |
| `hr_field_mapping` schema | **Deferred to Plan 1.2** |
| Sync scoping rule | **Deferred to Plan 1.2** |
| River sync job | **Deferred to Plan 1.2** |
| Admin UI connector/mapping/history | **Deferred to Plan 1.2** |
| Add Person form with worker-type selector | Tasks 13-14 |
| Roster view with worker-type filter | Task 16 |
| River cron sweep (contract_end → terminated event) | **Deferred to Block 2** (requires `people_employment_event`) |
| Default mapping config (source_wins / kopexa_wins) | **Deferred to Plan 1.2** |

All schema fields required by Plan 1.2 are present after Plan 1.1 ships. Plan 1.2 is pure code, no migration.

**Placeholder scan:** no `TBD` / `TODO` / "add appropriate error handling". All step code is either complete Go/TSX or explicit pattern references with search commands.

**Type consistency check:** `WorkerType` / `PeopleSource` / `SyncState` enum symbols are used identically in Tasks 1-3 (definition), Task 4 (schema), Tasks 7-9 (tests), Tasks 12-14 (frontend). The field `workerType` / `vendorId` / `contractEndDate` naming is the same in zod schema, form components, and mutation payload.

**Bootstrap caveat:** Task 4 explicitly warns about the known code-gen bootstrap problem and confirms Plan 1.1 does not trip it (no hooks reference the new fields).

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-04-14-people-core-1.1.md`. Two execution options:

1. **Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration.
2. **Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints.

Which approach?
