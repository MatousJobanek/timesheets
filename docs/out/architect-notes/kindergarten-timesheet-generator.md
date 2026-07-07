# Kindergarten Timesheet Generator — Architect Notes

**Status:** Advisory — based on discussion on 2026-07-06

## Problem Statement

A cross-platform desktop application is needed that generates yearly timesheets in Microsoft Excel format for kindergarten employees in Austria. The application must run on Windows, macOS, and Linux, present a form-based UI for inputting employee data and working hours, and produce a properly formatted `.xlsx` file with one tab per month. The timesheets must respect Austrian public holidays, school holidays, Fenstertage, and support per-day working hours with optional even/odd week scheduling. Kindergarten policy requires a 30-minute break after 4 hours when total daily hours exceed 6 (this exceeds the legal minimum under § 11 AZG, which only requires a break before 6 consecutive hours).

This is a greenfield project written in Go.

## Key Decisions

### 1. GUI Framework — Fyne

**Why this matters:** The framework determines developer experience, look-and-feel, distribution complexity, and whether a second language (JavaScript) is needed.

**Options considered:**
- **Fyne** (fyne.io/fyne/v2) — Pure Go, built-in form/widget system, single binary, `fyne package` for cross-platform distribution. Requires a C compiler on all platforms (Xcode CLT on macOS, `build-essential` + `libgl1-mesa-dev` on Linux, MinGW on Windows).
- **Wails v2** — Go backend + web frontend (React/Vue/Svelte), uses OS native WebView, ~15MB binary. Single binary output but two codebases in two languages.
- **Wails v3** — Latest Wails with mobile support, still in alpha (v3.0.0-alpha.102).

**Decision:** Fyne. The UI requirements (text fields, dropdowns, checkboxes, a generate button) are squarely in Fyne's sweet spot. No web frontend complexity, single Go codebase, trivial cross-platform packaging. Wails would add meaningful complexity (maintaining Go + JS codebases) without proportional benefit for a form-based tool.

### 2. Excel Generation Library — Excelize

**Why this matters:** The library must support multi-sheet XLSX files with cell styling, formulas, and proper formatting.

**Options considered:**
- **Excelize** (`github.com/xuri/excelize/v2`) — Pure Go, BSD-3, 20k+ stars, actively maintained (v2.10.1, Feb 2026). Supports multiple sheets, styling, formulas, merged cells.
- **tealeg/xlsx** — Older, essentially unmaintained, fewer features.

**Decision:** Excelize. No realistic alternative with comparable features and maintenance.

### 3. Austrian Public Holidays — Computed Algorithmically

**Why this matters:** Incorrect holidays would produce wrong timesheets. The approach determines whether the app needs yearly updates.

**Options considered:**
- **Compute algorithmically** — Computus algorithm for Easter date, all movable holidays derive from it. 9 fixed-date holidays hardcoded.
- **Hardcode per year** — Manual table of dates, needs updates each year.

**Decision:** Compute algorithmically. The Computus algorithm is well-established (~20 lines of Go), and it means the app works for any year without maintenance.

**Austrian public holidays (13 total):**

Fixed dates:
- Neujahr (1. Jänner)
- Heilige Drei Könige (6. Jänner)
- Staatsfeiertag (1. Mai)
- Mariä Himmelfahrt (15. August)
- Nationalfeiertag (26. Oktober)
- Allerheiligen (1. November)
- Mariä Empfängnis (8. Dezember)
- Weihnachten (25. Dezember)
- Stefanitag (26. Dezember)

Easter-dependent (computed via Computus):
- Ostermontag (Easter + 1)
- Christi Himmelfahrt (Easter + 39)
- Pfingstmontag (Easter + 50)
- Fronleichnam (Easter + 60)

### 4. School Holidays and Fenstertage Input — Date Range List

**Why this matters:** School holidays and Fenstertage vary by Bundesland and by individual kindergarten, so they cannot be reliably auto-filled.

**Options considered:**
- **Date range list with add/remove** — User adds labeled date ranges (e.g., "Weihnachtsferien: 24.12 – 06.01").
- **Bundesland dropdown + overrides** — Auto-fill by state, then adjust. Requires maintaining state-specific data that changes yearly.
- **Free-text input** — User types dates line by line. Error-prone.

**Decision:** Date range list with add/remove UI. Right balance of usability and flexibility without requiring yearly data maintenance.

### 5. Excel Timesheet Layout

**Why this matters:** The layout is the primary deliverable — it must be clear, correct, and match what kindergarten administrators expect.

**Columns per monthly sheet:**

| Datum | Wochentag | Beginn 1 | Ende 1 | Beginn 2 | Ende 2 | Stunden | Anmerkung |

- When an employee works > 6 hours, the day is split into two time blocks with the 30-minute break implicit in the gap between Ende 1 and Beginn 2.
- When an employee works ≤ 6 hours, only Beginn 1 / Ende 1 are filled; Beginn 2 / Ende 2 are empty.
- No explicit "Pause" column — the break is conveyed by the gap.
- No Urlaub/Krankenstand columns.
- Each sheet has a header with employee name and month/year, and a footer with a SUM formula for total hours.

**Styling:**
- Weekends: light gray background, "Wochenende" in Anmerkung
- Public holidays: colored background, holiday name in Anmerkung
- School holidays / Fenstertage: colored background, period label in Anmerkung
- Headers: bold
- Column widths: auto-fitted

### 6. Working Hours Data Model

**Why this matters:** The model must accurately represent the scheduling flexibility kindergarten employees have — different hours per day, and optionally different schedules for even vs. odd weeks.

**Decision:**
- Per weekday (Mon–Fri): **start time** (e.g., "07:00") and **total minutes** (e.g., 480 for 8h, 260 for 4h 20min)
- Working hours use arbitrary minute granularity (not limited to full/half hours)
- The app computes end times and break splits from these inputs
- **Optional even/odd week mode:** when enabled, both start time and total minutes can differ between even (gerade) and odd (ungerade) weeks
- **Break rule:** if total minutes > 360 (6h), insert 30 min break after first 4 hours of work. This reflects kindergarten policy (the legal minimum under § 11 AZG only requires the break before 6 consecutive hours)
- **0-hour days:** if an employee does not work on a given weekday (TotalMinutes = 0), the row is left blank (only Datum and Wochentag filled). Some employees have full days off (e.g., work Mon–Thu only)

**Example computation for 8h (480 min) day starting at 07:00:**
- Beginn 1: 07:00, Ende 1: 11:00 (4 hours)
- 30 min break (implicit)
- Beginn 2: 11:30, Ende 2: 15:30 (4 hours)
- Stunden: 8:00

**Example computation for 4h 20min (260 min) day starting at 07:30:**
- Beginn 1: 07:30, Ende 1: 11:50
- No break (total ≤ 6h)
- Beginn 2 / Ende 2: empty
- Stunden: 4:20

### 7. Language — German

**Decision:** Both the application UI and the Excel output are entirely in German. Labels, weekday names, month names (Jänner, Februar, ...), holiday names, column headers — all German.

**Implementation note:** Go's `time` package returns English names. The codebase must include explicit lookup tables for Austrian German month names (Jänner, Februar, März, April, Mai, Juni, Juli, August, September, Oktober, November, Dezember) and weekday names (Montag, Dienstag, Mittwoch, Donnerstag, Freitag, Samstag, Sonntag). Note "Jänner" (not "Januar") is the Austrian convention.

### 8. Distribution — Platform Packages via `fyne package`

**Why this matters:** Affects how easily users can install and run the app.

**Options considered:**
- **Raw binary** (`go build`) — functional but no icon, macOS Gatekeeper warnings.
- **Platform packages** (`fyne package`) — `.app` bundle for macOS, `.exe` with embedded icon for Windows, tarball for Linux. Essentially zero extra work (one command + an icon PNG).

**Decision:** Platform packages via `fyne package`. The effort difference is negligible (one command), and the result is more polished. macOS code signing is not needed for internal/known-group distribution.

### 9. Persistence — Save/Load Employee Configs to JSON

**Why this matters:** Filling in all employee data (working hours for 5 days, possibly even/odd weeks, school holiday date ranges) is a significant amount of input. Without persistence, users re-enter everything each time.

**Decision:** Save/Load to JSON files via native file dialogs. Users can maintain one config file per employee and reuse it across years (just change the year field). Implementation is straightforward — marshal/unmarshal a Go struct with `encoding/json`.

## Architectural Recommendations

### Technology Stack

- **Language:** Go
- **GUI:** Fyne v2 (`fyne.io/fyne/v2`)
- **Excel:** Excelize v2 (`github.com/xuri/excelize/v2`)
- **Packaging:** `fyne package`

### Project Structure

```
timesheets/
  main.go                  # Entry point, launches Fyne app
  ui/
    form.go                # Main form: employee name, year, working hours
    holidays_input.go      # Date range add/remove list for school holidays + Fenstertage
    working_hours.go       # Per-day start time + hours inputs, even/odd week toggle
  model/
    employee.go            # Employee struct, working hours, schedule config
    config.go              # JSON serialization for save/load
  holidays/
    public.go              # Austrian public holidays (Computus + fixed dates)
  generator/
    excel.go               # Excel file generation (one sheet per month)
    styles.go              # Cell styles (holiday colors, weekend gray, bold headers)
  go.mod
  go.sum
  Icon.png                 # App icon for fyne package
```

### Data Model

```go
type Employee struct {
    Name          string       `json:"name"`
    Year          int          `json:"year"`
    UseEvenOdd    bool         `json:"use_even_odd"`
    WeekSchedule  WeekSchedule `json:"week_schedule"`    // used when UseEvenOdd=false
    EvenSchedule  WeekSchedule `json:"even_schedule"`    // gerade Wochen
    OddSchedule   WeekSchedule `json:"odd_schedule"`     // ungerade Wochen
    FreePeriods   []DateRange  `json:"free_periods"`      // school holidays + Fenstertage
}

type WeekSchedule struct {
    Monday    DaySchedule `json:"monday"`
    Tuesday   DaySchedule `json:"tuesday"`
    Wednesday DaySchedule `json:"wednesday"`
    Thursday  DaySchedule `json:"thursday"`
    Friday    DaySchedule `json:"friday"`
}

// ForDay returns the DaySchedule for the given weekday.
// Returns a zero-value DaySchedule (TotalMinutes=0) for Saturday/Sunday.
func (ws WeekSchedule) ForDay(d time.Weekday) DaySchedule { ... }

type DaySchedule struct {
    StartTime    string `json:"start_time"`     // "07:00" (HH:MM format)
    TotalMinutes int    `json:"total_minutes"`  // 480 = 8h, 260 = 4h20m, 0 = not working
}

type DateRange struct {
    Label string `json:"label"`   // "Weihnachtsferien"
    From  string `json:"from"`    // "2026-12-24" (ISO 8601 in JSON)
    To    string `json:"to"`      // "2027-01-06" (ISO 8601 in JSON)
}
```

**Data model notes:**
- `TotalMinutes` uses integer minutes (not float hours) to avoid floating-point precision issues with arbitrary minute values like 4h 20min.
- `StartTime` is stored and validated as `HH:MM` format.
- `DateRange.From` / `DateRange.To` use ISO 8601 (`YYYY-MM-DD`) internally. The UI displays and accepts Austrian format (`DD.MM.YYYY`), with a conversion layer between UI and model.
- `TotalMinutes = 0` means the employee does not work on that weekday (row left blank in Excel).
- `WeekSchedule` uses named fields for readability in JSON config files. A `ForDay(time.Weekday)` helper method provides lookup by weekday for the generator.

### UI Layout (Fyne Form)

Organized in sections within a scrollable container:

1. **Mitarbeiterdaten** — Name (text entry), Jahr (number entry or dropdown)
2. **Arbeitszeiten** — Toggle "Gerade/Ungerade Wochen verwenden". Grid with Mon–Fri rows, columns for Beginn (HH:MM), Stunden, and Minuten (or a single HH:MM duration field). If even/odd enabled, two such grids labeled "Gerade Wochen" and "Ungerade Wochen"
3. **Freie Tage** — List of date ranges with add/remove buttons. Each row: Bezeichnung, Von (DD.MM.YYYY), Bis (DD.MM.YYYY). For single-day entries, Von = Bis
4. **Aktionen** — "Konfiguration speichern", "Konfiguration laden", "Dienstplan erstellen"

**Validation:** Basic validation runs when "Dienstplan erstellen" is clicked. Checks: name not empty, year valid, start times in HH:MM format, minutes >= 0, date formats valid (DD.MM.YYYY), From <= To in date ranges. On failure, an error dialog lists the problems.

### Excel Generation Flow

1. Create workbook with 12 sheets named "Jänner {year}" through "Dezember {year}"
2. For each sheet, write header (employee name, month/year) and column headers
3. Iterate each day of the month:
   - Classify as weekend, public holiday, school holiday/Fenstertag, or working day
   - For working days: look up the correct schedule (even/odd if enabled based on ISO week number), compute time blocks
   - Write row with appropriate values and styling
4. Write footer with SUM formula for Stunden column
5. Apply column widths and final styling

### Implementation Sequence

1. Project scaffolding — `go mod init`, dependencies, directory structure
2. Data model — structs with JSON tags
3. Holiday calculator — Computus algorithm + fixed dates
4. Excel generator — core logic with styling
5. Fyne UI — form with all sections
6. Save/Load — JSON persistence with file dialogs
7. Integration — wire Generate button to generator
8. Packaging — `fyne package` for all three platforms

## Risks and Open Items

1. **Fyne look-and-feel:** Fyne uses Material Design, which looks identical on all platforms but not "native." For a utility tool this is acceptable, but worth noting.

2. **Date input in Fyne:** Fyne does not have a built-in date picker widget. Date ranges for school holidays will be entered as text in DD.MM.YYYY format with validation on generate.

3. **Even/odd week definition:** ISO 8601 defines week numbers, but the user's definition of "even/odd" may differ (e.g., does week 1 count as odd?). The app should use ISO week numbers and document this clearly.

4. **Cross-compilation:** Fyne requires CGO on all platforms, which makes cross-compilation non-trivial. Options: build on each target platform natively, use `fyne-cross` (Docker-based), or use Zig as a cross-compiler. Consider CI/CD (GitHub Actions) with matrix builds if frequent releases are needed.

5. **Edge case — year boundary:** Some school holiday date ranges may span the year boundary (e.g., Weihnachtsferien 24.12.2026 – 06.01.2027). The generator must handle this correctly when processing January.

## Design Review Resolutions (2026-07-06)

The following issues were identified during design review and resolved:

| # | Finding | Resolution |
|---|---------|------------|
| F1 | Document incorrectly stated Fyne needs no CGO on Windows | Fixed: Fyne requires a C compiler on all platforms (MinGW on Windows) |
| F2 | Break placement at 4h is a policy choice, not the legal minimum | Confirmed as actual kindergarten policy. Documented distinction from § 11 AZG |
| F3 | WeekSchedule named fields vs map for easier iteration | Keep named struct for JSON readability; add `ForDay(time.Weekday)` helper |
| F4 | `float64` for hours causes precision issues with arbitrary minutes | Switched to `TotalMinutes int`. Arbitrary minute granularity confirmed |
| F5 | Date format ambiguity between UI (DD.MM.YYYY) and JSON (ISO 8601) | Documented: UI uses DD.MM.YYYY, JSON stores YYYY-MM-DD, conversion layer between them |
| F6 | Go `time` package returns English names, need Austrian German | Add lookup tables for month names (Jänner, ...) and weekday names (Montag, ...) |
| F7 | No input validation strategy | Basic validation on "Dienstplan erstellen" click, error dialog on failure |
| F8 | Unspecified behavior for 0-hour (non-working) days | `TotalMinutes=0` means not working; row left blank (only Datum + Wochentag) |
