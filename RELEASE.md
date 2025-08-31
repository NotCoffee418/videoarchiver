# Release Process

This document explains how to create releases for Video Archiver using the automated GitHub Actions workflow.

## Creating a Release

1. **Prepare the release**:
   - Ensure all changes are merged to the main branch
   - Update version numbers in relevant files if needed
   - Test the application thoroughly

2. **Create and push a version tag**:
   ```bash
   # Create a new tag with semantic versioning
   git tag v1.2.3
   
   # Push the tag to trigger the release workflow
   git push origin v1.2.3
   ```

3. **Monitor the workflow**:
   - Go to the [Actions tab](https://github.com/NotCoffee418/videoarchiver/actions)
   - Watch the "Release" workflow progress
   - The workflow will build for Linux (amd64 + arm64) and Windows

4. **Release artifacts**:
   - Upon successful completion, check the [Releases page](https://github.com/NotCoffee418/videoarchiver/releases)
   - The following files will be automatically uploaded:
     - `videoarchiver-vX.X.X-linux-amd64.tar.gz`
     - `videoarchiver-vX.X.X-linux-arm64.tar.gz` 
     - `videoarchiver-vX.X.X-windows-installer.exe`

## Tag Format

The workflow triggers on tags matching the pattern `v*.*.*` (e.g., `v1.0.0`, `v2.1.5`, `v10.20.30`).

Examples of valid tags:
- `v1.0.0` ✅
- `v2.1.5` ✅
- `v10.20.30` ✅

Examples of invalid tags (won't trigger):
- `1.0.0` ❌ (missing 'v' prefix)
- `v1.0` ❌ (missing patch version)
- `release-1.0.0` ❌ (wrong format)

## Build Process

The workflow performs the following steps:

1. **Linux builds** (parallel):
   - Sets up Go 1.25 and Node.js 20
   - Installs Wails v2
   - Builds for amd64 and arm64 architectures
   - Creates compressed tarballs

2. **Windows build**:
   - Sets up Go 1.25 and Node.js 20 on Windows runner
   - Installs Wails v2
   - Builds with NSIS installer (`wails build -nsis`)
   - Generates .exe installer

3. **Release creation**:
   - Downloads all build artifacts
   - Creates a GitHub release with the tag
   - Uploads all platform binaries

## Troubleshooting

If the workflow fails:

1. Check the [Actions tab](https://github.com/NotCoffee418/videoarchiver/actions) for error logs
2. Common issues:
   - **Frontend build fails**: Check Node.js/npm dependencies in `frontend/package.json`
   - **Go build fails**: Verify Go version compatibility in `go.mod`
   - **Wails build fails**: Check `wails.json` configuration
   - **NSIS build fails**: Verify Windows installer configuration in `build/windows/`

3. Fix issues and create a new tag to retry the release