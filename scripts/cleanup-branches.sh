#!/bin/bash

# cleanup-branches.sh
# Script to identify and delete merged branches from the repository

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_BRANCH="${1:-master}"
DRY_RUN="${2:-false}"

echo -e "${BLUE}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║         Branch Cleanup Tool for Raqeem Repository        ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo -e "${RED}❌ Error: Not a git repository${NC}"
    exit 1
fi

# Fetch all branches
echo -e "${BLUE}📡 Fetching all branches...${NC}"
git fetch --all --prune

# Get current branch
CURRENT_BRANCH=$(git branch --show-current)
echo -e "${GREEN}📍 Current branch: ${CURRENT_BRANCH}${NC}"
echo -e "${GREEN}📍 Base branch: ${BASE_BRANCH}${NC}"
echo ""

# Find merged branches
echo -e "${BLUE}🔍 Finding branches merged into ${BASE_BRANCH}...${NC}"
MERGED_BRANCHES=$(git branch -r --merged origin/${BASE_BRANCH} | \
    grep -v "origin/${BASE_BRANCH}" | \
    grep -v "origin/HEAD" | \
    grep "origin/" | \
    sed 's/origin\///' | \
    tr -d ' ' || echo "")

# Count branches
if [ -z "$MERGED_BRANCHES" ]; then
    echo -e "${GREEN}✅ No merged branches found to delete${NC}"
    exit 0
fi

BRANCH_COUNT=$(echo "$MERGED_BRANCHES" | grep -c ^ || echo "0")
echo -e "${YELLOW}📊 Found ${BRANCH_COUNT} merged branches:${NC}"
echo ""
echo "$MERGED_BRANCHES" | while read -r branch; do
    if [ ! -z "$branch" ]; then
        # Get the last commit date for this branch
        LAST_COMMIT_DATE=$(git log -1 --format="%ci" origin/"$branch" 2>/dev/null || echo "N/A")
        echo -e "   ${YELLOW}•${NC} $branch (last commit: ${LAST_COMMIT_DATE})"
    fi
done
echo ""

# Confirm deletion
if [ "$DRY_RUN" = "true" ]; then
    echo -e "${BLUE}🔍 DRY RUN MODE: No branches will be deleted${NC}"
    exit 0
fi

# Check if we're doing this interactively
if [ -t 0 ]; then
    echo -e "${YELLOW}⚠️  WARNING: This will delete ${BRANCH_COUNT} branches from the remote repository!${NC}"
    read -p "Do you want to proceed? (yes/no): " CONFIRM
    
    if [ "$CONFIRM" != "yes" ]; then
        echo -e "${RED}❌ Aborted by user${NC}"
        exit 0
    fi
fi

# Delete branches
echo ""
echo -e "${BLUE}🗑️  Deleting merged branches...${NC}"
DELETED_COUNT=0
FAILED_COUNT=0

echo "$MERGED_BRANCHES" | while read -r branch; do
    if [ ! -z "$branch" ]; then
        echo -n "Deleting $branch... "
        if git push origin --delete "$branch" 2>/dev/null; then
            echo -e "${GREEN}✅${NC}"
            DELETED_COUNT=$((DELETED_COUNT + 1))
        else
            echo -e "${RED}❌${NC}"
            FAILED_COUNT=$((FAILED_COUNT + 1))
        fi
    fi
done

echo ""
echo -e "${BLUE}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                         Summary                           ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════╝${NC}"
echo -e "${GREEN}✅ Deleted: ${DELETED_COUNT} branches${NC}"
if [ $FAILED_COUNT -gt 0 ]; then
    echo -e "${RED}❌ Failed: ${FAILED_COUNT} branches${NC}"
fi
echo ""
echo -e "${GREEN}✨ Branch cleanup completed!${NC}"
