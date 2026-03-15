# mew — terminal Markdown viewer

mew is a small terminal utility that renders Markdown for comfortable reading in a terminal. It hides most Markdown markup, styles headers, lists and blockquotes, and highlights fenced code blocks with Chroma. Output is ANSI-coloured and paged when appropriate.

## Install

From source:

    cd $GOPATH/src/github.com/grimdork/mew
    go build ./...

Or (recommended) install to $GOBIN:

    go install github.com/grimdork/mew@latest

## Usage

Render one or more files:

    mew README.md
    mew docs/*.md

Read from stdin (useful for piping):

    cat README.md | mew

If run with no files and attached to a terminal, mew prints a short help message and exits (so it won't block waiting for stdin). When input is coming from a pipe, mew reads stdin as usual.

Examples:

    # view with default pager (respects $PAGER, falls back to `less -R`)
    mew README.md

    # disable paging (prints directly to stdout)
    mew -pager off README.md

    # write rendered output to a file (no pager)
    mew -pager off README.md > rendered.txt

## Flags

- `-style string` — Chroma style/theme for code blocks (default: "monokai").
- `-pager string` — Pager to use. Options: `auto` (default), `less`, `off`, or a command. `auto` will page when stdout is a TTY.

## Current features

- CommonMark parsing via Goldmark
- Syntax highlighting using Chroma
- Basic styling of headings, lists and blockquotes using termenv
- Pages output to `$PAGER` (defaults to `less -R`) when running interactively

## Known limitations / TODO

- Nested list indentation and complex list/blockquote combinations need refinement
- Tables are rendered as plain text (no column alignment yet)
- Inline code and link rendering could be improved (currently simple inline markers)
- Add CLI flags for width/wrapping and to disable highlighting for smaller binaries
- Tests and CI

Contributions, bug reports and suggestions welcome.
