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

**Workflow conventions for this plan:**
- **Migrations**: Do NOT generate database migrations during the plan. The dev `kopexa serve` auto-migrate picks up schema changes on boot. ONE consolidated migration is generated at the very end of the branch, exclusively via `./kopexa-gen migrate`.
- **Generated-code tests**: Do NOT write Go tests that just verify ent / gqlgen-generated behavior (defaults, edge resolution, where-input filtering). Manual e2e in Task 17 covers the integration surface; hooks and custom logic get their own unit tests when they are introduced.

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
| `db/migrations/**` | **Create at end only** | One consolidated migration via `./kopexa-gen migrate` as the final task. Skipped for every intermediate task. |
| `internal/channels/graphapi/**` | Regenerate | `task gen:gql` |
| `src/modules/people/validation/index.ts` | Modify | Add `workerType`, `vendorId`, `contractEndDate` to zod schema |
| `src/modules/people/components/create-form/people-create-form.tsx` | Modify | Add `WorkerTypeField` + conditional vendor picker + conditional `contractEndDate` |
| `src/modules/people/components/edit-form/people-edit-form.tsx` | Modify | Same fields, editable |
| `src/modules/people/components/form/worker-type-field.tsx` | Create | Reusable `<WorkerTypeField />` (Select with enum values + i18n labels) |
| `src/modules/people/components/table/table-config.tsx` | Modify | Add `workerType` column with badge rendering |
| `src/modules/people/hooks/use-people-filter-where.ts` | Create | nuqs-backed filter state + `convertFiltersToWhere` mapping for People (mirrors `use-vendor-filter-where.ts`) |
| `src/modules/people/components/table/people-table-toolbar.tsx` | Rewrite | Replace bare `SearchInput` with `SearchFilterBar` (from `@kopexa/sight`), pulling state from `use-people-filter-where` |
| `messages/en/people.json` | Modify | i18n keys for worker types + new form labels + filter labels |
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

## Task 5: ~~Create the database migration~~ — DEFERRED TO END OF PLAN

**Status:** Skipped at this position. Migrations are NOT generated per task during the plan; the dev server's auto-migrate picks up schema changes on `kopexa serve` boot. ONE consolidated migration is generated in the very last task of the branch via `./kopexa-gen migrate`, never via any other tool.

If you are an implementer and you find yourself about to run `./kopexa-gen migrate` here, **stop and skip this task**. Move on to Task 6.

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

## Tasks 7-9: ~~Backend integration tests for People schema~~ — SKIPPED

**Status:** Skipped. These tests would only verify ent/gqlgen-generated behavior (default values, FK edge resolution, where-input filtering) — nothing the team wrote and nothing that would break on its own. Generated-code coverage is the wrong place to spend test time. The end-of-plan manual e2e check (Task 17) is the canonical verification of the integration surface; hooks and custom logic get unit tests when they are introduced (none in this plan).

If you are an implementer reading this, **skip Tasks 7-9 entirely** and proceed to Task 10.

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

## Task 16: Migrate People roster to `SearchFilterBar` + `workerType` column

**Goal:** Replace the bare `SearchInput` in `people-table-toolbar.tsx` with the new `SearchFilterBar` pattern from `@kopexa/sight` (the same component that vendors and issues already use). At the same time, introduce the worker-type column and worker-type filter.

**Why the migration:** The People roster is currently the odd one out — vendors and issues both have a unified, URL-state-backed filter+search bar via `SearchFilterBar`. We want all roster surfaces to look and feel the same, and we want filter state to be shareable via URL.

**Reference implementation:** `src/modules/vendor/vendor-list/components/vendor-table-toolbar.tsx` and its sibling hook `src/modules/vendor/vendor-list/hooks/use-vendor-filter-where.ts`. Read both files first to understand the pattern — they are the canonical example.

**Files:**
- Modify: `src/modules/people/components/table/table-config.tsx` — add the `workerType` column with badge rendering
- Create: `src/modules/people/hooks/use-people-filter-where.ts` — nuqs-backed filter state + `convertFiltersToWhere` mapping for People
- Rewrite: `src/modules/people/components/table/people-table-toolbar.tsx` — replace bare `SearchInput` with `SearchFilterBar`
- Modify: wherever the People list page consumes the toolbar — wire it to read `whereCondition` from the new hook (search the people page tree for where `PeopleTableToolbar` is rendered and where the people list query gets its `where` input)

### Step 1: Add the `workerType` column to `table-config.tsx`

- [ ] Open `src/modules/people/components/table/table-config.tsx`. Find where existing columns like `employmentStatus` or `jobTitle` are declared. Add a new column right after `employmentStatus` (or wherever fits the visual order):

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

Match the column-config shape (tanstack `ColumnDef`, custom wrapper, etc.) used by neighboring columns in that file. Do not introduce a new convention.

### Step 2: Create `use-people-filter-where.ts`

- [ ] Create `src/modules/people/hooks/use-people-filter-where.ts`. Mirror the structure of `src/modules/vendor/vendor-list/hooks/use-vendor-filter-where.ts`. The shape should be:

```ts
"use client";

import type { FilterBarValue } from "@kopexa/sight";
import { parseAsJson, parseAsString, useQueryState } from "nuqs";
import { useCallback, useMemo } from "react";
import type {
	PeopleEmploymentStatus,
	PeopleSource,
	PeopleWhereInput,
	PeopleWorkerType,
} from "@/generated/graphql";
import {
	convertFiltersToWhere,
	type FilterToWhereConfig,
} from "@/lib/filter-bar";

const parseAsFilterValues = parseAsJson<FilterBarValue[]>((value) => {
	if (!Array.isArray(value)) return [];
	return value.map((filter) => {
		if (typeof filter !== "object" || filter === null) return filter;
		if (typeof filter.value === "string" && filter.value.startsWith("[")) {
			try {
				return { ...filter, value: JSON.parse(filter.value) };
			} catch {
				return filter;
			}
		}
		return filter;
	}) as FilterBarValue[];
});

const peopleFilterConfig: FilterToWhereConfig<PeopleWhereInput> = {
	fields: [
		{
			fieldId: "workerType",
			toWhere: (value) => {
				if (Array.isArray(value)) {
					return { workerTypeIn: value as PeopleWorkerType[] };
				}
				return { workerType: value as PeopleWorkerType };
			},
		},
		{
			fieldId: "employmentStatus",
			toWhere: (value) => ({
				employmentStatus: value as PeopleEmploymentStatus,
			}),
		},
		{
			fieldId: "source",
			toWhere: (value) => ({
				source: value as PeopleSource,
			}),
		},
	],
	searchToWhere: (search) => ({
		or: [
			{ firstNameContainsFold: search },
			{ lastNameContainsFold: search },
			{ emailContainsFold: search },
		],
	}),
};

export const usePeopleFilterWhere = (spaceId: string) => {
	const [filterValues, setFilterValuesRaw] = useQueryState(
		"filters",
		parseAsFilterValues,
	);
	const [search, setSearchRaw] = useQueryState("q", parseAsString);

	const setFilterValues = useCallback(
		(values: FilterBarValue[]) => {
			setFilterValuesRaw(values.length > 0 ? values : null);
		},
		[setFilterValuesRaw],
	);

	const setSearch = useCallback(
		(value: string) => {
			setSearchRaw(value || null);
		},
		[setSearchRaw],
	);

	const whereCondition = useMemo((): PeopleWhereInput => {
		const filterWhere = convertFiltersToWhere(
			filterValues ?? [],
			search,
			peopleFilterConfig,
		);
		const baseCondition: PeopleWhereInput = { spaceID: spaceId };
		if (Object.keys(filterWhere).length > 0) {
			return { ...baseCondition, ...filterWhere };
		}
		return baseCondition;
	}, [filterValues, search, spaceId]);

	return {
		filterValues: filterValues ?? [],
		setFilterValues,
		search: search ?? "",
		setSearch,
		whereCondition,
	};
};
```

**Verify the field/where names against the regenerated `src/generated/graphql.ts`** — the exact spelling of `PeopleWhereInput.workerType` / `workerTypeIn` / `firstNameContainsFold` may differ slightly. If you find different names, adjust accordingly. Do not invent names.

### Step 3: Rewrite `people-table-toolbar.tsx`

- [ ] Replace the entire body of `src/modules/people/components/table/people-table-toolbar.tsx` with a `SearchFilterBar`-based implementation modeled on `vendor-table-toolbar.tsx`:

```tsx
"use client";

import { type FilterBarFieldConfig, SearchFilterBar } from "@kopexa/sight";
import { BriefcaseIcon, UserCheckIcon, UsersIcon } from "lucide-react";
import { useTranslations } from "next-intl";
import { useMemo } from "react";
import {
	PeopleEmploymentStatus,
	PeopleWorkerType,
} from "@/generated/graphql";
import { enumValues } from "@/lib/utils";
import { usePeopleFilterWhere } from "../../hooks/use-people-filter-where";

type PeopleTableToolbarProps = {
	spaceId: string;
};

export const PeopleTableToolbar = ({ spaceId }: PeopleTableToolbarProps) => {
	const t = useTranslations("people");
	const tCommon = useTranslations("common");

	const { filterValues, setFilterValues, search, setSearch } =
		usePeopleFilterWhere(spaceId);

	const workerTypeOptions = useMemo(
		() =>
			enumValues(PeopleWorkerType).map((wt) => ({
				value: wt,
				label: t(`worker_type_${wt.toLowerCase()}`),
			})),
		[t],
	);

	const employmentStatusOptions = useMemo(
		() =>
			enumValues(PeopleEmploymentStatus).map((status) => ({
				value: status,
				label: t(`employment_status.${status}`),
			})),
		[t],
	);

	const fields: FilterBarFieldConfig[] = useMemo(
		() => [
			{
				id: "workerType",
				label: t("worker_type"),
				type: "select",
				icon: <UsersIcon className="size-4" />,
				operators: ["equals", "in"],
				options: workerTypeOptions,
			},
			{
				id: "employmentStatus",
				label: tCommon("status"),
				type: "select",
				icon: <UserCheckIcon className="size-4" />,
				operators: ["equals"],
				options: employmentStatusOptions,
			},
			{
				id: "source",
				label: t("source"),
				type: "select",
				icon: <BriefcaseIcon className="size-4" />,
				operators: ["equals"],
				options: [
					{ value: "PERSONIO", label: "Personio" },
					{ value: "CSV", label: "CSV" },
					{ value: "MANUAL", label: t("source_manual") },
					{ value: "VENDOR_ONBOARDING", label: t("source_vendor_onboarding") },
				],
			},
		],
		[t, tCommon, workerTypeOptions, employmentStatusOptions],
	);

	return (
		<SearchFilterBar
			fields={fields}
			filters={filterValues}
			onFiltersChange={setFilterValues}
			defaultSearch={search}
			onSearchValueChange={setSearch}
			searchDebounce={300}
			allowMultiple={false}
		/>
	);
};
```

If `enumValues` or `FilterBarFieldConfig` are imported from a slightly different path, follow whatever the vendor toolbar uses. Do not invent imports — confirm by reading the vendor file.

### Step 4: Wire the toolbar into the People list page

- [ ] Find where `PeopleTableToolbar` is currently rendered. From the existing minimal component, it likely takes no props. Find the parent (probably `src/modules/people/components/people-list-header.tsx` or the page at `src/app/(protected)/(space)/s/[spaceId]/(root)/people/page.tsx`). Update the parent so that:
  - It passes `spaceId` to `PeopleTableToolbar`
  - It calls `usePeopleFilterWhere(spaceId)` itself (or lifts the call to a higher component) and passes the resulting `whereCondition` into the People list query (`useGetPeoples` or similar)
  - The toolbar and the list share the same hook instance — either by lifting the hook into a shared parent or by re-calling it (the `nuqs` URL state means both calls return the same values)

If the People list query is already filtering by `spaceId`, replace that with the `whereCondition` returned from the hook (which already includes the `spaceID` filter).

### Step 5: Build + manual verify

- [ ] Run:

```bash
cd /Users/juliankoehn/workbox/github.com/kopexa-grc/kopexa-frontend
pnpm run type-check
```

Expected: no errors. If `enumValues` is missing or the `FilterBarFieldConfig` type isn't found, fix the imports — do not stub them.

Then start the dev server and walk through:
1. People list shows the new SearchFilterBar at the top (search input + add-filter button)
2. The "Worker type" column shows a badge per row
3. Click "+ Filter" → "Worker type" → select "Contractor" → only contractors show
4. Switch to "is in" → select multiple (Contractor, Freelancer) → both show
5. Search by name + filter combined → both apply
6. URL shows `?filters=...&q=...` — filters survive page reload
7. Other filters (employment status, source) apply in the same way

### Step 6: Commit

- [ ] 

```bash
git add src/modules/people/components/table/table-config.tsx \
        src/modules/people/components/table/people-table-toolbar.tsx \
        src/modules/people/hooks/use-people-filter-where.ts \
        src/modules/people/components/people-list-header.tsx \
        src/app/\(protected\)/\(space\)/s/\[spaceId\]/\(root\)/people/page.tsx
git commit -m "feat(people): migrate roster to SearchFilterBar + worker_type column"
```

(Adjust the file list to match what you actually touched.)

---

## Task 17: Generate consolidated migration, e2e sanity check, push

This is the **final task** of the plan and the only place a database migration is generated.

### Step 1: Generate the consolidated migration

- [ ] Make sure the dev `kopexa serve` has been booted at least once since the last schema change so the dev DB is in sync. Then:

```bash
cd /Users/juliankoehn/workbox/github.com/kopexa-grc/kopexa
./kopexa-gen migrate --name people_external_workforce
```

Expected: new SQL files in both `db/migrations/` and `db/migrations-goose-postgres/`, plus updated `atlas.sum`. **Use only `kopexa-gen`** — never call atlas/goose directly.

- [ ] Inspect the generated SQL — verify it contains all 8 new columns (`worker_type`, `source`, `external_id`, `vendor_id`, `contract_end_date`, `last_synced_at`, `sync_state`, `sync_error`), the partial unique index `(space_id, source, external_id) WHERE external_id IS NOT NULL AND deleted_at IS NULL`, and the FK from `people.vendor_id → vendors.id`.

- [ ] Commit just the migration:

```bash
git add db/migrations/ db/migrations-goose-postgres/
git commit -m "chore(db): migrate people for worker_type and vendor link"
```

### Step 2: Backend lint

- [ ] 

```bash
make lint
```

Expected: no errors.

### Step 3: Frontend typecheck + build

- [ ] 

```bash
cd /Users/juliankoehn/workbox/github.com/kopexa-grc/kopexa-frontend
pnpm run type-check
pnpm run build
```

Expected: no errors.

### Step 4: Manual golden-path check in browser

- [ ] Start both servers and walk through:
1. Create an employee → defaults to Employee, no extra fields shown
2. Create a contractor → vendor + contract end date appear conditionally
3. Submit without vendor → zod validation error on the vendor field
4. Submit with vendor + future date → success, new row appears in the list
5. New `SearchFilterBar` is visible at the top of the people list (search input + add-filter button)
6. New "Worker type" column shows a badge per row
7. Add filter "Worker type = Contractor" → only the contractor shows
8. Switch the operator to "is in" → select multiple types → all selected types show
9. Combine search + filter — both apply, URL shows `?filters=…&q=…`, reload preserves them
10. Open contractor detail page → worker type + vendor + contract end visible
11. Edit contractor → change vendor → save → detail page shows new vendor

- [ ] Note: do NOT run `go test ./internal/channels/graphapi/...` or `./internal/store/...` as part of this sweep — those are generated-code surfaces and we explicitly skipped per-task tests for them. The enum test suite is the only Go suite the plan adds:

```bash
go test ./pkg/enums/ -count=1
```

Expected: green.

### Step 5: Push both branches

- [ ] 

```bash
cd /Users/juliankoehn/workbox/github.com/kopexa-grc/kopexa
git push -u origin feat/people-platform

cd /Users/juliankoehn/workbox/github.com/kopexa-grc/kopexa-frontend
git push -u origin feat/people-platform
```

### Step 6: Open PRs

- [ ] For each repo, open a PR titled `feat(people): worker_type + external workforce support` with a body that links to the spec at `comms/docs/superpowers/specs/2026-04-14-people-platform-roadmap.md` (Block 1) and this plan (Plan 1.1).

---

## Plan Self-Review

**Spec coverage check** — each Block 1 spec deliverable mapped to a task or explicit deferral:

| Spec deliverable | Where |
|---|---|
| Schema deltas (`external_id`, `source`, `worker_type`, `vendor_id`, `contract_end_date`, `last_synced_at`, `sync_state`, `sync_error`) | Task 4 |
| Database migration (one consolidated SQL file via `kopexa-gen`) | Task 17 step 1 (final task only) |
| HR connector framework | **Deferred to Plan 1.2** (header) |
| `hr_field_mapping` schema | **Deferred to Plan 1.2** |
| Sync scoping rule | **Deferred to Plan 1.2** |
| River sync job | **Deferred to Plan 1.2** |
| Admin UI connector/mapping/history | **Deferred to Plan 1.2** |
| Add Person form with worker-type selector | Tasks 13-14 |
| Roster view with worker-type column + SearchFilterBar | Task 16 |
| River cron sweep (contract_end → terminated event) | **Deferred to Block 2** (requires `people_employment_event`) |
| Default mapping config (source_wins / kopexa_wins) | **Deferred to Plan 1.2** |

All schema fields required by Plan 1.2 are present after Plan 1.1 ships. Plan 1.2 is pure code, no migration.

**Placeholder scan:** no `TBD` / `TODO` / "add appropriate error handling". All step code is either complete Go/TSX or explicit pattern references with search commands.

**Type consistency check:** `WorkerType` / `PeopleSource` / `SyncState` enum symbols are used identically in Tasks 1-3 (definition) and Task 4 (schema). The field `workerType` / `vendorId` / `contractEndDate` naming is the same in zod schema, form components, mutation payload, and filter hook.

**Workflow conventions:** Tasks 5 and 7-9 from the original draft are explicitly skipped (migration deferred to Task 17; generated-code tests not worth writing). The plan documents this inline at each skipped section so an implementer reading sequentially does not get confused.

**Bootstrap caveat:** Task 4 explicitly warns about the known code-gen bootstrap problem and confirms Plan 1.1 does not trip it (no hooks reference the new fields).

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-04-14-people-core-1.1.md`. Two execution options:

1. **Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration.
2. **Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints.

Which approach?
