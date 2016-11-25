// © 2016 Steve McCoy under the ISC License. See LICENSE for details. 

package main

import (
	"bytes"
	"fmt"
	"io"
	"unicode"
)

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
		if err == io.EOF || done {
			return t, nil
		}
		if err != nil {
			return token{}, err
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
