# Git Commit Browser

A simple terminal-based Git commit browser built with the [Bubble Tea](https://github.com/charmbracelet/bubbletea) library. This tool allows you to view and search through your Git commit history in a user-friendly interface.

## Features

- **View Commit History:** Browse through your Git commit history with a list view.
- **Search Commits:** Filter commits by hash or message using a fuzzy search.
- **Select and Display Commits:** Select a commit to view its hash in the interface.
- **Keyboard Navigation:** Use keyboard shortcuts to interact with the application.

## Installation

Ensure you have [Go](https://golang.org/doc/install) installed on your machine. Then, clone this repository and build the project:

```bash
git clone https://github.com/jitesh117/git-commit-browser.git
cd git-commit-browser
go build -o git-commit-browser
```

## Future enhancements

- [ ] Copy selected commit hash to clipboard.
- [ ] Create a git branch from the selected commit hash.
- [ ] Make it prettier
