# Local Employee Store — Architect Notes

**Status:** Advisory — based on discussion on 2026-07-07

## Problem Statement

The app currently starts with an empty form every session. Save/load is manual via file dialogs — the user must remember where they saved JSON configs and re-load them each time. There is no way to manage multiple employees without juggling individual files.

The goal is an automatic local store: fill in data for all employees, edit later, generate timesheets — and have everything persist across app restarts without manual file management.

## Key Decisions

### 1. Storage location & format

**Decision:** Single JSON file in Fyne's app storage directory.

Fyne provides a stable, platform-appropriate storage path via `app.Storage().RootURI()` (requires the app ID in `FyneApp.toml`, which is already set as `com.example.kindergarten-dienstplan`). All data lives in one `store.json` file containing an array of employees and shared year configs.

Rationale: Zero new dependencies, reuses existing `json.Marshal`/`Unmarshal` code, human-readable for debugging. The data volume is small (< 50 employees, each a few KB) so re-serializing the whole file on each save is negligible.

Alternatives considered:
- SQLite/bbolt: overkill for this scale, adds significant binary size
- Fyne `Preferences` API: designed for simple key-value settings, not structured data

### 2. Data model — identity & keying

**Decision:** UUID-keyed, one entry per employee per year.

Each `Employee` gets a `ID string` field (UUID). "Maria Huber 2025" and "Maria Huber 2026" are separate entries. The store is a flat `[]Employee`.

Rationale: Matches the current model 1:1 (just add an ID field). Each entry is fully self-contained. Name renames are safe, collisions impossible. Duplication across years is acceptable — a "copy to next year" feature can address it later.

Alternatives considered:
- One entry per employee (year separate): complicates the flow because free periods are year-specific
- Name+Year as natural key: fragile on rename, possible collisions

### 3. UI paradigm

**Decision:** Sidebar list grouped by employee name with year sub-items, plus detail form.

The sidebar looks like:

```
▾ Freie Tage
    2025
    2026
▾ Maria Huber
    2025
    2026
▾ Thomas Berger
    2026
[+ Neu]
```

Clicking an item in the sidebar loads the corresponding form in the right panel. The main layout uses `container.NewHSplit`. Window width increases from 700px to ~900px to accommodate the sidebar.

Rationale: Classic master-detail pattern. Intuitive, scalable from 1 to 50 employees, provides a full overview at a glance.

### 4. Auto-save behavior

**Decision:** Explicit "Speichern" button + auto-save on employee switch and app close.

Save triggers:
1. User clicks "Speichern" in the detail form
2. User selects a different employee/year in the sidebar (current entry auto-saved)
3. App window closes (`w.SetCloseIntercept`)

Rationale: Balances convenience with predictability. The user won't lose work by switching or closing.

**Validation policy:** No validation on save or auto-save. The store may contain incomplete/invalid data — this is intentional to allow incremental editing. Validation only gates generation ("Dienstplan erstellen"), which is the only point where correctness has consequences.

**Close-intercept error handling:** If `store.Save` fails on app close, show an error dialog: "Speichern fehlgeschlagen: {error}. Trotzdem schließen?" with "Schließen" (close anyway) and "Abbrechen" (stay open). Only call `w.Close()` if the user confirms or if save succeeds.

### 5. Free periods — shared with per-employee overrides

**Decision:** Shared free periods per year (in a dedicated form), with per-employee override capability.

A new `YearConfig` holds the shared free periods (school holidays, Fenstertage) for a given year. Each employee can add override free periods on top. The employee form shows inherited shared periods (read-only) plus editable override rows.

At generation time, the generator merges: `YearConfig.FreePeriods` + `Employee.FreePeriodOverrides`.

Rationale: School holidays are identical for all employees in the same kindergarten. Entering them once per year eliminates tedious repetition across 10+ employees. Per-employee overrides allow flexibility for individual cases.

### 6. Import/export

**Decision:** Drop manual save/load entirely.

The local store is the only persistence mechanism. The old "Konfiguration speichern/laden" file dialog buttons are removed.

Rationale: Simplifies the UI and mental model. One clear persistence path. If data needs to be moved between machines, the user can copy the app storage directory.

### 7. Batch generation

**Decision:** Not implemented now — one employee at a time.

The existing per-employee "Dienstplan erstellen" flow is retained. Batch generation ("generate all for year X") can be added later.

## Architectural Recommendations

### Data model

Both `Employee` and `YearConfig` live in the `model` package (not `store`) to keep data types separate from persistence logic:

```go
// model/employee.go

type Employee struct {
    ID                  string       `json:"id"`
    Name                string       `json:"name"`
    Year                int          `json:"year"`
    UseEvenOdd          bool         `json:"use_even_odd"`
    WeekSchedule        WeekSchedule `json:"week_schedule"`
    EvenSchedule        WeekSchedule `json:"even_schedule"`
    OddSchedule         WeekSchedule `json:"odd_schedule"`
    FreePeriodOverrides []DateRange  `json:"free_period_overrides"`
}

type YearConfig struct {
    Year        int         `json:"year"`
    FreePeriods []DateRange `json:"free_periods"`
}
```

The `Employee` struct gains two changes:
- New `ID string` field (UUID, generated via `crypto/rand` + `fmt.Sprintf` — no new dependency)
- `FreePeriods` renamed to `FreePeriodOverrides` (employee-specific free days only)

The `Store` struct lives in the `store` package and imports `model`:

```go
// store/store.go

type Store struct {
    Employees   []model.Employee   `json:"employees"`
    YearConfigs []model.YearConfig `json:"year_configs"`
}
```

### New package: `store`

Handles all persistence. Takes a plain directory path (not `fyne.App`) to stay framework-agnostic and testable. The caller resolves `app.Storage().RootURI().Path()` and passes it in.

- `Load(dir string) (*Store, error)` — reads `store.json` from `dir` on startup; returns empty store if file doesn't exist
- `Save(dir string, s *Store) error` — writes full store to disk (atomic: write temp file, then `os.Rename`)
- `AddEmployee(s *Store, emp model.Employee) string` — generates UUID via `crypto/rand`, appends to store, auto-creates `YearConfig` for the employee's year if not present, returns the new ID
- `UpdateEmployee(s *Store, emp model.Employee)` — replaces by ID
- `DeleteEmployee(s *Store, id string)` — removes by ID (does not auto-delete orphaned `YearConfig`)
- `FindYearConfig(s *Store, year int) *model.YearConfig` — returns nil if not found (no side effects)
- `EnsureYearConfig(s *Store, year int) *model.YearConfig` — creates if absent, returns the entry
- `MergeFreePeriods(shared []model.DateRange, overrides []model.DateRange) []model.DateRange` — combines for generation

### UI structure

The main layout changes from a single scrollable `VBox` to:

```
HSplit
├── Left: Sidebar (tree/list with grouped employees + "Freie Tage" section)
└── Right: Detail panel (employee form OR year config form, depending on selection)
```

Key UI components:
- **Sidebar widget** — custom `VBox` with collapsible sections or Fyne `widget.Tree`. "Freie Tage" group at the top, then employee groups. Each group has year sub-items. "Neu" button at the bottom. `YearConfig` entries are auto-created when an employee for that year is added; deletion is explicit via a "Löschen" button in the year config form (with confirmation dialog).
- **Employee detail form** — refactored from current `BuildForm`. Removes save/load file dialog buttons. Adds "Speichern" and "Löschen" buttons. "Löschen" shows a confirmation dialog ("Mitarbeiter {Name} ({Year}) wirklich löschen?"). After deletion, the detail panel shows an empty placeholder: "Mitarbeiter auswählen oder erstellen." Shows inherited shared free periods (read-only) and editable override rows.
- **Year config form** — reuses the `HolidaysInput` widget for editing shared free periods.
- **Empty placeholder** — shown when no employee/year config is selected, or after deletion. Centered message: "Mitarbeiter auswählen oder erstellen." Same initial state when the app launches with an empty store.

### Generator changes

The `Generate` function signature changes to accept merged free periods as a separate parameter:

```go
func Generate(emp model.Employee, freePeriods []model.DateRange) (*excelize.File, error)
```

The generator uses `freePeriods` directly (passed to `freePeriodLabel`) and never reads `emp.FreePeriodOverrides`. The caller (UI layer) is responsible for merging shared + override periods via `store.MergeFreePeriods` before calling `Generate`. This keeps the generator unaware of the store and the shared/override distinction.

**Required code change in `excel.go`:** Line 82 changes from `emp.FreePeriods` to the new `freePeriods` parameter. The `writeSheet` function also gains the `freePeriods` parameter.

### File layout after changes

```
timesheets/
├── main.go                        # passes fyne.App to UI builder
├── model/
│   ├── employee.go                # Employee + ID field, FreePeriodOverrides
│   └── config.go                  # date helpers (save/load config removed)
├── store/
│   └── store.go                   # Store struct, Load, Save, CRUD, merge logic
├── ui/
│   ├── app.go                     # top-level HSplit layout, wires sidebar + detail
│   ├── sidebar.go                 # grouped employee/year tree + actions
│   ├── employee_form.go           # employee detail form (refactored from form.go)
│   ├── year_config_form.go        # shared free periods form
│   ├── working_hours.go           # unchanged
│   └── holidays_input.go          # unchanged (reused in both forms)
├── generator/
│   ├── excel.go                   # Generate signature changes: adds freePeriods parameter
│   └── styles.go                  # unchanged
└── holidays/
    └── public.go                  # unchanged
```

## Risks and Open Items

### Data loss on crash
Since changes are not saved on every keystroke, a crash mid-edit loses unsaved work. This is acceptable for a desktop tool. The explicit save button and auto-save on switch/close mitigate most scenarios.

### Store file corruption
A partial write could corrupt `store.json`. **Mitigation:** write to a temp file first, then `os.Rename` to the final path (atomic on most filesystems).

### Fyne storage path stability
The app ID (`com.example.kindergarten-dienstplan`) in `FyneApp.toml` determines the storage path. Changing the ID would orphan the store file. The current ID uses `com.example` — this should be updated to a real identifier before distribution.

### Employee rename UX
Since employees are UUID-keyed, renaming is safe at the data level. But the sidebar needs to update its grouping when a name changes. The sidebar should re-render from the store on save.

### Future considerations
- **"Copy to next year"** action: duplicate an employee entry with a new year, pre-filling the same schedule
- **Batch generation:** generate all employees for a year in one action
- **Import from JSON:** if users have existing config files, a one-time import action could be added later

## Design Review Resolutions (2026-07-07)

The following issues were identified during design review and resolved:

| # | Finding | Resolution |
|---|---------|------------|
| F1 | `excel.go` marked "unchanged" but `emp.FreePeriods` breaks after rename | Acknowledged as required change. `Generate` signature updated to accept `freePeriods` parameter |
| F2 | Generator integration presented two contradictory approaches without choosing one | Resolved: `Generate(emp, freePeriods)` — separate parameter, no wrapper needed |
| F3 | `YearConfig` placed in `store` package, coupling UI to persistence layer | Moved to `model` package alongside `Employee` and `DateRange` |
| F4 | `store.Load/Save` takes `fyne.App`, coupling persistence to GUI framework | Changed to `store.Load(dir string)` / `store.Save(dir string, ...)` — caller resolves Fyne storage path |
| F5 | Auto-save on switch may persist invalid half-typed data | Save always without validation. Validation only gates generation. Incomplete data in store is intentional — allows incremental editing |
| F6 | `GetYearConfig` has surprising create-on-read side effect | Split into `FindYearConfig` (nil if not found) and `EnsureYearConfig` (creates if absent) |
| F7 | UUID generation dependency unspecified | Generated via `crypto/rand` + `fmt.Sprintf` — no new dependency |
| F8 | `SetCloseIntercept` error handling unspecified if save fails | Show error dialog with "Schließen" / "Abbrechen" options. Only close if user confirms or save succeeds |
| F9 | UI state after deleting the currently viewed employee unspecified | Show empty placeholder: "Mitarbeiter auswählen oder erstellen." Deletion requires confirmation dialog |
| F10 | `YearConfig` creation and deletion lifecycle unspecified | Auto-created when employee for that year is added. Explicit deletion only (no auto-delete of orphaned entries). Deletion requires confirmation dialog |
