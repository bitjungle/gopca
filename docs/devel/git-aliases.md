# Git Aliases

A small set of productivity aliases that match the GoPCA workflow
(`feature branches → develop → main`; see
[git-workflow-simple.md](git-workflow-simple.md)).

These are **optional** developer conveniences. Nothing in the build or CI depends
on them.

## Installing

Pick one of the two approaches below.

### Option A — one command per alias

Run these once; they write to your **global** `~/.gitconfig`:

```bash
git config --global alias.st  "status -sb"
git config --global alias.co  "checkout"
git config --global alias.br  "branch"
git config --global alias.last "log -1 HEAD --stat"
git config --global alias.lg  "log --graph --abbrev-commit --decorate --date=relative --format=format:'%C(bold blue)%h%C(reset) %C(dim white)(%ar)%C(reset) %C(white)%s%C(reset) %C(dim white)- %an%C(reset)%C(auto)%d%C(reset)'"
git config --global alias.pushup "push -u origin HEAD"

# Start a feature branch from a freshly-pulled develop:
#   git feature 723-my-thing
git config --global alias.feature "!f() { git checkout develop && git pull && git checkout -b \"\$1\"; }; f"

# Sync develop from main after a release (matches the release guide):
git config --global alias.sync-develop "!git checkout develop && git fetch origin && git merge origin/main && git push"

# Delete local branches already merged into develop (conservative — see note below):
git config --global alias.cleanup "!git checkout develop >/dev/null 2>&1 && git branch --merged | grep -vE '^\\*|^\\+| (develop|main)\$' | xargs -r git branch -d && git remote prune origin"
```

### Option B — paste into `~/.gitconfig`

```ini
[alias]
    st      = status -sb
    co      = checkout
    br      = branch
    last    = log -1 HEAD --stat
    lg      = log --graph --abbrev-commit --decorate --date=relative --format=format:'%C(bold blue)%h%C(reset) %C(dim white)(%ar)%C(reset) %C(white)%s%C(reset) %C(dim white)- %an%C(reset)%C(auto)%d%C(reset)'
    pushup  = push -u origin HEAD
    feature = "!f() { git checkout develop && git pull && git checkout -b \"$1\"; }; f"
    sync-develop = "!git checkout develop && git fetch origin && git merge origin/main && git push"
    cleanup = "!git checkout develop >/dev/null 2>&1 && git branch --merged | grep -vE '^\\*|^\\+| (develop|main)$' | xargs -r git branch -d && git remote prune origin"
```

## What each alias does

| Alias | Purpose |
|-------|---------|
| `git st` | Short, branch-aware status. |
| `git co` / `git br` | Shorthand for `checkout` / `branch`. |
| `git last` | Show the most recent commit with its diffstat. |
| `git lg` | Compact, colored commit graph. |
| `git pushup` | Push the current branch and set its upstream in one step. |
| `git feature <name>` | Check out `develop`, pull, and create `<name>` — the standard way to start work. Use an issue-numbered name, e.g. `git feature 723-fix-thing`. |
| `git sync-develop` | Bring `develop` up to date with `main` after a release. |
| `git cleanup` | Delete local branches **merged into `develop`** and prune stale remote refs. **Conservative** — see the caveat below. |

## Cleaning up branches — the squash-merge caveat

GoPCA **squash-merges** pull requests. A squash merge replays the branch's
changes as a single *new* commit on `develop`, so the original branch is **not**
an ancestor of `develop`. That means `git branch --merged` (and therefore the
`git cleanup` alias above) will **not** list squash-merged branches — they linger
even though their work is fully in `develop`. This is how a working copy can
accumulate dozens of stale branches.

To delete local branches whose PR was actually **merged** (squash or not), match
against the merge state on GitHub instead of ancestry. This needs the
[`gh` CLI](https://cli.github.com/) and **bash** (it uses process substitution,
which is not POSIX `sh`), so keep it as a script rather than a git alias:

```bash
#!/usr/bin/env bash
# cleanup-merged.sh — delete local branches whose PR is merged on GitHub.
# Dry-run by default; pass --force to actually delete.
set -euo pipefail

git fetch origin --quiet

merged_pr_heads="$(gh pr list --state merged --limit 600 --json headRefName --jq '.[].headRefName' | sort -u)"
local_branches="$(git branch --format='%(refname:short)' | grep -vE '^(develop|main)$' | sort)"

# Branches that exist locally AND have a merged PR:
to_delete="$(comm -12 <(echo "$local_branches") <(echo "$merged_pr_heads"))"

if [[ -z "$to_delete" ]]; then
    echo "Nothing to clean up."
    exit 0
fi

if [[ "${1:-}" == "--force" ]]; then
    echo "$to_delete" | xargs -r -n1 git branch -D
    git remote prune origin
else
    echo "Would delete (run with --force to apply):"
    echo "$to_delete" | sed 's/^/  /'
fi
```

This deletes only branches whose name matches a **merged** pull request.
Work-in-progress branches, long-lived maintenance lines (`maintenance/*`), and
manual backups have no merged PR, so they are left untouched. The only
name-based exclusions are `develop` and `main`; a branch that happens to share a
name with a merged PR would still be removed.

> Local branch deletion is reversible: `git reflog` keeps the tips for ~90 days,
> and any branch that reached `develop`/`main` is preserved in that history.
> Capture `git rev-parse <branch>` first if you want an explicit restore point.
