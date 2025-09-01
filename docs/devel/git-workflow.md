# Git Workflow Guide for GoPCA Development

This guide describes the Git workflow for managing multiple versions of GoPCA simultaneously, enabling feature development for future versions while maintaining stable releases.

## Table of Contents
- [Overview](#overview)
- [Branch Structure](#branch-structure)
- [Workflows](#workflows)
- [Version Management](#version-management)
- [Best Practices](#best-practices)

## Overview

GoPCA uses a modified Git Flow strategy to manage:
- **Active development** for next minor/major version (v1.1.0)
- **Maintenance** of current stable version (v1.0.x)
- **Feature development** in isolated branches
- **Hotfixes** for critical production issues

### Key Principles
1. `main` branch always contains the latest stable release
2. `develop` branch contains features for the next release
3. `maintenance/v1.0.x` branch for current version bug fixes
4. All work happens in feature branches via Pull Requests

## Branch Structure

```
main                    # Latest stable release (currently v1.0.1)
├── develop            # Next version development (v1.1.0 features)
├── maintenance/v1.0.x # Current version maintenance
├── feature/*          # Feature branches from develop
├── bugfix/*           # Bug fix branches
└── hotfix/*           # Emergency fixes from main
```

### Branch Descriptions

#### `main`
- **Purpose**: Production-ready code
- **Contains**: Latest stable release
- **Protected**: Yes - no direct pushes
- **Merges from**: Release branches, hotfixes

#### `develop`
- **Purpose**: Integration branch for next release
- **Contains**: Completed features for v1.1.0
- **Protected**: Yes - requires PR
- **Merges from**: Feature branches

#### `maintenance/v1.0.x`
- **Purpose**: Bug fixes for current stable version
- **Contains**: v1.0.1 + patches
- **Protected**: Yes - requires PR
- **Creates**: v1.0.2, v1.0.3, etc.

#### Feature Branches
- **Naming**: `feature/<issue-number>-<description>`
- **Example**: `feature/415-data-export-formats`
- **Created from**: `develop`
- **Merged to**: `develop`

#### Bugfix Branches
- **Naming**: `bugfix/<issue-number>-<description>`
- **Example**: `bugfix/420-csv-parsing-error`
- **Created from**: `maintenance/v1.0.x` or `develop`
- **Merged to**: Source branch + cherry-pick if needed

#### Hotfix Branches
- **Naming**: `hotfix/<version>-<description>`
- **Example**: `hotfix/v1.0.2-critical-crash`
- **Created from**: `main`
- **Merged to**: `main` and `develop`

## Workflows

### Setting Up the Branch Structure

```bash
# One-time setup (after v1.0.1 release)
git checkout main
git pull origin main

# Create develop branch for v1.1.0
git checkout -b develop
git push -u origin develop

# Create maintenance branch for v1.0.x
git checkout main
git checkout -b maintenance/v1.0.x
git push -u origin maintenance/v1.0.x
```

### Feature Development (v1.1.0)

```bash
# 1. Start from develop
git checkout develop
git pull origin develop

# 2. Create feature branch
git checkout -b feature/415-data-export-formats

# 3. Implement feature
# ... make changes ...
git add .
git commit -m "feat: add support for Excel export format"

# 4. Push and create PR
git push -u origin feature/415-data-export-formats
gh pr create --base develop --title "feat: Excel export support"

# 5. After PR approval and merge
git checkout develop
git pull origin develop
git branch -d feature/415-data-export-formats
```

### Bug Fix for Current Version (v1.0.x)

```bash
# 1. Start from maintenance branch
git checkout maintenance/v1.0.x
git pull origin maintenance/v1.0.x

# 2. Create bugfix branch
git checkout -b bugfix/421-export-crash

# 3. Fix the bug
# ... make changes ...
git add .
git commit -m "fix: prevent crash when exporting empty dataset"

# 4. Push and create PR
git push -u origin bugfix/421-export-crash
gh pr create --base maintenance/v1.0.x --title "fix: export crash"

# 5. After merge, cherry-pick to develop if applicable
git checkout develop
git pull origin develop
git cherry-pick <commit-hash>
git push origin develop
```

### Creating a Patch Release (v1.0.2)

```bash
# 1. Ensure maintenance branch is ready
git checkout maintenance/v1.0.x
git pull origin maintenance/v1.0.x

# 2. Run tests
make test

# 3. Update version and create release branch
./scripts/prepare-release.sh v1.0.2

# 4. Push and create PR to main
git push -u origin release-v1.0.2
gh pr create --base main --title "Release v1.0.2"

# 5. After PR merge, create tag
git checkout main
git pull origin main
./scripts/release.sh v1.0.2

# 6. Update maintenance branch
git checkout maintenance/v1.0.x
git merge main
git push origin maintenance/v1.0.x
```

### Creating a Minor Release (v1.1.0)

```bash
# 1. Create release branch from develop
git checkout develop
git pull origin develop
git checkout -b release-v1.1.0

# 2. Final testing and version update
./scripts/prepare-release.sh v1.1.0

# 3. Push and create PR to main
git push -u origin release-v1.1.0
gh pr create --base main --title "Release v1.1.0"

# 4. After PR merge, create tag
git checkout main
git pull origin main
./scripts/release.sh v1.1.0

# 5. Merge back to develop
git checkout develop
git merge main
git push origin develop

# 6. Create new maintenance branch
git checkout main
git checkout -b maintenance/v1.1.x
git push -u origin maintenance/v1.1.x
```

### Emergency Hotfix

```bash
# 1. Create hotfix from main
git checkout main
git pull origin main
git checkout -b hotfix/v1.0.2-security-fix

# 2. Apply fix
# ... make changes ...
git add .
git commit -m "fix(security): patch SQL injection vulnerability"

# 3. Test thoroughly
make test

# 4. Update version
./scripts/prepare-release.sh v1.0.2

# 5. Create PR to main
git push -u origin hotfix/v1.0.2-security-fix
gh pr create --base main --title "Hotfix v1.0.2: Security patch"

# 6. After merge, update other branches
git checkout main
git pull origin main
./scripts/release.sh v1.0.2

# Update develop
git checkout develop
git merge main
git push origin develop

# Update maintenance if exists
git checkout maintenance/v1.0.x
git merge main
git push origin maintenance/v1.0.x
```

## Version Management

### Semantic Versioning

GoPCA follows [Semantic Versioning](https://semver.org/):

```
MAJOR.MINOR.PATCH
  │     │     └─ Bug fixes (backward compatible)
  │     └─────── New features (backward compatible)
  └───────────── Breaking changes
```

### Version Planning

| Branch | Version Range | Purpose | Example Changes |
|--------|--------------|---------|-----------------|
| `main` | Latest stable | Production | v1.0.1 |
| `develop` | Next minor/major | New features | v1.1.0-dev |
| `maintenance/v1.0.x` | v1.0.2+ | Bug fixes only | Crashes, security |
| `maintenance/v1.1.x` | v1.1.1+ | After v1.1.0 release | Future patches |

### Choosing Where to Fix Bugs

**Fix in `maintenance/v1.0.x` when:**
- Bug affects current users
- Fix is low risk
- No new features required
- Cherry-pick to `develop` after

**Fix in `develop` only when:**
- Bug is minor/cosmetic
- Fix requires refactoring
- Related to v1.1.0 features
- Can wait for next release

**Create hotfix when:**
- Critical security issue
- Data loss potential
- Widespread crashes
- Cannot wait for planned release

## Best Practices

### 1. Branch Naming Conventions

```bash
feature/issue-description    # feature/415-excel-export
bugfix/issue-description     # bugfix/420-csv-parsing
hotfix/version-description   # hotfix/v1.0.2-security
release-version              # release-v1.1.0
```

### 2. Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
feat: add Excel export functionality
fix: correct CSV parsing for quoted fields
docs: update Git workflow guide
test: add tests for export module
refactor: simplify data transformation logic
perf: optimize large dataset handling
chore: update dependencies
```

### 3. Pull Request Guidelines

**Title Format:**
```
<type>: <description> (#issue)
```

**Examples:**
- `feat: add Excel export support (#415)`
- `fix: CSV parsing error with commas (#420)`

**PR Description Template:**
```markdown
## Summary
Brief description of changes

## Related Issue
Fixes #415

## Testing
- [ ] Unit tests pass
- [ ] Manual testing completed
- [ ] Documentation updated

## Screenshots (if UI changes)
```

### 4. Cherry-Picking Best Practices

```bash
# Cherry-pick with original author preserved
git cherry-pick -x <commit-hash>

# If conflicts occur
git cherry-pick --continue  # After resolving
git cherry-pick --abort     # To cancel

# Verify the cherry-pick
git log --oneline -5
```

### 5. Keeping Branches Updated

```bash
# Update develop with latest from main
git checkout develop
git pull origin develop
git merge main
git push origin develop

# Update feature branch with develop
git checkout feature/my-feature
git rebase develop  # or merge if preferred
```

### 6. Branch Protection Rules

Configure on GitHub:
- `main`: Require PR, status checks, no force push
- `develop`: Require PR, status checks
- `maintenance/*`: Require PR, status checks

### 7. Release Checklist

Before any release:
- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version numbers consistent
- [ ] No uncommitted changes
- [ ] CI/CD checks pass

## Common Scenarios

### Scenario 1: Bug in Both v1.0.x and v1.1.0

```bash
# Fix in maintenance first
git checkout maintenance/v1.0.x
git checkout -b bugfix/425-memory-leak
# ... fix bug ...
git push -u origin bugfix/425-memory-leak
gh pr create --base maintenance/v1.0.x

# After merge, cherry-pick to develop
git checkout develop
git pull origin develop
git cherry-pick <commit-hash>
git push origin develop
```

### Scenario 2: Feature Only for v1.1.0

```bash
# Work directly from develop
git checkout develop
git checkout -b feature/430-new-algorithm
# ... implement feature ...
git push -u origin feature/430-new-algorithm
gh pr create --base develop
```

### Scenario 3: Critical Production Issue

```bash
# Hotfix from main
git checkout main
git checkout -b hotfix/v1.0.2-critical
# ... fix issue ...
./scripts/prepare-release.sh v1.0.2
git push -u origin hotfix/v1.0.2-critical
gh pr create --base main --title "URGENT: Hotfix v1.0.2"
```

## Troubleshooting

### Merge Conflicts

```bash
# During cherry-pick
git status  # See conflicted files
# ... resolve conflicts ...
git add .
git cherry-pick --continue
```

### Wrong Base Branch

```bash
# Change PR base branch
gh pr edit --base develop
```

### Accidental Direct Push (if pre-protection)

```bash
# Revert the push
git push --force-with-lease origin main:<previous-commit>
# Create proper PR
```

## Migration Plan

To transition from current single-branch to Git Flow:

1. **Week 1**: Create branch structure
2. **Week 2**: Update CI/CD for multiple branches
3. **Week 3**: Team training on new workflow
4. **Week 4**: Full implementation

## Resources

- [Git Flow Original](https://nvie.com/posts/a-successful-git-branching-model/)
- [GitHub Flow](https://guides.github.com/introduction/flow/)
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)

## Quick Reference Card

```bash
# Start new feature
git checkout develop && git pull
git checkout -b feature/XXX-description

# Fix bug in v1.0.x
git checkout maintenance/v1.0.x && git pull
git checkout -b bugfix/XXX-description

# Emergency fix
git checkout main && git pull
git checkout -b hotfix/vX.X.X-description

# After PR merge
git branch -d <branch-name>
git remote prune origin
```

---

*For release procedures, see [release-guide.md](release-guide.md)*  
*For general development, see the main [README](../../README.md)*