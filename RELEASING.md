# Release Process

This document describes how to create a new release for go-anki-deck.

## Automatic Release Process

Releases are automatically created when you push a new tag that follows semantic versioning (v*.*.*)

### Steps to Create a Release

1. **Update version in code if needed** (optional)

2. **Create and push a new tag:**
   ```bash
   # Create a new tag (replace X.Y.Z with your version)
   git tag -a v0.X.Y -m "Release v0.X.Y"
   
   # Push the tag to GitHub
   git push origin v0.X.Y
   ```

3. **GitHub Actions will automatically:**
   - Run all tests to ensure code quality
   - Generate a changelog from commit messages
   - Create a GitHub release with the changelog

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

### Pre-release Checklist

Before creating a release tag:
- [ ] All tests pass (`make test`)
- [ ] Code is properly formatted (`make fmt`)
- [ ] Linter passes (`make lint`)
- [ ] Documentation is up to date
- [ ] CHANGELOG is updated (if maintaining one manually)

### Manual Release (if needed)

If the automatic process fails, you can create a release manually:
1. Go to https://github.com/ezynda3/go-anki-deck/releases
2. Click "Create a new release"
3. Choose the tag you created
4. Add release notes