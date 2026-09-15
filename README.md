# Stundenzettel-Generator

Desktop and Android app that generates yearly timesheets as Excel files.

## Download

Get the latest **Android** and **Windows** builds from
[GitHub Releases](https://github.com/MatousJobanek/timesheets/releases/latest).

| File | Platform |
| --- | --- |
| `Stundenzettel-Generator.apk` | Android (sideload) |
| `Stundenzettel-Generator-windows.zip` | Windows |

**Android:** allow installing from this source if asked. If you previously installed a debug-signed APK, uninstall it once before installing a GitHub build (different signing key). Updates from one GitHub version to the next should install over the previous GitHub APK.

**Windows:** SmartScreen may warn because the `.exe` is unsigned. Use “More info” → “Run anyway” if you trust the file.

Linux is not published; build locally with `make package-linux` or `make run`.

## Release (maintainers)

```bash
make release VERSION=1.0.1   # bumps FyneApp.toml, commits, tags v1.0.1 (no push)
# smoke-test on Linux
git push && git push --tags  # GitHub Actions attaches the APK and Windows zip
```

The first GitHub APK must use **Build ≥ 37** (the release target increments `Build` from `FyneApp.toml`).
