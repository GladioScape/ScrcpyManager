# SignPath.io Code Signing Setup for ScrcpyManager

This guide explains how to set up free code signing for ScrcpyManager using SignPath.io.

## Overview

SignPath.io provides **free code signing for open source projects**. This eliminates Windows SmartScreen warnings and ensures users can trust the executable.

## Prerequisites

1. **GitHub Repository**: Your project must be hosted on GitHub as a **public** repository
2. **GitHub Account**: You need a GitHub account with admin access to the repository
3. **SignPath Account**: Free for open source projects
4. **Open Source License**: Your project must have an open source license (MIT, Apache, GPL, etc.)

## Setup Steps

### 1. Create a GitHub Repository (if not already done)

```bash
# Initialize git repository
git init
git add .
git commit -m "Initial commit"

# Create repository on GitHub and push
git remote add origin https://github.com/GladioScape/ScrcpyManager.git
git branch -M main
git push -u origin main
```

### 2. Register for SignPath.io

1. Go to https://signpath.io
2. Click **"Sign up for free"**
3. Choose **"Open Source Project"**
4. Sign in with your GitHub account
5. Authorize SignPath to access your GitHub repositories

### 3. Create a SignPath Project

1. In SignPath dashboard, click **"Create Project"**
2. Enter project details:
   - **Project Name**: ScrcpyManager
   - **Project Type**: Open Source
   - **Repository**: Select your GitHub repository
3. Click **"Create"**

### 4. Configure Artifact Configuration

The artifact configuration file `.signpath/artifact-configuration.xml` has already been created in your project with this content:

```xml
<?xml version="1.0" encoding="utf-8" ?>
<artifact-configuration xmlns="http://signpath.io/artifact-configuration/v1">
  <pe-file path="scrcpy-manager.exe" 
           product-name="ScrcpyManager" 
           product-version="1.0.0">
    <authenticode-sign />
  </pe-file>
</artifact-configuration>
```

#### Configuration Options

You can add metadata constraints to verify the PE file before signing:

```xml
<?xml version="1.0" encoding="utf-8" ?>
<artifact-configuration xmlns="http://signpath.io/artifact-configuration/v1">
  <pe-file path="scrcpy-manager.exe" 
           product-name="ScrcpyManager" 
           product-version="1.0.0"
           file-version="1.0.0.0"
           company-name="GladioScape"
           copyright="Copyright (c) 2024 GladioScape"
           original-filename="scrcpy-manager.exe">
    <authenticode-sign />
  </pe-file>
</artifact-configuration>
```

Commit and push this file to your repository:

```bash
git add .signpath/artifact-configuration.xml
git commit -m "Add SignPath artifact configuration"
git push
```

### 5. Set Up GitHub Actions Workflow

Create a GitHub Actions workflow to automatically build and sign your releases. Create `.github/workflows/release.yml`:

```yaml
name: Build and Sign Release

on:
  release:
    types: [created]
  workflow_dispatch:

jobs:
  build-and-sign:
    runs-on: windows-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.21'
        
    - name: Build executable
      run: go build -o scrcpy-manager.exe
      
    - name: Submit to SignPath
      uses: signpath/github-action-submit-signing-request@v1
      with:
        api-token: ${{ secrets.SIGNPATH_API_TOKEN }}
        organization-id: ${{ secrets.SIGNPATH_ORGANIZATION_ID }}
        project-slug: 'ScrcpyManager'
        signing-policy-slug: 'release-signing'
        artifact-configuration-slug: 'default'
        github-artifact-id: ''
        input-artifact-path: 'scrcpy-manager.exe'
        output-artifact-path: 'scrcpy-manager-signed.exe'
        wait-for-completion: true
        
    - name: Upload signed executable
      uses: actions/upload-artifact@v4
      with:
        name: scrcpy-manager-signed
        path: scrcpy-manager-signed.exe
        
    - name: Upload to Release
      if: github.event_name == 'release'
      uses: softprops/action-gh-release@v1
      with:
        files: scrcpy-manager-signed.exe
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 6. Configure GitHub Secrets

1. Go to your GitHub repository
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Click **"New repository secret"**
4. Add these secrets (get values from SignPath dashboard):

   - **SIGNPATH_API_TOKEN**: Your SignPath API token
   - **SIGNPATH_ORGANIZATION_ID**: Your SignPath organization ID

To get these values:
- In SignPath dashboard, go to **Settings** → **API Tokens**
- Create a new API token for CI/CD
- Copy the token and organization ID

### 7. Create a Release

1. Build your executable locally:
   ```bash
   go build -o scrcpy-manager.exe
   ```

2. Create a GitHub release:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

3. Go to GitHub → **Releases** → **Create a new release**
4. Select the tag you just created
5. Publish the release

The GitHub Action will automatically:
- Build the executable
- Submit it to SignPath for signing
- Upload the signed executable as a release asset

### 8. Verify Signing

After the workflow completes:

1. Download the signed executable from GitHub releases
2. Right-click the file → **Properties** → **Digital Signatures**
3. You should see a valid signature from SignPath

Windows SmartScreen will no longer block the executable!

## Manual Signing (Alternative)

If you prefer to sign manually:

1. Build your executable:
   ```bash
   go build -o scrcpy-manager.exe
   ```

2. In SignPath dashboard:
   - Go to your project
   - Click **"Submit Signing Request"**
   - Upload `scrcpy-manager.exe`
   - Wait for signing to complete
   - Download the signed executable

## Troubleshooting

### "Artifact configuration not found"
- Ensure `.signpath/artifact-configuration.xml` is in your repository root
- Verify the file is committed and pushed to GitHub
- Check that the `path` attribute matches your executable name exactly

### "Signing request failed"
- Verify your API token is valid and not expired
- Check that the organization ID is correct
- Ensure you have permissions for the project

### "SmartScreen still shows warning"
- First few downloads may still show warnings while Microsoft builds reputation
- This is normal and will improve over time
- Users can click "More info" → "Run anyway"

### "PE file validation failed"
- Ensure metadata in artifact-configuration.xml matches your built executable
- Check product-name, product-version match what's embedded in the .exe
- Use optional metadata constraints only if you embed version info in your Go build

## Benefits of Code Signing

✅ **No SmartScreen warnings** for users  
✅ **Verified publisher identity**  
✅ **Tamper protection** - users can verify integrity  
✅ **Professional distribution**  
✅ **Free for open source projects**

## Resources

- SignPath Documentation: https://about.signpath.io/documentation
- Artifact Configuration: https://about.signpath.io/documentation/artifact-configuration
- GitHub Actions Integration: https://about.signpath.io/documentation/build-system-integration/github-actions
- Code Signing Best Practices: https://about.signpath.io/documentation/build-system-integration

## License Compliance

SignPath.io free tier requires:
- **Public GitHub repository** (private repositories require paid plan)
- Open source license (MIT, Apache, GPL, etc.)
- Non-commercial use

Make sure your project meets these requirements.

---

**Note**: The signing process typically takes 2-5 minutes. The first signing request may require manual approval by SignPath team (usually within 24 hours for open source projects).