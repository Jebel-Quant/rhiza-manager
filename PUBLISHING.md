# Publishing to VS Code Marketplace

This document explains how to publish the Rhiza Manager extension to the VS Code Marketplace.

## Prerequisites

Before you can publish, you need:

1. A **Visual Studio Marketplace publisher account**
   - Visit [Visual Studio Marketplace Publisher Portal](https://marketplace.visualstudio.com/manage)
   - Sign in with your Microsoft account
   - Create a publisher if you don't have one (use the publisher ID: `jebel-quant`)

2. A **Personal Access Token (PAT)** with marketplace publish permissions
   - Go to [Azure DevOps](https://dev.azure.com/)
   - Navigate to User Settings > Personal Access Tokens
   - Create a new token with:
     - Organization: All accessible organizations
     - Scopes: Marketplace > Manage
   - Copy the token (you won't be able to see it again)

3. Add the **VS_MARKETPLACE_TOKEN** secret to GitHub
   - Go to your repository settings
   - Navigate to Secrets and variables > Actions
   - Create a new repository secret named `VS_MARKETPLACE_TOKEN`
   - Paste your Personal Access Token as the value

## Publishing Methods

### Automated Publishing (Recommended)

The extension is configured to publish automatically when you push a version tag:

```bash
# Update version in package.json (e.g., to 0.0.2)
pnpm version patch  # or minor, or major

# Push the version tag
git push --follow-tags
```

The GitHub Actions workflow (`.github/workflows/publish.yml`) will automatically:
1. Build the extension
2. Run tests and linting
3. Package the extension
4. Publish to VS Code Marketplace

### Manual Publishing

If you need to publish manually:

```bash
# 1. Make sure you're on a clean branch
git status

# 2. Install dependencies
pnpm install

# 3. Build the extension
pnpm run build

# 4. Package the extension (creates .vsix file)
pnpm run package

# 5. Publish to marketplace (requires VSCE_PAT environment variable)
export VSCE_PAT=your-personal-access-token
pnpm run publish
```

## Testing Before Publishing

Always test the packaged extension before publishing:

```bash
# 1. Package the extension
pnpm run package

# 2. Install the .vsix file in VS Code
code --install-extension rhiza-manager-0.0.1.vsix

# 3. Test all features
# 4. Uninstall when done
code --uninstall-extension jebel-quant.rhiza-manager
```

## Version Management

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.0.0): Breaking changes
- **MINOR** (0.1.0): New features, backward compatible
- **PATCH** (0.0.1): Bug fixes, backward compatible

Update the version using pnpm:

```bash
pnpm version patch  # 0.0.1 -> 0.0.2
pnpm version minor  # 0.0.1 -> 0.1.0
pnpm version major  # 0.0.1 -> 1.0.0
```

Don't forget to update the `CHANGELOG.md` file with release notes!

## Troubleshooting

### "Publisher not found" error

Make sure the `publisher` field in `package.json` matches your publisher ID in the marketplace.

### "Personal Access Token is invalid" error

- Check that your PAT hasn't expired
- Verify it has the correct scopes (Marketplace > Manage)
- Make sure the secret is correctly set in GitHub Actions

### Package validation errors

Run `pnpm run package` locally to see detailed validation errors before pushing.

### Images not showing in marketplace

- Only PNG, JPG, and GIF formats are supported (no SVG)
- Make sure images are committed to the repository
- Check that image paths in README.md are relative

## Resources

- [VS Code Extension Publishing](https://code.visualstudio.com/api/working-with-extensions/publishing-extension)
- [vsce CLI Documentation](https://github.com/microsoft/vscode-vsce)
- [VS Code Marketplace Publisher Portal](https://marketplace.visualstudio.com/manage)
