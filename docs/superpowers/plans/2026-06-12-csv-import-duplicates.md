# CSV Import Duplicate Handling Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Configurable asset CSV import (CREATE/SKIP/UPSERT) with row-level result reporting, column-mapping overrides and parenting via `parent_external_id` — fixes kopexa-grc/kopexa#878.

**Architecture:** A generic import pipeline (`bulk_import.go`) in the graphapi layer owns parse → normalize → prefetch → classify → execute → post → report. Entity-specific logic lives in a small adapter; assets are the first adapter. Updates go through normal ent `Update().SetInput()` so hooks and privacy policies apply. The frontend import drawer becomes a 4-step wizard (file → mapping → options → result).

**Tech Stack:** Go (ent, gqlgen, csvutil), GraphQL, Next.js/React (sight Drawer), next-intl.

**Spec:** `docs/superpowers/specs/2026-06-12-csv-import-duplicates-design.md`

**Repos:**
- Backend: `/Users/juliankoehn/workbox/github.com/kopexa-grc/kopexa` (branch work directly on `main` per repo habit)
- Frontend: `/Users/juliankoehn/workbox/github.com/kopexa-grc/kopexa-frontend`

**Conventions that bind every task:**
- During dev rely on `kopexa serve` auto-migrate. Generate ONE consolidated migration only in Task 9 — never per task.
- Don't write tests for generated ent/gqlgen code — only for pipeline logic, hooks, builders.
- `task gen:gql` regenerates resolvers; the `graphapi/generate` "main undeclared" error during `go build ./...` is expected noise.
- All code/comments in English. German UI strings use informal "du".

---

### Task 1: Spike — csvutil empty-cell behavior (backend)

The whole normalization design depends on what csvutil does with an empty CSV cell for a `*string` field.

**Files:**
- Test: `internal/channels/graphapi/bulk_import_test.go` (new)

- [ ] **Step 1: Write the probe test**

```go
// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package graphapi

import (
	"bytes"
	"mime/multipart"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/kopexa-grc/kopexa/internal/store/ent"
	"github.com/stretchr/testify/require"
)

func uploadFromString(t *testing.T, csv string) graphql.Upload {
	t.Helper()
	return graphql.Upload{
		File:     bytes.NewReader([]byte(csv)),
		Filename: "test.csv",
		Size:     int64(len(csv)),
	}
}

// Documents how csvutil hands empty cells to *string fields. The
// normalization in Task 4 depends on this behavior.
func TestUnmarshalBulkData_EmptyCells(t *testing.T) {
	csv := "Name,InternalID,ExternalID\nServer A,,ext-1\n"
	rows, err := unmarshalBulkData[ent.CreateAssetInput](uploadFromString(t, csv))
	require.NoError(t, err)
	require.Len(t, rows, 1)

	// Log the actual behavior; the assertion below is adjusted to it.
	t.Logf("InternalID: %#v", rows[0].InternalID)

	// EXPECTED (verify on first run): csvutil allocates a pointer to ""
	// for present-but-empty cells. If it turns out to be nil instead,
	// flip this assertion AND note it in the Task 4 normalize comment.
	if rows[0].InternalID != nil {
		require.Equal(t, "", *rows[0].InternalID)
	}
	require.NotNil(t, rows[0].ExternalID)
	require.Equal(t, "ext-1", *rows[0].ExternalID)
}
```

Check first whether `multipart` import is actually needed (it is not — remove it if the compiler complains; `graphql.Upload.File` takes any `io.Reader`-like; check the struct: it's `io.Reader`).

- [ ] **Step 2: Run the test, record the actual behavior**

Run: `go test ./internal/channels/graphapi/ -run TestUnmarshalBulkData_EmptyCells -v`
Expected: PASS (with `t.Logf` output documenting pointer-to-"" or nil). Fix the assertion to match reality; this test is the living documentation.

- [ ] **Step 3: Commit**

```bash
git add internal/channels/graphapi/bulk_import_test.go
git commit -m "test(import): document csvutil empty-cell behavior for asset bulk import"
```

---

### Task 2: GraphQL schema — options, report payload, enums (backend)

gqlgen generates the Go enums/inputs from the schema; no `pkg/enums` types needed (nothing is persisted).

**Files:**
- Modify: `internal/channels/graphapi/schema/asset.graphql` (mutation arg at ~line 61, payload at ~line 136)
- Create: `internal/channels/graphapi/schema/bulk_import.graphql`

- [ ] **Step 1: Add shared import types in `bulk_import.graphql`**

```graphql
"""
Strategy for handling rows that match an existing record
(by internal_id or external_id within the space).
"""
enum ImportStrategy {
    """Strict create: any duplicate aborts the whole import (default)."""
    CREATE
    """Skip duplicate rows, create the rest."""
    SKIP
    """Update matched records, create the rest."""
    UPSERT
}

"""
How empty CSV cells are treated when updating an existing record (UPSERT).
"""
enum ImportEmptyCellMode {
    """Empty cells leave the existing value untouched (default)."""
    IGNORE
    """Empty cells clear the field (CSV is the full truth)."""
    CLEAR
}

"""
Maps a CSV column header to a target input field name.
"""
input ImportColumnMapping {
    """CSV header exactly as found in the file."""
    column: String!
    """Target input field name (e.g. "Name", "InternalID")."""
    field: String!
}

"""
Options controlling bulk CSV import behavior.
"""
input BulkImportOptions {
    strategy: ImportStrategy! = CREATE
    emptyCells: ImportEmptyCellMode! = IGNORE
    columnMapping: [ImportColumnMapping!]
}

enum ImportRowIssueCode {
    """Row matches an existing record (CREATE/SKIP)."""
    DUPLICATE
    """internal_id and external_id match two different records."""
    CONFLICTING_MATCH
    """parent_external_id could not be resolved."""
    PARENT_NOT_FOUND
    """Row failed parsing or validation."""
    INVALID_ROW
}

"""
A problem with a single CSV row. Row numbers are 1-based file lines
(the header is line 1, the first data row is line 2).
"""
type ImportRowIssue {
    row: Int!
    """internal_id, external_id or name of the row when available."""
    identifier: String
    code: ImportRowIssueCode!
    message: String!
}
```

- [ ] **Step 2: Extend the asset mutation + payload in `asset.graphql`**

Change the existing `createBulkCSVAsset` definition to:

```graphql
    createBulkCSVAsset(
        """
        CSV file containing the assets to import
        """
        input: Upload!
        """
        Import behavior options (defaults to strict CREATE)
        """
        options: BulkImportOptions
    ): AssetBulkCreatePayload!
```

Replace the `AssetBulkCreatePayload` type with:

```graphql
type AssetBulkCreatePayload {
    """
    Created assets
    """
    assets: [Asset!]
    """
    Number of newly created assets
    """
    createdCount: Int!
    """
    Number of updated assets (UPSERT)
    """
    updatedCount: Int!
    """
    Number of skipped rows (SKIP)
    """
    skippedCount: Int!
    """
    Row-level problems encountered during the import
    """
    rowIssues: [ImportRowIssue!]!
}
```

- [ ] **Step 3: Regenerate gqlgen**

Run: `task gen:gql`
Expected: `internal/channels/graphapi/asset.resolvers.go` signature changes to `CreateBulkCSVAsset(ctx context.Context, input graphql.Upload, options *model.BulkImportOptions)`; `model/gen_models.go` gains `BulkImportOptions`, `ImportStrategy`, `ImportEmptyCellMode`, `ImportColumnMapping`, `ImportRowIssue`, `ImportRowIssueCode`.

**Check the generated enum constant names in `model/gen_models.go` before using them** (gqlgen casing is inconsistent for abbreviations — expect `ImportStrategyCreate`, `ImportStrategySkip`, `ImportStrategyUpsert`, but verify).

- [ ] **Step 4: Make it compile**

The regenerated resolver stub will panic-or-mismatch; temporarily adapt `CreateBulkCSVAsset` to the new signature, ignoring `options` (pass-through to existing `bulkCreateAsset`, return zero counts/empty issues):

```go
func (r *mutationResolver) CreateBulkCSVAsset(ctx context.Context, input graphql.Upload, options *model.BulkImportOptions) (*model.AssetBulkCreatePayload, error) {
	data, err := unmarshalBulkData[ent.CreateAssetInput](input)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal bulk data")
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("no input provided for bulk create")
	}
	// options are wired up in Task 6
	_ = options
	payload, err := r.bulkCreateAsset(ctx, data)
	if err != nil {
		return nil, err
	}
	payload.CreatedCount = len(payload.Assets)
	payload.RowIssues = []*model.ImportRowIssue{}
	return payload, nil
}
```

Run: `go build ./internal/...`
Expected: builds (ignore the known `graphapi/generate` error if running `go build ./...`).

- [ ] **Step 5: Commit**

```bash
git add internal/channels/graphapi/schema/ internal/channels/graphapi/generated/ internal/channels/graphapi/model/ internal/channels/graphapi/asset.resolvers.go
git commit -m "feat(graphapi): add BulkImportOptions and row-issue report to asset bulk import schema"
```

---

### Task 3: Column-mapping aware CSV parsing (backend)

Generalize `unmarshalBulkData` so a header rewrite can be applied before csvutil decodes. csvutil's `NewDecoder(reader, header...)` accepts an explicit header — read the first record ourselves, rewrite it, pass it in.

**Files:**
- Modify: `internal/channels/graphapi/helpers.go:63-107`
- Test: `internal/channels/graphapi/bulk_import_test.go`

- [ ] **Step 1: Write the failing tests**

```go
func TestUnmarshalBulkDataMapped_RewritesHeaders(t *testing.T) {
	csv := "Hostname;Seriennummer\nsrv-01;SN123\n"
	mapping := map[string]string{"Hostname": "Name", "Seriennummer": "Serial"}
	rows, err := unmarshalBulkDataMapped[ent.CreateAssetInput](uploadFromString(t, csv), mapping)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "srv-01", rows[0].Name)
	require.NotNil(t, rows[0].Serial)
	require.Equal(t, "SN123", *rows[0].Serial)
}

func TestUnmarshalBulkDataMapped_NilMappingKeepsHeaders(t *testing.T) {
	csv := "Name,Serial\nsrv-01,SN123\n"
	rows, err := unmarshalBulkDataMapped[ent.CreateAssetInput](uploadFromString(t, csv), nil)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "srv-01", rows[0].Name)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/channels/graphapi/ -run TestUnmarshalBulkDataMapped -v`
Expected: FAIL — `unmarshalBulkDataMapped` undefined.

- [ ] **Step 3: Implement**

In `helpers.go`, refactor `unmarshalBulkData` into a mapped variant; the old name stays as a thin wrapper (other entities keep calling it unchanged):

```go
func unmarshalBulkData[T any](input graphql.Upload) ([]*T, error) {
	return unmarshalBulkDataMapped[T](input, nil)
}

// unmarshalBulkDataMapped decodes CSV rows into T. columnMapping renames
// CSV headers before decoding (e.g. {"Hostname": "Name"}); headers without
// a mapping entry are kept as-is. Unknown headers are ignored by csvutil.
func unmarshalBulkDataMapped[T any](input graphql.Upload, columnMapping map[string]string) ([]*T, error) {
	stream, err := io.ReadAll(input.File)
	if err != nil {
		return nil, err
	}

	stream = convertToUTF8(stream)

	reader := csv.NewReader(bytes.NewReader(stream))
	reader.Comma = detectDelimiter(stream)

	headerRecord, err := reader.Read()
	if err != nil {
		return nil, err
	}
	header := make([]string, len(headerRecord))
	for i, col := range headerRecord {
		col = strings.TrimSpace(col)
		if mapped, ok := columnMapping[col]; ok {
			col = mapped
		}
		header[i] = col
	}

	dec, err := csvutil.NewDecoder(reader, header...)
	if err != nil {
		return nil, err
	}

	// Register custom decoder for []string - splits comma-separated values
	dec.Register(func(data []byte, v *[]string) error {
		s := strings.TrimSpace(string(data))
		if s == "" {
			*v = nil
			return nil
		}
		parts := strings.Split(s, ",")
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		*v = parts
		return nil
	})

	var result []*T
	for {
		item := new(T)
		if err := dec.Decode(item); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, nil
}
```

Note: passing the header explicitly changes nothing for existing callers — csvutil just no longer reads it itself. Verify `csvutil.NewDecoder` variadic header signature against the vendored version if the compiler disagrees.

- [ ] **Step 4: Run all graphapi tests**

Run: `go test ./internal/channels/graphapi/ -run "TestUnmarshalBulkData" -v`
Expected: all PASS (including Task 1's test — proves the refactor kept behavior).

- [ ] **Step 5: Commit**

```bash
git add internal/channels/graphapi/helpers.go internal/channels/graphapi/bulk_import_test.go
git commit -m "feat(graphapi): support column-mapping header rewrite in bulk CSV parsing"
```

---

### Task 4: Import pipeline core — normalize + classify (backend)

Pure functions, fully unit-testable without a DB.

**Files:**
- Create: `internal/channels/graphapi/bulk_import.go`
- Test: `internal/channels/graphapi/bulk_import_test.go`

- [ ] **Step 1: Write the failing tests**

```go
func strPtr(s string) *string { return &s }

func TestNormalizeAssetImportRow(t *testing.T) {
	row := &ent.CreateAssetInput{
		Name:       "  Server A  ",
		InternalID: strPtr("  "),
		ExternalID: strPtr(" ext-1 "),
	}
	normalizeAssetImportRow(row)
	require.Equal(t, "Server A", row.Name)
	require.Nil(t, row.InternalID, "empty internal_id must become unset, not \"\"")
	require.Equal(t, "ext-1", *row.ExternalID)
}

func TestClassifyImportRow(t *testing.T) {
	existingByInternal := map[string]string{"int-1": "asset-a"}
	existingByExternal := map[string]string{"ext-9": "asset-b"}

	t.Run("no ids -> create", func(t *testing.T) {
		got := classifyImportRow(&ent.CreateAssetInput{Name: "x"}, existingByInternal, existingByExternal)
		require.Equal(t, rowActionCreate, got.action)
	})

	t.Run("internal id match -> update", func(t *testing.T) {
		got := classifyImportRow(&ent.CreateAssetInput{Name: "x", InternalID: strPtr("int-1")}, existingByInternal, existingByExternal)
		require.Equal(t, rowActionUpdate, got.action)
		require.Equal(t, "asset-a", got.matchID)
	})

	t.Run("external id match -> update", func(t *testing.T) {
		got := classifyImportRow(&ent.CreateAssetInput{Name: "x", ExternalID: strPtr("ext-9")}, existingByInternal, existingByExternal)
		require.Equal(t, rowActionUpdate, got.action)
		require.Equal(t, "asset-b", got.matchID)
	})

	t.Run("conflicting matches -> conflict", func(t *testing.T) {
		got := classifyImportRow(&ent.CreateAssetInput{
			Name: "x", InternalID: strPtr("int-1"), ExternalID: strPtr("ext-9"),
		}, existingByInternal, existingByExternal)
		require.Equal(t, rowActionConflict, got.action)
	})

	t.Run("both ids same asset -> update", func(t *testing.T) {
		byExt := map[string]string{"ext-1": "asset-a"}
		got := classifyImportRow(&ent.CreateAssetInput{
			Name: "x", InternalID: strPtr("int-1"), ExternalID: strPtr("ext-1"),
		}, existingByInternal, byExt)
		require.Equal(t, rowActionUpdate, got.action)
		require.Equal(t, "asset-a", got.matchID)
	})

	t.Run("no match -> create", func(t *testing.T) {
		got := classifyImportRow(&ent.CreateAssetInput{Name: "x", InternalID: strPtr("other")}, existingByInternal, existingByExternal)
		require.Equal(t, rowActionCreate, got.action)
	})
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/channels/graphapi/ -run "TestNormalizeAssetImportRow|TestClassifyImportRow" -v`
Expected: FAIL — functions undefined.

- [ ] **Step 3: Implement in `bulk_import.go`**

```go
// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package graphapi

import (
	"strings"

	"github.com/kopexa-grc/kopexa/internal/store/ent"
)

type rowAction int

const (
	rowActionCreate rowAction = iota
	rowActionUpdate
	rowActionConflict
)

type rowClassification struct {
	action  rowAction
	matchID string
}

// trimToNil trims the pointee and unsets pointers to blank strings so the
// partial unique indexes on (space_id, internal_id/external_id) never see
// "" as a value (root cause of kopexa#878).
func trimToNil(v **string) {
	if *v == nil {
		return
	}
	trimmed := strings.TrimSpace(**v)
	if trimmed == "" {
		*v = nil
		return
	}
	**v = trimmed
}

func normalizeAssetImportRow(row *ent.CreateAssetInput) {
	row.Name = strings.TrimSpace(row.Name)
	trimToNil(&row.InternalID)
	trimToNil(&row.ExternalID)
}

// classifyImportRow decides what to do with a row. internal_id wins over
// external_id; if both resolve to different assets the row is a conflict.
// Rows without either ID always create — name is NOT a match key (not unique).
func classifyImportRow(
	row *ent.CreateAssetInput,
	existingByInternalID map[string]string,
	existingByExternalID map[string]string,
) rowClassification {
	var internalMatch, externalMatch string
	if row.InternalID != nil {
		internalMatch = existingByInternalID[*row.InternalID]
	}
	if row.ExternalID != nil {
		externalMatch = existingByExternalID[*row.ExternalID]
	}

	switch {
	case internalMatch != "" && externalMatch != "" && internalMatch != externalMatch:
		return rowClassification{action: rowActionConflict}
	case internalMatch != "":
		return rowClassification{action: rowActionUpdate, matchID: internalMatch}
	case externalMatch != "":
		return rowClassification{action: rowActionUpdate, matchID: externalMatch}
	default:
		return rowClassification{action: rowActionCreate}
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/channels/graphapi/ -run "TestNormalizeAssetImportRow|TestClassifyImportRow" -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/channels/graphapi/bulk_import.go internal/channels/graphapi/bulk_import_test.go
git commit -m "feat(graphapi): add import row normalization and classification"
```

---

### Task 5: Update-input builder with empty-cell modes (backend)

Builds `ent.UpdateAssetInput` from a CSV row for UPSERT. `IGNORE`: only non-empty values are applied. `CLEAR`: a column that is present in the file but empty clears the field — which requires knowing the present columns (missing column ≠ empty cell).

**Files:**
- Modify: `internal/channels/graphapi/bulk_import.go`
- Test: `internal/channels/graphapi/bulk_import_test.go`

- [ ] **Step 1: Write the failing tests**

```go
func TestBuildAssetUpdateInput_Ignore(t *testing.T) {
	row := &ent.CreateAssetInput{
		Name:        "Server A",
		Description: nil,                 // empty cell
		Serial:      strPtr("SN-NEW"),    // value
	}
	upd := buildAssetUpdateInput(row, model.ImportEmptyCellModeIgnore, nil)
	require.NotNil(t, upd.Name)
	require.Equal(t, "Server A", *upd.Name)
	require.Nil(t, upd.Description)
	require.False(t, upd.ClearDescription)
	require.Equal(t, "SN-NEW", *upd.Serial)
}

func TestBuildAssetUpdateInput_Clear(t *testing.T) {
	present := map[string]bool{"Name": true, "Description": true}
	row := &ent.CreateAssetInput{Name: "Server A", Description: nil}
	upd := buildAssetUpdateInput(row, model.ImportEmptyCellModeClear, present)
	require.Equal(t, "Server A", *upd.Name)
	require.True(t, upd.ClearDescription, "present-but-empty column must clear")
	require.False(t, upd.ClearSerial, "absent column must NOT clear")
}

func TestBuildAssetUpdateInput_EmptyNameNeverApplied(t *testing.T) {
	present := map[string]bool{"Name": true}
	row := &ent.CreateAssetInput{Name: ""}
	upd := buildAssetUpdateInput(row, model.ImportEmptyCellModeClear, present)
	require.Nil(t, upd.Name, "name is required and cannot be cleared or blanked")
}
```

(Adjust the `model.ImportEmptyCellMode*` constant names to whatever gqlgen actually generated in Task 2.)

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/channels/graphapi/ -run TestBuildAssetUpdateInput -v`
Expected: FAIL — `buildAssetUpdateInput` undefined.

- [ ] **Step 3: Implement**

Append to `bulk_import.go` (import `"github.com/kopexa-grc/kopexa/internal/channels/graphapi/model"`):

```go
// buildAssetUpdateInput maps a CSV row onto an UpdateAssetInput.
//
//   - IGNORE: only non-nil/non-empty values are applied; nothing is cleared.
//   - CLEAR: a column present in the file but empty clears the field.
//     presentColumns holds the (mapped) header names of the uploaded file.
//
// Only scalar fields that exist in the CSV template are handled; edges and
// metadata are intentionally untouched by upsert.
func buildAssetUpdateInput(
	row *ent.CreateAssetInput,
	mode model.ImportEmptyCellMode,
	presentColumns map[string]bool,
) ent.UpdateAssetInput {
	upd := ent.UpdateAssetInput{}
	clear := mode == model.ImportEmptyCellModeClear

	// Name is required: apply when non-empty, never clear.
	if name := strings.TrimSpace(row.Name); name != "" {
		upd.Name = &name
	}

	setString := func(column string, value *string, target **string, clearFlag *bool) {
		if value != nil && strings.TrimSpace(*value) != "" {
			trimmed := strings.TrimSpace(*value)
			*target = &trimmed
			return
		}
		if clear && presentColumns[column] {
			*clearFlag = true
		}
	}

	setString("InternalID", row.InternalID, &upd.InternalID, &upd.ClearInternalID)
	setString("ExternalID", row.ExternalID, &upd.ExternalID, &upd.ClearExternalID)
	setString("Description", row.Description, &upd.Description, &upd.ClearDescription)
	setString("Location", row.Location, &upd.Location, &upd.ClearLocation)
	setString("Serial", row.Serial, &upd.Serial, &upd.ClearSerial)
	setString("Manufacturer", row.Manufacturer, &upd.Manufacturer, &upd.ClearManufacturer)
	setString("Model", row.Model, &upd.Model, &upd.ClearModel)
	setString("Os", row.Os, &upd.Os, &upd.ClearOs)
	setString("Source", row.Source, &upd.Source, &upd.ClearSource)

	// Enums: apply only when set; never cleared via CSV.
	if row.Kind != "" {
		kind := row.Kind
		upd.Kind = &kind
	}
	if row.Status != nil {
		upd.Status = row.Status
	}
	if row.IsPortable != nil {
		upd.IsPortable = row.IsPortable
	}

	return upd
}
```

Check `UpdateAssetInput` for the exact `ClearExternalID`/`ClearOs` field names before assuming (read `internal/store/ent/gql_mutation_input.go:1006-1100`).

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/channels/graphapi/ -run TestBuildAssetUpdateInput -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/channels/graphapi/bulk_import.go internal/channels/graphapi/bulk_import_test.go
git commit -m "feat(graphapi): build asset update input with IGNORE/CLEAR empty-cell modes"
```

---

### Task 6: Pipeline runner + resolver wiring (backend)

The orchestration: prefetch, classify, execute, report. Runs inside the existing entgql transaction.

**Files:**
- Modify: `internal/channels/graphapi/bulk_import.go`
- Modify: `internal/channels/graphapi/asset.resolvers.go` (CreateBulkCSVAsset)
- Test: manual against dev server (integration; DB-backed unit tests don't exist in this package — keep logic pure where possible)

- [ ] **Step 1: Add the asset import row wrapper + runner**

Append to `bulk_import.go`:

```go
import (
	// add to existing imports:
	"context"
	"fmt"

	"github.com/kopexa-grc/kopexa/internal/store/ent/asset"
)

// assetImportRow embeds the generated create input and adds CSV-only
// virtual columns. ParentExternalIDs uses the registered []string decoder
// (comma-separated cell).
type assetImportRow struct {
	ent.CreateAssetInput
	ParentExternalIDs []string `csv:"ParentExternalID"`
}

type assetImportResult struct {
	created []*ent.Asset
	updated int
	skipped int
	issues  []*model.ImportRowIssue
}

func rowIdentifier(row *ent.CreateAssetInput) *string {
	switch {
	case row.InternalID != nil:
		return row.InternalID
	case row.ExternalID != nil:
		return row.ExternalID
	case row.Name != "":
		name := row.Name
		return &name
	default:
		return nil
	}
}

// runAssetImport executes the configurable import. The caller's context
// carries the entgql transaction; any returned error rolls everything back.
func runAssetImport(
	ctx context.Context,
	client *ent.Client,
	spaceID string,
	rows []*assetImportRow,
	opts model.BulkImportOptions,
	presentColumns map[string]bool,
) (*assetImportResult, error) {
	res := &assetImportResult{issues: []*model.ImportRowIssue{}}

	// --- normalize + collect match keys
	internalIDs := make([]string, 0, len(rows))
	externalIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		normalizeAssetImportRow(&row.CreateAssetInput)
		if row.InternalID != nil {
			internalIDs = append(internalIDs, *row.InternalID)
		}
		if row.ExternalID != nil {
			externalIDs = append(externalIDs, *row.ExternalID)
		}
	}

	// --- prefetch existing assets (two queries, no N+1)
	existingByInternalID := map[string]string{}
	existingByExternalID := map[string]string{}
	if len(internalIDs) > 0 {
		found, err := client.Asset.Query().
			Where(asset.SpaceID(spaceID), asset.InternalIDIn(internalIDs...)).
			All(ctx)
		if err != nil {
			return nil, parseRequestError(err, action{action: ActionGet, object: "asset"})
		}
		for _, a := range found {
			existingByInternalID[a.InternalID] = a.ID
		}
	}
	if len(externalIDs) > 0 {
		found, err := client.Asset.Query().
			Where(asset.SpaceID(spaceID), asset.ExternalIDIn(externalIDs...)).
			All(ctx)
		if err != nil {
			return nil, parseRequestError(err, action{action: ActionGet, object: "asset"})
		}
		for _, a := range found {
			existingByExternalID[a.ExternalID] = a.ID
		}
	}

	// --- classify + execute
	// externalIDToAssetID resolves parent references afterwards; seeded with
	// existing assets, extended with freshly created rows.
	externalIDToAssetID := map[string]string{}
	for extID, id := range existingByExternalID {
		externalIDToAssetID[extID] = id
	}

	type pendingCreate struct {
		row     *assetImportRow
		csvLine int
	}
	type pendingParent struct {
		assetID string
		refs    []string
		csvLine int
	}
	var creates []pendingCreate
	var parents []pendingParent

	for i, row := range rows {
		csvLine := i + 2 // header is line 1

		if row.Name == "" {
			res.issues = append(res.issues, &model.ImportRowIssue{
				Row: csvLine, Identifier: rowIdentifier(&row.CreateAssetInput),
				Code: model.ImportRowIssueCodeInvalidRow, Message: "name is required",
			})
			continue
		}

		cls := classifyImportRow(&row.CreateAssetInput, existingByInternalID, existingByExternalID)

		switch cls.action {
		case rowActionConflict:
			res.issues = append(res.issues, &model.ImportRowIssue{
				Row: csvLine, Identifier: rowIdentifier(&row.CreateAssetInput),
				Code:    model.ImportRowIssueCodeConflictingMatch,
				Message: "internal_id and external_id match two different assets",
			})

		case rowActionUpdate:
			switch opts.Strategy {
			case model.ImportStrategyCreate:
				res.issues = append(res.issues, &model.ImportRowIssue{
					Row: csvLine, Identifier: rowIdentifier(&row.CreateAssetInput),
					Code: model.ImportRowIssueCodeDuplicate, Message: "asset already exists",
				})
			case model.ImportStrategySkip:
				res.skipped++
				res.issues = append(res.issues, &model.ImportRowIssue{
					Row: csvLine, Identifier: rowIdentifier(&row.CreateAssetInput),
					Code: model.ImportRowIssueCodeDuplicate, Message: "asset already exists, row skipped",
				})
			case model.ImportStrategyUpsert:
				upd := buildAssetUpdateInput(&row.CreateAssetInput, opts.EmptyCells, presentColumns)
				if err := client.Asset.UpdateOneID(cls.matchID).SetInput(upd).Exec(ctx); err != nil {
					return nil, parseRequestError(err, action{action: ActionUpdate, object: "asset"})
				}
				res.updated++
				if len(row.ParentExternalIDs) > 0 {
					parents = append(parents, pendingParent{assetID: cls.matchID, refs: row.ParentExternalIDs, csvLine: csvLine})
				}
			}

		case rowActionCreate:
			creates = append(creates, pendingCreate{row: row, csvLine: csvLine})
		}
	}

	// CREATE strategy is all-or-nothing: any duplicate aborts before writes.
	if opts.Strategy == model.ImportStrategyCreate && len(res.issues) > 0 {
		return res, nil // caller turns issues into an error + report
	}

	// --- batch create
	if len(creates) > 0 {
		builders := make([]*ent.AssetCreate, len(creates))
		for i, pc := range creates {
			pc.row.ParentID = spaceID
			builders[i] = client.Asset.Create().SetInput(pc.row.CreateAssetInput)
		}
		created, err := client.Asset.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return nil, parseRequestError(err, action{action: ActionCreate, object: "asset"})
		}
		res.created = created
		for i, a := range created {
			pc := creates[i]
			if pc.row.ExternalID != nil {
				externalIDToAssetID[*pc.row.ExternalID] = a.ID
			}
			if len(pc.row.ParentExternalIDs) > 0 {
				parents = append(parents, pendingParent{assetID: a.ID, refs: pc.row.ParentExternalIDs, csvLine: pc.csvLine})
			}
		}
	}

	// --- post phase: resolve parent references
	for _, p := range parents {
		var parentIDs []string
		for _, ref := range p.refs {
			if id, ok := externalIDToAssetID[ref]; ok {
				parentIDs = append(parentIDs, id)
			} else {
				ref := ref
				res.issues = append(res.issues, &model.ImportRowIssue{
					Row: p.csvLine, Identifier: &ref,
					Code:    model.ImportRowIssueCodeParentNotFound,
					Message: fmt.Sprintf("parent with external_id %q not found", ref),
				})
			}
		}
		if len(parentIDs) > 0 {
			if err := client.Asset.UpdateOneID(p.assetID).AddParentAssetIDs(parentIDs...).Exec(ctx); err != nil {
				return nil, parseRequestError(err, action{action: ActionUpdate, object: "asset"})
			}
		}
	}

	return res, nil
}
```

Verify before coding: `asset.SpaceID(...)` predicate name, `AddParentAssetIDs` on the update builder, and the gqlgen enum constant names from Task 2.

- [ ] **Step 2: Wire the resolver**

Replace the Task-2 interim `CreateBulkCSVAsset` in `asset.resolvers.go`:

```go
// CreateBulkCSVAsset imports assets from a CSV file with configurable
// duplicate handling (see BulkImportOptions).
func (r *mutationResolver) CreateBulkCSVAsset(ctx context.Context, input graphql.Upload, options *model.BulkImportOptions) (*model.AssetBulkCreatePayload, error) {
	opts := model.BulkImportOptions{
		Strategy:   model.ImportStrategyCreate,
		EmptyCells: model.ImportEmptyCellModeIgnore,
	}
	if options != nil {
		opts = *options
	}

	columnMapping := map[string]string{}
	presentColumns := map[string]bool{}
	for _, m := range opts.ColumnMapping {
		columnMapping[m.Column] = m.Field
	}

	rows, header, err := unmarshalAssetImportRows(input, columnMapping)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal bulk data")
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("no input provided for bulk create")
	}
	for _, col := range header {
		presentColumns[col] = true
	}

	spaceID := auth.SpaceFromContext(ctx)
	client := withTransactionalMutation(ctx)

	res, err := runAssetImport(ctx, client, spaceID, rows, opts, presentColumns)
	if err != nil {
		return nil, err
	}

	// Strict CREATE: duplicates abort the import (nothing was written).
	if opts.Strategy == model.ImportStrategyCreate {
		for _, issue := range res.issues {
			if issue.Code == model.ImportRowIssueCodeDuplicate {
				return &model.AssetBulkCreatePayload{
					Assets:    nil,
					RowIssues: res.issues,
				}, newAlreadyExistsError("asset")
			}
		}
	}

	return &model.AssetBulkCreatePayload{
		Assets:       res.created,
		CreatedCount: len(res.created),
		UpdatedCount: res.updated,
		SkippedCount: res.skipped,
		RowIssues:    res.issues,
	}, nil
}
```

Note: gqlgen drops the payload when an error is returned — for strict CREATE the client keeps getting `ALREADY_EXISTS` (backward compatible). If we want the report alongside, that is a follow-up (error extensions); don't build it now (YAGNI).

`unmarshalAssetImportRows` is a small wrapper in `bulk_import.go` that also returns the (mapped) header for `presentColumns`:

```go
func unmarshalAssetImportRows(input graphql.Upload, columnMapping map[string]string) ([]*assetImportRow, []string, error) {
	// re-read header for presentColumns, then decode rows
	rows, err := unmarshalBulkDataMapped[assetImportRow](input, columnMapping)
	if err != nil {
		return nil, nil, err
	}
	header := make([]string, 0, len(columnMapping))
	// presentColumns must reflect the actual file: capture the header inside
	// unmarshalBulkDataMapped instead of re-parsing. Refactor it to also
	// return the final header ([]*T, []string, error) and adjust the two
	// existing tests from Task 3 accordingly.
	_ = header
	return rows, header, nil
}
```

**Implementation note:** do the small refactor described in the comment — `unmarshalBulkDataMapped` returns `([]*T, []string, error)`; `unmarshalBulkData` wrapper discards the header. Update Task 3 tests.

- [ ] **Step 3: Build + run all package tests**

Run: `go build ./internal/... && go test ./internal/channels/graphapi/ -run "TestUnmarshal|TestNormalize|TestClassify|TestBuildAsset" -v`
Expected: builds, all PASS.

- [ ] **Step 4: Manual integration check against dev server**

Restart `kopexa serve`, then via GraphQL playground or curl: import a small CSV twice — once `SKIP` (expect skippedCount > 0), once `UPSERT` (expect updatedCount > 0), once without options against existing data (expect `ALREADY_EXISTS`).

- [ ] **Step 5: Commit**

```bash
git add internal/channels/graphapi/bulk_import.go internal/channels/graphapi/asset.resolvers.go internal/channels/graphapi/helpers.go internal/channels/graphapi/bulk_import_test.go
git commit -m "feat(graphapi): configurable asset CSV import with skip/upsert and row report (#878)"
```

---

### Task 7: Empty-string normalization hook (backend)

Prevents the `""`-collision bug from returning through any mutation path (UI forms, API, integrations).

**Files:**
- Modify: `internal/store/hooks/asset.go`
- Test: `internal/store/hooks/asset_test.go` if a hook-test pattern exists in the package — check `ls internal/store/hooks/*_test.go` and follow it; if hooks are only integration-tested, verify manually via dev server instead.

- [ ] **Step 1: Extend the asset hook**

In `HookCreateAsset` (and the update path — check whether the file has a combined mutation hook; follow its structure), normalize before save:

```go
// inside the hook func, before mut/next call:
if v, ok := m.InternalID(); ok && strings.TrimSpace(v) == "" {
	m.ClearInternalID()
}
if v, ok := m.ExternalID(); ok && strings.TrimSpace(v) == "" {
	m.ClearExternalID()
}
```

Mutation accessors return `(string, bool)` — no pointer dereference (project convention). Register for `ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne` — check how the existing hook is registered in the schema and extend the op mask there if needed.

- [ ] **Step 2: Build + run hook tests (or manual check)**

Run: `go build ./internal/... && go test ./internal/store/hooks/ -v 2>&1 | tail -5`
Expected: builds; existing tests PASS. Manual: create an asset with empty internal_id via UI twice — no constraint error.

- [ ] **Step 3: Commit**

```bash
git add internal/store/hooks/asset.go
git commit -m "fix(assets): normalize empty internal/external IDs to unset (#878)"
```

---

### Task 8: Frontend — mutation document + codegen

**Files:**
- Modify: `src/generated/queries/asset.ts:380-388` (IMPORT_ASSETS)
- Generated: `src/generated/graphql.ts` (codegen)

- [ ] **Step 1: Update the mutation document**

```typescript
export const IMPORT_ASSETS = gql`
mutation ImportAssets($input: Upload!, $options: BulkImportOptions) {
  createBulkCSVAsset(input: $input, options: $options) {
    assets {
      id
    }
    createdCount
    updatedCount
    skippedCount
    rowIssues {
      row
      identifier
      code
      message
    }
  }
}
`;
```

- [ ] **Step 2: Regenerate types**

Backend with the new schema must run on :8080 (ask Julian to restart or verify `curl -s localhost:8080/graphql/query -X POST -H 'Content-Type: application/json' -d '{"query":"{ __type(name: \"BulkImportOptions\") { name } }"}'` returns the type). Then:

Run: `pnpm run codegen:gql`
Expected: SUCCESS; `ImportAssetsMutationVariables` gains `options`.

- [ ] **Step 3: Typecheck + commit**

Run: `pnpm exec tsc --noEmit` — expect clean (only known pre-existing errors, if any).

```bash
git add src/generated/ schema.graphql
git commit -m "feat(assets): add import options and row report to import mutation"
```

---

### Task 9: Frontend — import wizard (file → mapping → options → result)

**Files:**
- Rewrite: `src/modules/inventory/assets/components/import-asset-drawer.tsx`
- Create: `src/modules/inventory/assets/components/import/csv-header.ts` (tiny header parser + auto-match)
- Create: `src/modules/inventory/assets/components/import/csv-header.test.ts`
- Modify: `messages/de/common.json`, `messages/en/common.json` (new `file_import.*` keys)

- [ ] **Step 1: Write the failing header-parser tests**

```typescript
import { describe, expect, it } from "vitest";
import { autoMatchColumns, parseCsvHeader } from "./csv-header";

describe("parseCsvHeader", () => {
	it("parses comma and semicolon delimited headers", () => {
		expect(parseCsvHeader("Name,Serial\nrow")).toEqual(["Name", "Serial"]);
		expect(parseCsvHeader("Hostname;Seriennummer\nrow")).toEqual([
			"Hostname",
			"Seriennummer",
		]);
	});

	it("strips quotes and whitespace", () => {
		expect(parseCsvHeader('"Name" , "Serial"\n')).toEqual(["Name", "Serial"]);
	});
});

describe("autoMatchColumns", () => {
	it("matches exact field names case-insensitively", () => {
		expect(autoMatchColumns(["name", "InternalID", "Unbekannt"])).toEqual({
			name: "Name",
			InternalID: "InternalID",
		});
	});
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `pnpm exec vitest run src/modules/inventory/assets/components/import/csv-header.test.ts`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement `csv-header.ts`**

```typescript
/** Asset fields offered in the mapping step. Mirrors the CSV template. */
export const ASSET_IMPORT_FIELDS = [
	"Name",
	"InternalID",
	"ExternalID",
	"Description",
	"Kind",
	"Status",
	"Location",
	"Serial",
	"Manufacturer",
	"Model",
	"Os",
	"IsPortable",
	"Source",
	"ParentExternalID",
] as const;

export type AssetImportField = (typeof ASSET_IMPORT_FIELDS)[number];

const detectDelimiter = (line: string): string => {
	const candidates = [",", ";", "\t", "|"];
	let best = ",";
	let bestCount = 0;
	for (const c of candidates) {
		const count = line.split(c).length - 1;
		if (count > bestCount) {
			best = c;
			bestCount = count;
		}
	}
	return best;
};

/** Returns the trimmed, unquoted column names of the first CSV line. */
export function parseCsvHeader(content: string): string[] {
	const firstLine = content.split(/\r?\n/, 1)[0] ?? "";
	const delimiter = detectDelimiter(firstLine);
	return firstLine
		.split(delimiter)
		.map((col) => col.trim().replace(/^"(.*)"$/, "$1").trim())
		.filter((col) => col.length > 0);
}

/** Auto-assigns CSV columns whose name equals a known field (case-insensitive). */
export function autoMatchColumns(
	columns: string[],
): Record<string, AssetImportField> {
	const byLower = new Map<string, AssetImportField>(
		ASSET_IMPORT_FIELDS.map((f) => [f.toLowerCase(), f]),
	);
	const result: Record<string, AssetImportField> = {};
	for (const col of columns) {
		const match = byLower.get(col.toLowerCase());
		if (match) result[col] = match;
	}
	return result;
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `pnpm exec vitest run src/modules/inventory/assets/components/import/csv-header.test.ts`
Expected: PASS.

- [ ] **Step 5: Rewrite the drawer as a wizard**

Rewrite `import-asset-drawer.tsx`. Structure (complete file; reuse existing imports where unchanged):

```tsx
"use client";

import { Button, Drawer, Select } from "@kopexa/sight";
import { useQueryClient } from "@tanstack/react-query";
import { DownloadIcon, ImportIcon } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import type { ImportAssetsMutation } from "@/generated/graphql";
import {
	ImportEmptyCellMode,
	ImportStrategy,
} from "@/generated/graphql";
import { useAssetsList, useImportAssets } from "@/generated/hooks/asset";
import { useTemplateDownload } from "@/modules/common/export/use-csv-export";
import { ImportInfoCallout } from "@/modules/common/import/import-info-callout";
import {
	ASSET_IMPORT_FIELDS,
	autoMatchColumns,
	parseCsvHeader,
} from "./import/csv-header";

type Step = "file" | "mapping" | "options" | "result";

type ImportAssetDrawerProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
};

const IGNORE_COLUMN = "__ignore__";

export const ImportAssetDrawer = ({
	open,
	onOpenChange,
}: ImportAssetDrawerProps) => {
	const t = useTranslations("common");
	const { mutateAsync: importCsv, isPending } = useImportAssets();
	const queryClient = useQueryClient();

	const [step, setStep] = useState<Step>("file");
	const [file, setFile] = useState<File | null>(null);
	const [columns, setColumns] = useState<string[]>([]);
	const [mapping, setMapping] = useState<Record<string, string>>({});
	const [strategy, setStrategy] = useState<ImportStrategy>(
		ImportStrategy.SKIP,
	);
	const [emptyCells, setEmptyCells] = useState<ImportEmptyCellMode>(
		ImportEmptyCellMode.IGNORE,
	);
	const [result, setResult] = useState<
		ImportAssetsMutation["createBulkCSVAsset"] | null
	>(null);

	const { isDownloading, handleDownloadTemplate } = useTemplateDownload({
		templateName: "asset",
		downloadFilename: "asset-template.csv",
	});

	const reset = () => {
		setStep("file");
		setFile(null);
		setColumns([]);
		setMapping({});
		setStrategy(ImportStrategy.SKIP);
		setEmptyCells(ImportEmptyCellMode.IGNORE);
		setResult(null);
	};

	const handleFileSelected = async (selected: File) => {
		const head = await selected.slice(0, 64 * 1024).text();
		const cols = parseCsvHeader(head);
		setFile(selected);
		setColumns(cols);
		setMapping(autoMatchColumns(cols));
		setStep("mapping");
	};

	const nameIsMapped = Object.values(mapping).includes("Name");

	const handleImport = async () => {
		if (!file) return;
		try {
			const data = await importCsv({
				input: file,
				options: {
					strategy,
					emptyCells,
					columnMapping: Object.entries(mapping)
						.filter(([, field]) => field !== IGNORE_COLUMN)
						.map(([column, field]) => ({ column, field })),
				},
			});
			queryClient.invalidateQueries({ queryKey: useAssetsList.getKey() });
			setResult(data.createBulkCSVAsset);
			setStep("result");
		} catch {
			toast.error(t("file_import.import_error"));
		}
	};

	return (
		<Drawer.Root
			open={open}
			onOpenChange={(o) => {
				if (!o) reset();
				onOpenChange(o);
			}}
			size="lg"
		>
			<Drawer.Content aria-describedby={undefined}>
				<Drawer.Header>
					<Drawer.Title className="flex items-center gap-2">
						<ImportIcon className="size-5" />
						{t("import_assets")}
					</Drawer.Title>
				</Drawer.Header>
				<div className="p-6 space-y-6">
					{step === "file" && (
						<>
							<ImportInfoCallout />
							<div className="space-y-3">
								<h3 className="text-sm font-medium">
									{t("download_template")}
								</h3>
								<p className="text-sm text-muted-foreground">
									{t("download_template_description_assets")}
								</p>
								<Button
									variant="outline"
									startContent={<DownloadIcon className="size-4" />}
									onClick={handleDownloadTemplate}
									isLoading={isDownloading}
								>
									{t("download_csv_template")}
								</Button>
							</div>
							<div className="space-y-3">
								<h3 className="text-sm font-medium">{t("upload_csv")}</h3>
								<input
									type="file"
									accept=".csv"
									onChange={(e) => {
										const f = e.target.files?.[0];
										if (f) {
											handleFileSelected(f);
											e.target.value = "";
										}
									}}
									className="hidden"
									id="csv-upload-asset"
								/>
								<Button
									variant="outline"
									onClick={() =>
										document.getElementById("csv-upload-asset")?.click()
									}
								>
									{t("upload_csv_file")}
								</Button>
							</div>
						</>
					)}

					{step === "mapping" && (
						<div className="space-y-4">
							<h3 className="text-sm font-medium">
								{t("file_import.mapping_title")}
							</h3>
							<p className="text-sm text-muted-foreground">
								{t("file_import.mapping_description")}
							</p>
							<div className="space-y-2">
								{columns.map((col) => (
									<div key={col} className="flex items-center gap-3">
										<span className="w-1/2 truncate text-sm font-mono">
											{col}
										</span>
										<Select
											value={mapping[col] ?? IGNORE_COLUMN}
											onValueChange={(field) =>
												setMapping((prev) => ({ ...prev, [col]: field }))
											}
											options={[
												{
													value: IGNORE_COLUMN,
													label: t("file_import.ignore_column"),
												},
												...ASSET_IMPORT_FIELDS.map((f) => ({
													value: f,
													label: f,
												})),
											]}
										/>
									</div>
								))}
							</div>
							{!nameIsMapped && (
								<p className="text-sm text-destructive">
									{t("file_import.name_mapping_required")}
								</p>
							)}
							<div className="flex justify-between">
								<Button variant="ghost" onClick={() => setStep("file")}>
									{t("back")}
								</Button>
								<Button
									onClick={() => setStep("options")}
									disabled={!nameIsMapped}
								>
									{t("continue")}
								</Button>
							</div>
						</div>
					)}

					{step === "options" && (
						<div className="space-y-4">
							<h3 className="text-sm font-medium">
								{t("file_import.strategy_title")}
							</h3>
							{(
								[
									ImportStrategy.SKIP,
									ImportStrategy.UPSERT,
									ImportStrategy.CREATE,
								] as const
							).map((s) => (
								<label key={s} className="flex items-start gap-2 cursor-pointer">
									<input
										type="radio"
										name="import-strategy"
										checked={strategy === s}
										onChange={() => setStrategy(s)}
									/>
									<span>
										<span className="block text-sm font-medium">
											{t(`file_import.strategy_${s.toLowerCase()}`)}
										</span>
										<span className="block text-xs text-muted-foreground">
											{t(`file_import.strategy_${s.toLowerCase()}_description`)}
										</span>
									</span>
								</label>
							))}
							{strategy === ImportStrategy.UPSERT && (
								<label className="flex items-center gap-2 cursor-pointer pl-6">
									<input
										type="checkbox"
										checked={emptyCells === ImportEmptyCellMode.CLEAR}
										onChange={(e) =>
											setEmptyCells(
												e.target.checked
													? ImportEmptyCellMode.CLEAR
													: ImportEmptyCellMode.IGNORE,
											)
										}
									/>
									<span className="text-sm">
										{t("file_import.clear_empty_cells")}
									</span>
								</label>
							)}
							<div className="flex justify-between">
								<Button variant="ghost" onClick={() => setStep("mapping")}>
									{t("back")}
								</Button>
								<Button onClick={handleImport} isLoading={isPending}>
									{t("file_import.start_import")}
								</Button>
							</div>
						</div>
					)}

					{step === "result" && result && (
						<div className="space-y-4">
							<h3 className="text-sm font-medium">
								{t("file_import.result_title")}
							</h3>
							<ul className="text-sm space-y-1">
								<li>
									{t("file_import.result_created", {
										count: result.createdCount,
									})}
								</li>
								<li>
									{t("file_import.result_updated", {
										count: result.updatedCount,
									})}
								</li>
								<li>
									{t("file_import.result_skipped", {
										count: result.skippedCount,
									})}
								</li>
							</ul>
							{result.rowIssues.length > 0 && (
								<details className="text-sm">
									<summary className="cursor-pointer font-medium">
										{t("file_import.row_issues", {
											count: result.rowIssues.length,
										})}
									</summary>
									<ul className="mt-2 space-y-1 text-muted-foreground">
										{result.rowIssues.map((issue) => (
											<li key={`${issue.row}-${issue.code}-${issue.identifier}`}>
												{t("file_import.row_prefix", { row: issue.row })}{" "}
												{issue.identifier ? `${issue.identifier}: ` : ""}
												{t(`file_import.issue_${issue.code.toLowerCase()}`)}
											</li>
										))}
									</ul>
								</details>
							)}
							<div className="flex justify-end">
								<Button onClick={() => onOpenChange(false)}>
									{t("close")}
								</Button>
							</div>
						</div>
					)}
				</div>
			</Drawer.Content>
		</Drawer.Root>
	);
};
```

Adjust to reality while implementing: sight `Select` API (check an existing usage for the exact props — `options` vs `Select.Item` children), enum constant names from codegen (`ImportStrategy.SKIP` vs `ImportStrategySkip` — check `src/generated/graphql.ts`; codegen config uses `namingConvention: { enumValues: "keep" }` so values keep their SCREAMING_CASE), and whether a sight RadioGroup component exists to replace the raw inputs.

- [ ] **Step 6: Add i18n keys**

Append under `common.file_import` in `messages/de/common.json` (and EN equivalents — informal "du" in German):

```json
{
	"mapping_title": "Spalten zuordnen",
	"mapping_description": "Ordne die Spalten deiner CSV-Datei den Asset-Feldern zu. Nicht zugeordnete Spalten werden ignoriert.",
	"ignore_column": "Ignorieren",
	"name_mapping_required": "Die Spalte für den Namen muss zugeordnet sein.",
	"strategy_title": "Wie sollen Duplikate behandelt werden?",
	"strategy_skip": "Duplikate überspringen",
	"strategy_skip_description": "Bereits vorhandene Assets bleiben unverändert, neue werden angelegt.",
	"strategy_upsert": "Bestehende aktualisieren",
	"strategy_upsert_description": "Bereits vorhandene Assets werden mit den CSV-Werten aktualisiert, neue werden angelegt.",
	"strategy_create": "Nur anlegen (strikt)",
	"strategy_create_description": "Der Import schlägt fehl, wenn ein Asset bereits existiert.",
	"clear_empty_cells": "Leere Zellen löschen bestehende Werte",
	"start_import": "Import starten",
	"result_title": "Import abgeschlossen",
	"result_created": "{count, plural, one {# Asset angelegt} other {# Assets angelegt}}",
	"result_updated": "{count, plural, one {# Asset aktualisiert} other {# Assets aktualisiert}}",
	"result_skipped": "{count, plural, one {# Zeile übersprungen} other {# Zeilen übersprungen}}",
	"row_issues": "{count, plural, one {# Hinweis} other {# Hinweise}}",
	"row_prefix": "Zeile {row}:",
	"issue_duplicate": "Asset existiert bereits",
	"issue_conflicting_match": "Interne und externe ID treffen verschiedene Assets",
	"issue_parent_not_found": "Übergeordnetes Asset nicht gefunden",
	"issue_invalid_row": "Ungültige Zeile"
}
```

English mirror:

```json
{
	"mapping_title": "Map columns",
	"mapping_description": "Map your CSV columns to asset fields. Unmapped columns are ignored.",
	"ignore_column": "Ignore",
	"name_mapping_required": "The name column must be mapped.",
	"strategy_title": "How should duplicates be handled?",
	"strategy_skip": "Skip duplicates",
	"strategy_skip_description": "Existing assets stay untouched, new ones are created.",
	"strategy_upsert": "Update existing",
	"strategy_upsert_description": "Existing assets are updated with the CSV values, new ones are created.",
	"strategy_create": "Create only (strict)",
	"strategy_create_description": "The import fails if an asset already exists.",
	"clear_empty_cells": "Empty cells clear existing values",
	"start_import": "Start import",
	"result_title": "Import finished",
	"result_created": "{count, plural, one {# asset created} other {# assets created}}",
	"result_updated": "{count, plural, one {# asset updated} other {# assets updated}}",
	"result_skipped": "{count, plural, one {# row skipped} other {# rows skipped}}",
	"row_issues": "{count, plural, one {# issue} other {# issues}}",
	"row_prefix": "Row {row}:",
	"issue_duplicate": "Asset already exists",
	"issue_conflicting_match": "Internal and external ID match different assets",
	"issue_parent_not_found": "Parent asset not found",
	"issue_invalid_row": "Invalid row"
}
```

Also check `common.back` / `common.continue` / `common.close` exist (grep first; they almost certainly do).

**JSON edit caution:** insert keys with a node script as done previously, and inspect `git diff` — these files contain duplicate keys that materialize away on rewrite (runtime-neutral but reviewable).

- [ ] **Step 7: Lint + typecheck + tests**

Run: `pnpm exec biome check --write src/modules/inventory/assets && pnpm exec tsc --noEmit && pnpm exec vitest run src/modules/inventory/assets`
Expected: clean / PASS.

- [ ] **Step 8: Manual walkthrough**

Dev server: import a CSV with custom headers → mapping auto-match → SKIP → result counts. Re-import the same file with UPSERT → updatedCount. CSV with `ParentExternalID` referencing another row of the same file → parent linked.

- [ ] **Step 9: Commit**

```bash
git add src/modules/inventory/assets messages/de/common.json messages/en/common.json
git commit -m "feat(assets): import wizard with column mapping, strategies and result report (#878)"
```

---

### Task 10: Data migration for existing "" IDs (backend, end of feature)

ONE consolidated migration at the very end, per project convention.

**Files:**
- Generated: `db/migrations/...` + `db/migrations-goose-postgres/...`

- [ ] **Step 1: Generate the migration**

Run: `./kopexa-gen migrate --name normalize-empty-asset-ids`

Add to the generated SQL (both migration dirs):

```sql
UPDATE assets SET internal_id = NULL WHERE internal_id = '';
UPDATE assets SET external_id = NULL WHERE external_id = '';
```

Check the actual column nullability first: `\d assets` (or the ent migration files) — if the columns are NOT NULL, the hook from Task 7 plus ent's optional-field handling determine whether NULL or absence is the right target; align the SQL with what `ClearInternalID()` produces.

- [ ] **Step 2: Verify against dev DB**

Run `kopexa serve` (auto-migrate) or apply the migration manually; verify duplicates with empty IDs no longer collide on import.

- [ ] **Step 3: Commit**

```bash
git add db/
git commit -m "fix(assets): clear legacy empty-string internal/external IDs (#878)"
```

---

### Task 11: Push + issue follow-up

- [ ] Push backend: `git push origin main` (kopexa)
- [ ] Push frontend: `git push origin main` (kopexa-frontend)
- [ ] Comment on kopexa-grc/kopexa#878 with the shipped behavior (strategies, mapping, parenting, root-cause fix) — Julian writes/approves the comment.
- [ ] Create follow-up issues: (a) async import via River/FileImport reusing the pipeline, (b) roll out pipeline adapters to People/Vendors, (c) surface the CREATE-strict row report through error extensions.

---

## Self-Review Notes

- **Spec coverage:** strategies/defaults (T2/T6/T9), empty-cell modes (T5/T9), column mapping (T3/T9), parenting incl. same-file refs (T6/T9), row report (T2/T6/T9), `""` root cause (T4 normalize, T7 hook, T10 migration), result UI (T9), backward compat (T6). Async + other entities explicitly out of scope (T11 follow-ups).
- **Known uncertainty, called out inline:** csvutil empty-cell behavior (T1 resolves), gqlgen enum constant casing (T2/T5/T6), sight Select/Radio API (T9), `Clear*` field names (T5), `asset.SpaceID` predicate (T6), header-return refactor of `unmarshalBulkDataMapped` (T6 step 2 note — implement there, adjust T3 tests).
- **Type consistency:** `rowAction`/`rowClassification` (T4) used in T6; `buildAssetUpdateInput(row, mode, presentColumns)` (T5) called in T6; `unmarshalBulkDataMapped` (T3, refactored in T6 to also return header).
