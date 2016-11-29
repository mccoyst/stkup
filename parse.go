// © 2016 Steve McCoy under the ISC License. See LICENSE for details. 

package main

import (
	"bytes"
	"fmt"
	"io"
	"unicode"
)

type posScanner struct {
	r io.RuneScanner
	Rune int
	Line int
	prev rune
}

func (ps *posScanner) ReadRune() (r rune, size int, err error) {
	r, size, err = ps.r.ReadRune()
	if err != nil {
		return r, size, err
	}
	ps.prev = r
	ps.Rune++
	if r == '\n' {
		ps.Line++
	}
	return r, size, err
}

func (ps *posScanner) UnreadRune() error {
	err := ps.r.UnreadRune()
	if err != nil {
		return err
	}
	ps.Rune--
	if ps.prev == '\n' {
		ps.Line--
	}
	return nil
}	

type tokenType int8

const (
	tokenNone tokenType = iota
	tokenParabreak
	tokenOpen
	tokenClose
	tokenWord
)

type token struct {
	Type  tokenType
	Value string
}

type state func(io.RuneScanner) (state, token, bool, error)

type lexer struct {
	r io.RuneScanner
	s state
}

func NewLexer(r io.RuneScanner) *lexer {
	return &lexer{r, lexStart}
}

func (l *lexer) Next() (token, error) {
	for {
		s, t, done, err := l.s(l.r)
		l.s = s
		if err != nil || done {
			return t, err
		}
	}
}

func lexStart(rs io.RuneScanner) (state, token, bool, error) {
	r, _, err := rs.ReadRune()
	if err != nil {
		return lexStart, token{}, false, err
	}

	switch {
	case r == '{':
		return lexStart, token{Type: tokenOpen}, true, nil
	case r == '}':
		return lexStart, token{Type: tokenClose}, true, nil
	case r == '\n':
		return lexParabreak, token{}, false, nil
	case unicode.IsSpace(r):
		return lexStart, token{}, false, nil
	case unicode.IsPrint(r):
		return lexWord(r), token{}, false, nil
	default:
		return lexStart, token{}, true, fmt.Errorf("found a naughty character: %v", r)
	}
}

func lexParabreak(rs io.RuneScanner) (state, token, bool, error) {
	r, _, err := rs.ReadRune()
	if err != nil {
		return lexStart, token{}, false, err
	}

	if r == '\n' {
		return lexStart, token{Type: tokenParabreak}, true, nil
	}
	rs.UnreadRune()
	return lexStart, token{}, false, nil
}

func lexWord(r0 rune) state {
	var buf bytes.Buffer
	buf.WriteRune(r0)
	var lw state
	lw = func(rs io.RuneScanner) (state, token, bool, error) {
		r, _, err := rs.ReadRune()
		if err == io.EOF {
			return lexStart, token{Type: tokenWord, Value: buf.String()}, true, nil
		}
		if err != nil {
			return lexStart, token{}, false, err
		}

		if r != '{' && r != '}' && !unicode.IsSpace(r) && unicode.IsPrint(r) {
			buf.WriteRune(r)
			return lw, token{}, false, nil
		}

		rs.UnreadRune()
		return lexStart, token{Type: tokenWord, Value: buf.String()}, true, nil
	}
	return lw
}

type node interface {
	Emit(w io.Writer) error
}

type word string

func (s word) Emit(w io.Writer) error {
	return emitPsStrings(w, []string{string(s)})
}

type markup struct {
	Cmd string
	Args []string
}

func (m *markup) Emit(w io.Writer) error {
	if m.Cmd == "parabreak" {
		_, err := io.WriteString(w, "body_pad next_line\n")
		return err
	}

	if m.Cmd == "title" {
		_, err := io.WriteString(w, "head_pad next_line\nbody_font head_size selectfont\n")
		if err != nil {
			return err
		}
		err = emitPsStrings(w, m.Args)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "body_pad next_line\nbody_font body_size selectfont\n")
		return err
	}

	return fmt.Errorf("Unrecognized command: %s", m.Cmd)
}

func emitPsStrings(w io.Writer, ss []string) error {
	for _, s := range ss {
		_, err := io.WriteString(w, "(" + s + " ) wshow\n")
		if err != nil {
			return err
		}
	}
	return nil
}

func parse(name string, rs io.RuneScanner) ([]node, error) {
	r := &posScanner{r:rs}
	l := NewLexer(r)

	var doc []node
	var curmark *markup
	for {
		t, err := l.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%s:%d:%d: %v", name, r.Line, r.Rune, err)
		}

		switch t.Type {
		case tokenParabreak:
			if curmark == nil {
				doc = append(doc, &markup{Cmd:"parabreak"})
			}
		case tokenOpen:
			if curmark != nil {
				return nil, fmt.Errorf("%s:%d:%d: %v", name, r.Line, r.Rune, "Nested markup does not exist")
			}
			curmark = &markup{}
		case tokenClose:
			if curmark == nil {
				return nil, fmt.Errorf("%s:%d:%d: %v", name, r.Line, r.Rune, "Unopened markup")
			}
			doc = append(doc, curmark)
			curmark = nil
		case tokenWord:
			if curmark == nil {
				doc = append(doc, word(t.Value))
			} else if curmark.Cmd == "" {
				curmark.Cmd = t.Value
			} else {
				curmark.Args = append(curmark.Args, t.Value)
			}
		}
	}

	return doc, nil
}
