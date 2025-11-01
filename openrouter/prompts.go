package openrouter

// Default AI prompts for commit messages, branch names, and PR content
// These can be overridden by user-customized prompts in the config

const (
	// DefaultCommitPrompt generates a conventional commit message from git diff
	// The {diff} placeholder will be replaced with the actual git diff
	DefaultCommitPrompt = `Generate a conventional commit message for the following git diff.

Return ONLY valid JSON in this format (no markdown, no extra text):
{"subject": "...", "body": "..."}

Requirements:
- subject: Required. Max 72 characters. Follow conventional commits (feat:, fix:, refactor:, etc.)
- body: Optional. Max 500 characters. Explain why, not what. Use plain text (no code blocks).

Git diff:
{diff}`

	// DefaultBranchNamePrompt generates a semantic branch name from git diff
	// The {diff} placeholder will be replaced with the actual git diff
	DefaultBranchNamePrompt = `Generate a short, semantic git branch name for these changes.

Return ONLY the branch name (lowercase, kebab-case, max 40 characters). No explanations or markdown.

Examples: fix-login-bug, feat-dark-theme, refactor-api-client

Git diff:
{diff}`

	// DefaultPRPrompt generates a PR title and release notes style description from git diff
	// The {diff} placeholder will be replaced with the actual git diff
	DefaultPRPrompt = `Generate a pull request title and release notes style description for these changes.

You must respond with EXACTLY this format (no markdown, no extra text):
TITLE: [Your title here - max 72 characters, present tense]
DESCRIPTION: [Your release notes here - see format below]

Description Format:
## What's Changed

### Security & Fixes
- Brief user-facing description
- Another fix if applicable

### Improvements
- Enhancement description
- Another improvement if applicable

Important Guidelines:
- Use simple, user-friendly language (no technical jargon)
- Keep each item to ONE short line (max ~80 characters)
- Group changes logically by category
- Only include categories that have relevant changes
- Focus on user-facing benefits, not implementation details
- Skip internal refactoring or minor tweaks unless significant

Examples:
TITLE: Add dark mode support and improve performance
DESCRIPTION: ## What's Changed

### Improvements
- New dark mode theme with automatic system preference detection
- Reduced initial load time by optimizing image loading

TITLE: Fix critical security vulnerability
DESCRIPTION: ## What's Changed

### Security & Fixes
- Patched authentication bypass vulnerability in session handling
- Improved password validation to prevent weak credentials

Git diff:
{diff}`
)

// GetDefaultCommitPrompt returns the default commit message prompt
func GetDefaultCommitPrompt() string {
	return DefaultCommitPrompt
}

// GetDefaultBranchNamePrompt returns the default branch name prompt
func GetDefaultBranchNamePrompt() string {
	return DefaultBranchNamePrompt
}

// GetDefaultPRPrompt returns the default PR content prompt
func GetDefaultPRPrompt() string {
	return DefaultPRPrompt
}
