# Branch Cleanup Guide

This guide explains how to manage and clean up branches in the Raqeem repository.

## Overview

The repository includes automated and manual tools for cleaning up merged branches:

1. **Automated Cleanup** - GitHub Actions workflow that runs weekly
2. **Manual Cleanup** - Script for one-time cleanup operations

## Automated Branch Cleanup

### GitHub Actions Workflow

The repository includes a workflow (`.github/workflows/cleanup-merged-branches.yml`) that automatically deletes branches that have been merged into `master`.

**Schedule:** Runs every Sunday at 00:00 UTC

**Manual Trigger:**
1. Go to the "Actions" tab in GitHub
2. Select "Cleanup Merged Branches" workflow
3. Click "Run workflow"
4. Select the branch to run from (usually `master`)
5. Click "Run workflow"

### What Gets Deleted?

The workflow deletes branches that:
- Have been fully merged into `master`
- Are remote branches (in origin)
- Are not protected branches
- Are not `master` or `HEAD`

## Manual Branch Cleanup

### Using the Cleanup Script

For one-time cleanup or local management, use the provided script:

```bash
# Dry run - see what would be deleted without actually deleting
./scripts/cleanup-branches.sh master true

# Interactive mode - you'll be asked to confirm before deletion
./scripts/cleanup-branches.sh master

# Non-interactive mode (use with caution in scripts)
echo "yes" | ./scripts/cleanup-branches.sh master
```

### Script Parameters

- **Parameter 1:** Base branch (default: `master`)
- **Parameter 2:** Dry run mode (default: `false`, set to `true` for dry run)

### Examples

```bash
# Preview which branches would be deleted
./scripts/cleanup-branches.sh master true

# Delete merged branches (with confirmation prompt)
./scripts/cleanup-branches.sh master

# Delete branches merged into a different base branch
./scripts/cleanup-branches.sh main
```

## Manual Cleanup Using Git Commands

If you prefer using git commands directly:

```bash
# Fetch all branches and prune deleted ones
git fetch --all --prune

# List branches merged into master
git branch -r --merged origin/master | grep -v "origin/master" | grep -v "origin/HEAD"

# Delete a specific remote branch
git push origin --delete <branch-name>

# Delete multiple branches at once (use carefully!)
git branch -r --merged origin/master | \
  grep -v "origin/master" | \
  grep -v "origin/HEAD" | \
  sed 's/origin\///' | \
  xargs -I {} git push origin --delete {}
```

## Best Practices

### Branch Naming

Use descriptive branch names that indicate the purpose:
- `feature/add-new-api`
- `fix/authentication-bug`
- `copilot/improve-test-coverage`

### When to Delete Branches

Delete branches when:
- The pull request has been merged
- The feature/fix is no longer needed
- The branch is more than 90 days old and inactive

### When NOT to Delete Branches

Keep branches that:
- Are still under active development
- Contain work-in-progress features
- Serve as long-term integration points
- Are protected (like `master`, `main`, `develop`)

## Preventing Branch Accumulation

### GitHub Branch Protection

Configure branch protection rules in GitHub:
1. Go to Settings → Branches
2. Add rules for important branches (e.g., `master`)
3. Enable "Delete head branches automatically" in PR settings

### Delete Branches After Merge

When merging PRs, enable the option to automatically delete the head branch:
1. After PR is approved, click "Merge pull request"
2. Check the box "Delete branch" that appears after merging
3. The branch will be automatically deleted

### Configure Repository Settings

In GitHub repository settings:
1. Go to Settings → General
2. Scroll to "Pull Requests"
3. Enable "Automatically delete head branches"

This will automatically delete branches when PRs are merged.

## Troubleshooting

### Permission Denied

If you get a permission error when deleting branches:
- Ensure you have write access to the repository
- Check if the branch is protected
- Verify you're authenticated with GitHub

### Branch Already Deleted

If a branch was already deleted:
- Run `git fetch --all --prune` to update your local repository
- The script will skip already-deleted branches

### Cannot Delete Current Branch

You cannot delete the branch you're currently on:
- Switch to a different branch first: `git checkout master`
- Then run the cleanup script

## Current State

As of the last analysis, the repository has:
- **70+ branches** in the remote repository
- Most are `copilot/*` branches that have been merged
- Only 2 active branches: `master` and `copilot/delete-unused-branches`

Running the cleanup will significantly reduce the branch count and improve repository organization.

## Additional Resources

- [GitHub Docs: Managing Branches](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-branches)
- [Git Branching Best Practices](https://git-scm.com/book/en/v2/Git-Branching-Branching-Workflows)
