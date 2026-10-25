.PHONY: setup

# Run once after clone: enables the commit-msg hook (docs/git-rules.md §12.9).
setup:
	git config core.hooksPath .githooks
