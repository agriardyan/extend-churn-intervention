---
name: git-pr-descriptor
description: Generate comprehensive pull request descriptions by analyzing git diff between current branch and the primary branch (main/master). Use when creating PRs or when the user asks to describe changes, summarize commits, or generate PR descriptions.
allowed-tools: Bash(git *), Read
---

# Git Pull Request Descriptor

You are helping the user create a comprehensive pull request description by analyzing the git diff between the current branch and the primary branch (main or master).

## Step 1: Identify the Primary Branch

First, determine the primary branch of the repository:

1. Check if `main` exists: `git rev-parse --verify main`
2. If not, check if `master` exists: `git rev-parse --verify master`
3. Use whichever exists as the base branch

## Step 2: Get Current Branch Name

Run `git rev-parse --abbrev-ref HEAD` to get the current branch name.

## Step 3: Analyze the Git Diff

Run `git diff <primary-branch>...HEAD` to get the full diff between the primary branch and the current branch.

Analyze the diff to identify:
- **Files changed**: Count and categorize files (added, modified, deleted)
- **Lines changed**: Total additions and deletions
- **Key changes**: Major modifications by file or module
- **Patterns**: Common themes across changes (refactoring, new features, bug fixes, etc.)

## Step 4: Read Key Files for Context

Based on the diff analysis, read relevant files to understand:
- Purpose of new functions or classes
- Context of modifications
- Related code that wasn't changed but provides context
- Configuration changes and their implications

Focus on:
- New files (read entirely if under 200 lines)
- Modified files where changes are significant (>20% of file)
- README, CHANGELOG, or documentation updates

## Step 5: Analyze Commit Messages

Run `git log <primary-branch>..HEAD --oneline` to see commit messages in the branch.

Look for:
- Feature descriptions
- Bug fix references
- Breaking changes mentioned
- Issue or ticket numbers

## Step 6: Generate the PR Description

Create a comprehensive pull request description with these sections:

### **Title**
- Concise summary of the main change (50-70 characters)
- Use conventional commit style when appropriate (feat:, fix:, refactor:, etc.)

### **Overview**
- 2-3 sentence summary of what this PR accomplishes
- Explain the "why" behind the changes

### **Changes**
Break down changes by category:
- **Added**: New features, files, or functionality
- **Modified**: Changes to existing functionality
- **Removed**: Deprecated or deleted code
- **Fixed**: Bug fixes
- **Refactored**: Code improvements without behavior changes

### **Technical Details**
- Architecture changes
- New dependencies or configuration
- Database schema changes
- API changes
- Breaking changes (if any)

### **Files Changed**
List key files with brief descriptions:
```
path/to/file.ext - Description of changes
```

### **Testing**
- How the changes were tested
- New test files or test cases added
- Manual testing steps (if applicable)

### **Related Issues**
- Reference any issues, tickets, or PRs
- Use GitHub syntax: `Fixes #123`, `Closes #456`

### **Checklist**
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No breaking changes (or documented if present)
- [ ] Code follows project conventions
- [ ] Reviewed my own code

## Step 7: Present the Description

Format the final description in Markdown, ready to be copied into a pull request. You should write a markdown file with name format PR_DESCRIPTION_{latest-commit-hash}.md

## Additional Guidelines

- **Be specific**: Use concrete examples from the code
- **Highlight impact**: Explain how changes affect users or developers
- **Note risks**: Call out any potential issues or considerations
- **Keep it scannable**: Use bullet points, headings, and formatting
- **Include context**: Link to related documentation or discussions
- **Suggest reviewers**: If certain files or areas need specific expertise, mention it

## Example Output Format

```markdown
# [feat] Implement user authentication system

## Overview
This PR introduces a complete authentication system with JWT tokens, password hashing, and session management. It provides the foundation for user accounts and secured API endpoints.

## Changes

### Added
- JWT-based authentication middleware (`pkg/auth/jwt.go`)
- Password hashing utilities with bcrypt (`pkg/auth/password.go`)
- User session management service (`pkg/service/session.go`)
- Login and registration endpoints (`internal/handler/auth.go`)
- Authentication integration tests (`pkg/auth/auth_test.go`)

### Modified
- Server initialization to include auth middleware (`internal/server/grpc.go`)
- Configuration to support auth settings (`config/config.yaml`)
- User model with password and token fields (`pkg/models/user.go`)

## Technical Details

**Authentication Flow:**
1. User submits credentials to `/auth/login`
2. Password verified against bcrypt hash
3. JWT token generated with 24h expiry
4. Token stored in Redis for session tracking
5. Subsequent requests validated via middleware

**New Dependencies:**
- `github.com/golang-jwt/jwt/v5` for JWT handling
- `golang.org/x/crypto/bcrypt` for password hashing

```

## Notes

- If the diff is very large (>2000 lines), focus on the most significant changes and summarize the rest
- If there are many files, group them by directory or functionality
- For configuration files, explain the impact of changes
- For dependency updates, note security or compatibility implications
- Be honest about unknowns - if something is unclear from the diff, say so
