package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/grimdork/climate/arg"
	"github.com/muesli/termenv"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func main() {
	opt := arg.New("mew")
	opt.SetDefaultHelp(true)
	stylelist := make([]any, len(styles.Names()))
	for i, name := range styles.Names() {
		stylelist[i] = name
	}
	opt.SetOption(arg.GroupDefault, "s", "style", "Chroma style for code blocks.", "native", false, arg.VarString, stylelist)
	opt.SetOption(arg.GroupDefault, "p", "pager", "Pager to use.", "auto", false, arg.VarString,
		[]any{"auto", "less", "more", "off"})
	opt.SetFlag("Flags", "S", "show-styles", "Show available Chroma styles and exit.")
	err := opt.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error parsing arguments:", err)
		os.Exit(1)
	}

	paths := opt.Args

	if opt.GetBool("show-styles") {
		fmt.Println("Available Chroma styles:")
		for _, name := range stylelist {
			fmt.Println(" -", name)
		}
		return
	}

	var input []byte
	if len(paths) == 0 {
		// If stdin is a TTY, show help rather than block waiting for input.
		if isTerminal() {
			opt.PrintHelp()
			return
		}
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read stdin:", err)
			os.Exit(2)
		}
	} else {
		// concatenate files with separator
		var buf bytes.Buffer
		for i, fn := range paths {
			b, err := os.ReadFile(fn)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading %s: %v\n", fn, err)
				os.Exit(2)
			}
			if i > 0 {
				buf.WriteString("\n\n---\n\n")
			}
			buf.Write(b)
		}
		input = buf.Bytes()
	}

	outBuf := &bytes.Buffer{}
	println("Rendering markdown with style:", opt.GetString("style"))
	renderMarkdown(input, opt.GetString("style"), outBuf)

	// Pager logic
	usePager := false
	pagerCmd := "less -R"
	switch opt.GetString("pager") {
	case "off":
		usePager = false
	case "less":
		usePager = true
		pagerCmd = "less -R"
	case "more":
		usePager = true
		pagerCmd = "more"
	case "auto":
		if isTerminal() {
			usePager = true
			if p := os.Getenv("PAGER"); p != "" {
				pagerCmd = p
			}
		}
	}

	if usePager {
		parts := strings.Fields(pagerCmd)
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Stdin = bytes.NewReader(outBuf.Bytes())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "pager failed:", err)
			os.Exit(3)
		}
		return
	}

	_, _ = os.Stdout.Write(outBuf.Bytes())
}

func renderMarkdown(input []byte, chromaStyle string, w io.Writer) {
	md := goldmark.New()
	root := md.Parser().Parse(text.NewReader(input))

	// colour palette (hex RGB)
	cp := termenv.ColorProfile()
	h1col := cp.Color("#ff79c6")     // pink
	h2col := cp.Color("#8be9fd")     // cyan
	h3col := cp.Color("#50fa7b")     // green
	bulletCol := cp.Color("#f1fa8c") // yellow
	quoteCol := cp.Color("#6272a4")  // muted purple

	// basic traversal
	var walker func(n ast.Node, entering bool, level int)
	walker = func(n ast.Node, entering bool, level int) {
		switch node := n.(type) {
		case *ast.Document:
			// nothing
		case *ast.Heading:
			if entering {
				text := collectText(node, input)
				fmt.Fprintln(w, "")
				style := termenv.String(text).Bold()
				switch node.Level {
				case 1:
					style = style.Underline().Foreground(h1col)
				case 2:
					style = style.Foreground(h2col)
				default:
					style = style.Foreground(h3col)
				}
				fmt.Fprintln(w, style.String())
			}
		case *ast.Paragraph:
			if entering {
				text := collectText(node, input)
				fmt.Fprintln(w, wrapText(text))
				fmt.Fprintln(w, "")
			}
		case *ast.TextBlock:
			if entering {
				text := collectText(node, input)
				fmt.Fprintln(w, wrapText(text))
			}
		case *ast.FencedCodeBlock:
			if entering {
				lang := string(node.Language(input))
				code := string(node.Text(input))
				highlighted := highlightCode(code, lang, chromaStyle)
				fmt.Fprintln(w, highlighted)
				fmt.Fprintln(w, "")
			}
		case *ast.CodeBlock:
			if entering {
				code := string(node.Text(input))
				fmt.Fprintln(w, highlightCode(code, "", chromaStyle))
				fmt.Fprintln(w, "")
			}
		case *ast.Blockquote:
			if entering {
				text := collectText(node, input)
				fmt.Fprintln(w, termenv.String(text).Faint().Foreground(quoteCol).String())
				fmt.Fprintln(w, "")
			}
		case *ast.List:
			if entering {
				// iterate children
				for li := node.FirstChild(); li != nil; li = li.NextSibling() {
					line := collectText(li, input)
					bullet := termenv.String("-").Foreground(bulletCol).String()
					fmt.Fprintln(w, "  "+bullet+" "+line)
				}
				fmt.Fprintln(w, "")
			}
		}
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			walker(c, true, level+1)
		}
	}
	walker(root, true, 0)
}

func collectText(n ast.Node, src []byte) string {
	var buf strings.Builder
	ast.Walk(n, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if t, ok := node.(*ast.Text); ok && entering {
			buf.Write(t.Segment.Value(src))
		}
		if c, ok := node.(*ast.CodeSpan); ok && entering {
			val := string(c.Text(src))
			styled := termenv.String("`" + val + "`").Foreground(termenv.ColorProfile().Color("#ffb86c")).String()
			buf.WriteString(styled)
		}
		return ast.WalkContinue, nil
	})
	return buf.String()
}

func highlightCode(code, lang, styleName string) string {
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}
	formatter := formatters.TTY256
	var buf bytes.Buffer
	_ = formatter.Format(&buf, style, iterator)
	return buf.String()
}

func wrapText(s string) string {
	// naive: collapse runs of whitespace
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func isTerminal() bool {
	fd := int(os.Stdout.Fd())
	return terminalIsTTY(fd)
}

// minimal term detection to avoid importing x/term at top-level for portability
func terminalIsTTY(fd int) bool {
	file := os.Stdout
	stat, err := file.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}
