# GoFindCallers

[Sublime Text 2][subl] plug-in for Golang, that adds IDE functionality to find all callers of selected [Golang][go] function.

[subl]: http://www.sublimetext.com/2
[go]: http://golang.org/

*Important this plugin uses the GOPATH.

Functionality includes:

- `GoFind callers` - Finds all function calls in current file as well as in all directories listed in `GOPATH`.
- Handles package import renaming.
- Distinguishes between `func` declarations and calls, as well as handle selector-expressions, setting the search parameters accordingly.

## Installation
`goFindCallers` is available via [Package Control][pkg-ctrl] and can be found as `goFindCallers`.

[pkg-ctrl]: http://wbond.net/sublime_packages/package_control

## Usage

The `GoFind Callers` command is accessible via the Command Palette, `Ctrl + Shift + P` on Windows/Linux, `Command + Shift + P` on OS X.

Or via the keyboard shortcut: `Ctrl + Alt + F` on all platforms.

After a search has been completed, search results will be displayed in "Find Results". Jump to the current file and line, by `Double clicking` or `Ctrl + Enter`, taking the cursor position.

## Requirements

- [Golang][go] v1.0 or higher
- Access to the `GOPATH`

[go]: http://golang.org/

### WORK in PROGRESS

Please bear in mind this is a work in progress, there will be bugs.
