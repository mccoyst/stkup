// © 2016 Steve McCoy under the ISC License. See LICENSE for details. 

package main

import (
	"io"
	"text/template"
	//	"unicode/utf8"
)

type layout struct {
	PageWidth, PageHeight            float64
	VerticalMargin, HorizontalMargin float64

	FontName          string
	BodySize, BodyPad float64
	LineSpace         float64

	HeadSize, HeadPad float64
}

func NewLayout(fontName string) *layout {
	l := new(layout)
	l.FontName = fontName
	l.PageWidth = 612
	l.PageHeight = 792
	l.VerticalMargin = 72
	l.HorizontalMargin = 144
	l.BodySize = 10
	l.BodyPad = l.BodySize * 1.75
	l.LineSpace = l.BodySize * 1.25
	l.HeadSize = 11
	l.HeadPad = l.HeadSize * 3

	return l
}

func (l *layout) LeftMargin() float64 {
	return l.HorizontalMargin
}

func (l *layout) RightMargin() float64 {
	return l.PageWidth - l.HorizontalMargin
}

func (l *layout) BottomMargin() float64 {
	return l.VerticalMargin
}

func (l *layout) TopMargin() float64 {
	return l.PageHeight - l.VerticalMargin
}

func (l *layout) Print(w io.Writer) error {
	t := template.Must(template.New("preamble").Parse(layoutPreamble))
	return t.Execute(w, l)
}

var layoutPreamble = `%!
% Letter = 8.5 x 11 in² = 612 x 792 pt²
/body_font /{{.FontName}} def
/page_width {{.PageWidth}} def
/page_height {{.PageHeight}} def
/top_margin {{.TopMargin}} def
/bottom_margin {{.BottomMargin}} def
/left_margin {{.LeftMargin}} def
/right_margin {{.RightMargin}} def

/body_size {{.BodySize}} def
/head_size {{.HeadSize}} def

/line_space {{.LineSpace}} def

/body_pad {{.BodyPad}} def
/head_pad {{.HeadPad}} def

% newline or padding
/next_line { left_margin exch currentpoint exch pop exch sub moveto } bind def

% init page
/next_page { newpath left_margin top_margin moveto } bind def

% wrap paragraphs to body bounds
/wshow { dup stringwidth pop currentpoint pop add right_margin gt { line_space next_line currentpoint exch pop bottom_margin lt { showpage next_page line_space next_line show } { show } ifelse } { show } ifelse } bind def

body_font body_size selectfont

next_page

`
