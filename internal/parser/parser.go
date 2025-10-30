package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

// DSL is the root AST: a collection of cases
type DSL struct {
	Cases []*Case
}

type Case struct {
	Name string
	Body *Expr // root expression tree for the case
}

type Expr struct {
	Head string
	Args []*Expr
	Text string
}

// --------------------------------------------------------------------
// Public API
// --------------------------------------------------------------------

func ParseFile(path string) (*DSL, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(bytes.NewReader(data))
}

func Parse(r io.Reader) (*DSL, error) {
	tokens, err := tokenize(r)
	if err != nil {
		return nil, err
	}
	pos := 0
	parseExpr := func() (*Expr, error) {
		return readExpr(tokens, &pos)
	}
	var cases []*Case
	for pos < len(tokens) {
		expr, err := parseExpr()
		if err != nil {
			return nil, err
		}
		if expr.Head == "kyc-case" && len(expr.Args) > 0 {
			caseName := expr.Args[0].Head
			caseBody := &Expr{Head: "case-body", Args: expr.Args[1:]}
			cases = append(cases, &Case{Name: caseName, Body: caseBody})
		}
	}
	return &DSL{Cases: cases}, nil
}

// --------------------------------------------------------------------
// Tokenizer
// --------------------------------------------------------------------

func tokenize(r io.Reader) ([]string, error) {
	var tokens []string
	sc := bufio.NewScanner(r)
	sc.Split(bufio.ScanLines)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		// basic splitting on parentheses and whitespace
		var current strings.Builder
		for _, ch := range line {
			switch {
			case ch == '(' || ch == ')':
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}
				tokens = append(tokens, string(ch))
			case unicode.IsSpace(ch):
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}
			default:
				current.WriteRune(ch)
			}
		}
		if current.Len() > 0 {
			tokens = append(tokens, current.String())
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return tokens, nil
}

// --------------------------------------------------------------------
// Recursive Reader
// --------------------------------------------------------------------

func readExpr(tokens []string, pos *int) (*Expr, error) {
	if *pos >= len(tokens) {
		return nil, io.EOF
	}
	tok := tokens[*pos]
	*pos++
	switch tok {
	case "(":
		if *pos >= len(tokens) {
			return nil, fmt.Errorf("unexpected EOF after '('")
		}
		head := tokens[*pos]
		*pos++
		node := &Expr{Head: head}
		for *pos < len(tokens) && tokens[*pos] != ")" {
			child, err := readExpr(tokens, pos)
			if err != nil {
				return nil, err
			}
			node.Args = append(node.Args, child)
		}
		if *pos >= len(tokens) || tokens[*pos] != ")" {
			return nil, fmt.Errorf("missing closing ')'")
		}
		*pos++
		return node, nil
	case ")":
		return nil, fmt.Errorf("unexpected ')'")
	default:
		return &Expr{Head: tok}, nil
	}
}

// --------------------------------------------------------------------
// Utilities
// --------------------------------------------------------------------

func (e *Expr) String() string {
	if len(e.Args) == 0 {
		return e.Head
	}
	var parts []string
	for _, a := range e.Args {
		parts = append(parts, a.String())
	}
	return fmt.Sprintf("(%s %s)", e.Head, strings.Join(parts, " "))
}
