# Releasing

This project ships three platforms: Linux (x86-64), Windows (x86-64), and macOS (universal). Linux and Windows are built locally; macOS is built by GitHub Actions because it needs Xcode and can't be cross-compiled from Linux.

The final release is assembled manually with `gh release create` so partial CI failures can't publish a half-release.

## Prerequisites (one-time)

- A Fedora distrobox with the Wails Linux toolchain:
  ```bash
  distrobox create -i ghcr.io/ublue-os/fedora-toolbox:latest -n fedora
  distrobox enter fedora -- sudo dnf install -y webkit2gtk4.1-devel gtk3-devel gcc pkg-config nodejs
  distrobox enter fedora -- go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
- For Windows cross-compile (optional — skip if you don't plan to build Windows locally):
  ```bash
  distrobox enter fedora -- sudo dnf install -y mingw64-gcc mingw64-gcc-c++
  ```
- `gh` CLI on the host, authenticated as someone with push rights to `waveheadreport/genetica-resolutio`.

## Per-release recipe

All commands below run from the repo root on the host. `distrobox enter fedora -- ...` drops into the toolbox for build steps.

### 1. Update the app version

Bump `wails.json` → `info.productVersion` and any version string in the frontend (`APP_VERSION` in `frontend/src/main.js`). Commit.

### 2. Build Linux (native, via distrobox)

```bash
distrobox enter fedora -- bash -lc '
  cd ~/Downloads/genetica-resolutio-desktop &&
  PATH=$HOME/go/bin:$PATH wails build -tags webkit2_41 -o genetica-resolutio-linux-amd64
'
mkdir -p /tmp/artifacts
tar -czf /tmp/artifacts/genetica-resolutio-linux-amd64.tar.gz \
    -C build/bin genetica-resolutio-linux-amd64
```

### 3. Build Windows (cross-compile, optional)

```bash
distrobox enter fedora -- bash -lc '
  cd ~/Downloads/genetica-resolutio-desktop &&
  CC=x86_64-w64-mingw32-gcc PATH=$HOME/go/bin:$PATH \
    wails build -platform windows/amd64 -o genetica-resolutio-windows-amd64.exe
'
(cd build/bin && zip /tmp/artifacts/genetica-resolutio-windows-amd64.zip genetica-resolutio-windows-amd64.exe)
```

If you don't have mingw set up, skip this step — the release can ship without Windows, or you can use `workflow_dispatch` later to add a Windows CI job.

### 4. Build macOS (via CI)

Push a version tag. This triggers `.github/workflows/release.yml` which runs on `macos-latest` and uploads a single artifact named `macos-build`.

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

Wait for the run to finish (typically ~3 min), then pull the artifact:

```bash
RUN_ID=$(gh run list --workflow=release.yml --limit 1 --json databaseId --jq '.[0].databaseId')
gh run watch $RUN_ID --exit-status
gh run download $RUN_ID -n macos-build -D /tmp/artifacts
```

### 5. Generate checksums

```bash
cd /tmp/artifacts && sha256sum *.tar.gz *.zip | sort > SHA256SUMS.txt
```

### 6. Publish the release

```bash
gh release create vX.Y.Z \
  --repo waveheadreport/genetica-resolutio \
  --title "Genetica Resolutio vX.Y.Z" \
  --notes-file RELEASE_NOTES.md \
  /tmp/artifacts/*.tar.gz /tmp/artifacts/*.zip /tmp/artifacts/SHA256SUMS.txt
```

Use the v1.0.0 release body as a template for `RELEASE_NOTES.md` — keep the unsigned-binary warnings and the download table up to date.

## Known pitfalls

- **macOS runner name.** GitHub retired `macos-13`. The workflow uses `macos-latest` (currently Apple Silicon) and builds a `darwin/universal` binary so one artifact serves both architectures.
- **Wails bundle name.** `wails build -o NAME` renames the binary *inside* the `.app`, but the bundle folder is always `<productName>.app` from `wails.json` (i.e. `Genetica Resolutio.app`). The `ditto` step in the workflow packages exactly that path.
- **Linux apt occasionally hangs on GitHub runners.** If you ever restore Linux to CI, expect intermittent 20+ min stalls on `libwebkit2gtk-4.1-dev` installs. Local builds are faster and more reliable.
- **Don't force-push a release tag.** If a tag needs redoing, delete the release first (`gh release delete`), then delete and recreate the tag. Force-pushing a tag that a release points at leaves the release pointing at an orphaned commit.

## Checklist before tagging

- [ ] `wails.json` → `productVersion` bumped
- [ ] `APP_VERSION` in `frontend/src/main.js` matches
- [ ] About dialog in both desktop and web versions shows the new version
- [ ] Changelog or commit log reviewed — no WIP commits on main
- [ ] Clean working tree (`git status`) so `git commit -am` doesn't sweep in stray files
