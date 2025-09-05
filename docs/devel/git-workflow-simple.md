# Simplified Git Workflow for GoPCA

## Quick Start

**95% of your work follows this simple pattern:**

```bash
# 1. Start from develop
git checkout develop && git pull

# 2. Create feature branch with issue number
git checkout -b <issue-number>-<short-description>
# Example: git checkout -b 435-fix-markdown-tables

# 3. Make changes and commit
git add .
git commit -m "fix: markdown table rendering issue"

# 4. Push and create PR to develop
git push -u origin 435-fix-markdown-tables
gh pr create --base develop

# 5. After merge, clean up
git checkout develop && git pull
git branch -d 435-fix-markdown-tables
```

## Branch Structure

```
main                    # Production releases only
└── develop            # All development work
    └── feature/*      # Your work branches
```

## The Golden Rules

1. **ALWAYS branch from `develop`**
2. **ALWAYS PR to `develop`**
3. **NEVER push directly to main or develop**
4. **Use issue numbers in branch names**

## Common Tasks

### New Feature or Bug Fix

```bash
git checkout develop && git pull
git checkout -b 123-feature-name
# work...
git push -u origin 123-feature-name
gh pr create --base develop --title "feat: your feature" --body "Fixes #123"
```

### Update Your Branch

```bash
# If develop has new changes
git checkout develop && git pull
git checkout your-branch
git merge develop  # or rebase if you prefer
```

### Emergency Hotfix (Rare)

Only for critical production issues:

```bash
git checkout main && git pull
git checkout -b hotfix-critical-issue
# fix...
git push -u origin hotfix-critical-issue
gh pr create --base main --title "hotfix: critical issue"
# After merge, update develop
git checkout develop && git pull && git merge main && git push
```

## Branch Naming

- **Features/Bugs**: `<issue-number>-<description>`
  - ✅ `435-markdown-tables`
  - ✅ `420-csv-parsing`
  - ❌ `fix-stuff` (no issue number)
  - ❌ `my-feature` (no issue number)

## Commit Messages

Use conventional commits:
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation
- `test:` Tests
- `refactor:` Code refactoring
- `chore:` Maintenance

## PR Guidelines

**Title**: `<type>: <description> (#issue)`
- Example: `fix: markdown table rendering (#435)`

**Body**: 
- Reference issue: `Fixes #435` or `Closes #435`
- Brief summary of changes
- Test plan if applicable

## Cleanup

### After PR is merged
```bash
# Update local develop
git checkout develop && git pull

# Delete local branch
git branch -d your-branch-name

# Prune remote tracking
git remote prune origin
```

### Periodic cleanup (monthly)
```bash
# Delete all merged branches except main and develop
git branch --merged develop | grep -v -E "main|develop|\*" | xargs -r git branch -d

# See what remote branches exist
git branch -r

# Delete old remote branch (if yours)
git push origin --delete old-branch-name
```

## When Things Go Wrong

### Wrong base branch in PR?
```bash
gh pr edit --base develop
```

### Accidentally committed to develop?
```bash
# Create a branch from your changes
git checkout -b fix-branch
git push -u origin fix-branch

# Reset develop
git checkout develop
git reset --hard origin/develop

# Create PR from your branch
gh pr create --base develop
```

### Merge conflicts?
```bash
# Update your branch with latest develop
git checkout your-branch
git merge develop
# Resolve conflicts in your editor
git add .
git commit
git push
```

## The 80/20 Rule

80% of the time you only need:
1. `git checkout develop && git pull`
2. `git checkout -b 123-feature`  
3. `git push -u origin 123-feature`
4. `gh pr create --base develop`

Keep it simple. Don't overthink it.

## Why This Workflow?

- **Simple**: One main development branch (develop)
- **Clear**: All work goes through PRs to develop
- **Safe**: main is always stable for users
- **Clean**: Old branches get deleted after merge

## Questions?

- **Q**: When do changes reach main?
  - **A**: During releases (handled by maintainers)

- **Q**: What about maintenance branches?
  - **A**: We'll create them if needed, but 99% of work doesn't need them

- **Q**: Can I merge my own PR?
  - **A**: No, unless explicitly authorized

- **Q**: How often should I update from develop?
  - **A**: Before starting new work and if you have conflicts