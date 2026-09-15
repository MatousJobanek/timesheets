# GitHub Release Binaries — Architect Notes

**Status:** Advisory — based on discussion on 2026-09-15

## Problem Statement

The project lives on GitHub and is used as a Fyne GUI (`timesheets`, app ID `io.github.matousjobanek.stundenzettel`). There is no CI packaging today. Users need downloadable **Android** and **Windows** builds from GitHub. Day-to-day development is on **Linux**; that is enough to smoke-test features but does not validate Android `versionCode` updates.

The original design already assumed `fyne package` / `fyne-cross` (see kindergarten timesheet architect notes). This note records how to turn that into **tag-driven GitHub Releases** with a README download section.

Out of scope: Play Store, Windows Authenticode, macOS/Linux release artifacts, and rebranding (school vs kindergarten). Display-name changes are fine; **changing the Android app ID** would be a new app and would break in-place APK updates.

## Key Decisions

### 1. Platforms to ship

**Decision:** Android APK and Windows package only.

Rationale: Matches current and near-term users. Linux stays a local/dev distribution. macOS would add runners, signing, and Gatekeeper issues without a stated need.

Alternatives considered:
- Android-only: too narrow once Windows was required
- Full desktop matrix: extra CI cost and macOS signing for no current user

### 2. How a release is cut

**Decision:** Git tag `vX.Y.Z` creates a GitHub Release; CI attaches artifacts. Optional later: `workflow_dispatch` to rebuild the same tag if CI flakes. No nightlies on every `main` push.

Rationale: Tags are the usual contract for “this is the downloadable version.” Nightlies would confuse which file to install.

### 3. How binaries are built

**Decision:** GitHub Actions on Ubuntu with Docker **`fyne-cross`** for both Android and Windows.

Rationale: Fyne is CGO; Android needs the NDK. `fyne-cross` is the supported way to produce both targets from one Linux job. Switch Windows to `windows-latest` only if a cross-built `.exe` fails in real use.

Alternatives considered:
- Native Windows runner + Docker Android: two environments to debug
- Local `gh release upload`: not reproducible, easy to ship the wrong build

### 4. Android signing

**Decision:** Dedicated **release keystore** stored in GitHub Actions secrets (file + store password + alias + key password). Back up the keystore **outside** GitHub before the first signed CI APK. Not Play App Signing; GitHub sideload only.

Rationale: Android will not update an app signed with a different key. Debug signing is not a stable identity for a public Release page. Existing debug-signed installs will need a **one-time uninstall** when switching to the release key.

Windows binaries from this pipeline remain **unsigned**; SmartScreen warnings are expected unless a code-signing certificate is added later.

### 5. Version vs Build vs local Linux testing

**Decision:** Source of truth is `FyneApp.toml` in git. A small release command (`make release VERSION=1.1.0` or equivalent) sets `Version`, **increments `Build`**, commits, and creates tag `v` + Version. Developer smoke-tests that commit on Linux, then pushes commit and tag. CI packages the tagged tree **as-is** and **fails** if the tag does not match `Version`.

Rationale: Local `fyne package` / `fyne-cross` and CI then see the same metadata. Injecting version only in CI makes laptop builds lie. `Build` is Fyne’s Android **`versionCode`**: it must **increase** for an APK to replace an installed one. Linux does not use that rule; local Linux testing does not prove Android updates. Current metadata is `Version = "1.0.0"`, `Build = 36` — the **first GitHub APK must use Build ≥ 37** if devices already have 36.

Do **not** derive `Build` from `GITHUB_RUN_NUMBER`: rebuilding the same tag would look like a newer Android update. Same tag → same Version and Build.

### 6. How people find downloads

**Decision:** GitHub Releases plus a short README “Download” section linking `/releases/latest`. **Stable artifact filenames** (version in the Release title and inside the app, not in the file name). Direct `/releases/latest/download/<file>` links are optional later and only work if names stay fixed.

Rationale: First-time users need the two file names and sideload/SmartScreen notes. Releases-only is too easy to mis-click (source zip vs APK vs Windows).

Suggested names:
- `Stundenzettel-Generator.apk`
- `Stundenzettel-Generator-windows.zip` (or `.exe` if that is the single `fyne-cross` output)

Do not attach Linux packages to the Release if Linux is not a supported download.

## Architectural Recommendations

Fit this next to existing identity:

- `FyneApp.toml`: `Name`, `ID = "io.github.matousjobanek.stundenzettel"`, `Version`, `Build`
- `main.go`: same `appID` via `app.NewWithID`

Keep the app ID stable so GitHub APKs update existing installs (after the debug→release key cut-over).

### Release flow

```text
make release VERSION=1.1.0
  → bump FyneApp.toml Version + Build
  → commit, tag v1.1.0
  → smoke-test on Linux
  → git push && git push --tags

GitHub Actions (push tags v*)
  → fyne-cross android (keystore from secrets)
  → fyne-cross windows
  → rename to stable names
  → attach to GitHub Release for that tag

README
  → link .../releases/latest
  → name the two files; Android unknown sources; Windows SmartScreen
```

### CI sketch

- Trigger: `on.push.tags: ["v*"]`
- Runner: `ubuntu-latest` with Docker
- Pin `fyne-cross` (and be ready to adjust for Go 1.25 / Fyne v2.7.4)
- Secrets: keystore (e.g. base64), store password, key alias, key password
- Assert: tag `v1.1.0` equals `Version` in `FyneApp.toml`
- Publish: `softprops/action-gh-release` or `gh release upload` on that tag

Local Linux remains run-from-source / `fyne package -os linux`. Signing the APK locally is optional; CI is the signed Android source of truth.

### Implementation sequence

1. Generate keystore, configure Actions secrets, confirm an offline backup.
2. Add `.github/workflows/release.yml` as above.
3. Add release script/Make target (bump, commit, tag; push can stay manual).
4. README Download section.
5. First tag with **Build ≥ 37**; verify APK update (or clean install after key change) and Windows zip once.

If the GitHub repo is **private**, Release assets are not public downloads; make the repo public or accept that only people with repo access can download.

## Risks and Open Items

| Risk | Mitigation |
| --- | --- |
| Lost release keystore | Backup before first signed build; treat secrets as irreplaceable |
| Debug-signed apps already installed | Document one uninstall when the first CI-signed APK ships |
| `Build` not bumped | Release script + CI tag/`Version` check; first public Build ≥ 37 |
| Tag pushed without the version commit | Push the bump commit and the tag together |
| `fyne-cross` image vs Go 1.25 | Pin tool version; first workflow is the integration test |
| Windows SmartScreen | Document “More info → Run anyway”; cert later if needed |
| Private repository | Public repo or authenticated downloads |
| Rebrand / school vs kindergarten | Change display name only; do not change app ID if updates must continue |

Open items (not decided in this discussion): exact `fyne-cross` CLI flags for the keystore, whether the Windows artifact is zip vs exe, Make vs shell script for `release`, and whether `workflow_dispatch` is added in the first workflow or later.
