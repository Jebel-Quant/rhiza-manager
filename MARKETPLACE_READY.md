# VSCode Marketplace Publishing - Ready to Publish! ✅

## What Has Been Completed

This extension is now ready to be published to the VSCode Marketplace. All technical requirements have been met:

### ✅ Extension Icon
- Created a 128x128 PNG icon (`icon.png`)
- Added to `package.json` with the `icon` field
- Icon features a stylized "R" for Rhiza with repository indicators
- Successfully tested in package build

### ✅ Package.json Configuration
All required marketplace fields are present and properly configured:
- ✅ `name`: rhiza-manager
- ✅ `displayName`: Rhiza-Manager
- ✅ `description`: Manage Rhiza based repos
- ✅ `version`: 0.0.1
- ✅ `publisher`: jebel-quant
- ✅ `license`: MIT
- ✅ `icon`: icon.png
- ✅ `repository`: GitHub URL with HTTPS
- ✅ `bugs`: Issue tracker URL
- ✅ `homepage`: Repository homepage
- ✅ `keywords`: Relevant search terms
- ✅ `engines`: VSCode version requirement
- ✅ `categories`: Extension category

### ✅ Build & Package
- ✅ Extension builds successfully with esbuild
- ✅ Package script creates valid .vsix file
- ✅ All assets (icon, images, documentation) included in package
- ✅ Type checking passes
- ✅ Linting passes

### ✅ GitHub Actions Workflow
The `.github/workflows/publish.yml` workflow is configured to automatically publish when a version tag is pushed:
- ✅ Installs dependencies
- ✅ Runs type checking
- ✅ Runs linting
- ✅ Builds the extension
- ✅ Runs tests
- ✅ Publishes to marketplace

## What You Need to Do Next

To actually publish the extension to the VSCode Marketplace, follow these steps:

### Step 1: Create a Marketplace Publisher Account

1. Visit [Visual Studio Marketplace Publisher Portal](https://marketplace.visualstudio.com/manage)
2. Sign in with your Microsoft account
3. Create a publisher with ID: **`jebel-quant`** (this must match the `publisher` field in package.json)

### Step 2: Generate a Personal Access Token (PAT)

1. Go to [Azure DevOps](https://dev.azure.com/)
2. Navigate to **User Settings** > **Personal Access Tokens**
3. Click **"New Token"**
4. Configure:
   - Name: "VSCode Marketplace Publishing"
   - Organization: **All accessible organizations**
   - Scopes: **Marketplace** > **Manage**
5. Copy the token (you won't be able to see it again!)

### Step 3: Add Token to GitHub Secrets

1. Go to your repository on GitHub
2. Navigate to **Settings** > **Secrets and variables** > **Actions**
3. Click **"New repository secret"**
4. Name: `VS_MARKETPLACE_TOKEN`
5. Value: Paste your Personal Access Token from Step 2
6. Click **"Add secret"**

### Step 4: Publish the Extension

Once the secret is configured, you have two options:

#### Option A: Automated Publishing (Recommended)

Simply create and push a version tag:

```bash
# Make sure you're on the main branch
git checkout main
git pull

# Update version (this updates package.json and creates a git tag)
pnpm version patch  # 0.0.1 -> 0.0.2

# Push the tag (triggers the workflow)
git push --follow-tags
```

The GitHub Actions workflow will automatically build, test, and publish your extension!

#### Option B: Manual Publishing

If you prefer to publish manually:

```bash
# Install dependencies
pnpm install

# Build the extension
pnpm run build

# Set your PAT as an environment variable
export VSCE_PAT=your-personal-access-token-here

# Publish
pnpm run publish
```

## Monitoring the Publication

### Automated Publishing
- Watch the GitHub Actions workflow run at: https://github.com/Jebel-Quant/rhiza-manager/actions
- The workflow will show each step's progress
- If successful, your extension will appear on the marketplace within minutes

### Marketplace Page
Once published, your extension will be available at:
- https://marketplace.visualstudio.com/items?itemName=jebel-quant.rhiza-manager

## Testing Before Publishing

It's recommended to test the packaged extension locally before publishing:

```bash
# Package the extension
pnpm run package

# Install the .vsix file in VSCode
code --install-extension rhiza-manager-0.0.1.vsix

# Test all features in VSCode

# Uninstall when done testing
code --uninstall-extension jebel-quant.rhiza-manager
```

## Troubleshooting

### "Publisher not found" error
- Verify that you created a publisher with ID `jebel-quant` on the Marketplace portal
- Make sure the `publisher` field in `package.json` exactly matches your publisher ID

### "Personal Access Token is invalid" error
- Check that your PAT hasn't expired
- Verify it has the correct scope: **Marketplace** > **Manage**
- Ensure the `VS_MARKETPLACE_TOKEN` secret is correctly set in GitHub

### Package validation errors
- Run `pnpm run package` locally to see detailed validation messages
- All current tests pass, so this should not be an issue

## Additional Resources

- [VSCode Extension Publishing Documentation](https://code.visualstudio.com/api/working-with-extensions/publishing-extension)
- [vsce CLI Documentation](https://github.com/microsoft/vscode-vsce)
- [Marketplace Publisher Portal](https://marketplace.visualstudio.com/manage)

---

**Status**: ✅ Ready to publish! Just follow the steps above to make your extension live on the VSCode Marketplace.
