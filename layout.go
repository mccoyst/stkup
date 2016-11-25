// © 2016 Steve McCoy under the ISC License. See LICENSE for details. 

package main

import (
	"os"

	"io"
	"io/ioutil"
	"text/template"
//	"unicode/utf8"

	"github.com/golang/freetype/truetype"
)

type layout struct {
	font *truetype.Font

	PageWidth, PageHeight float64
	VerticalMargin, HorizontalMargin float64

	BodySize, BodyPad float64
	LineSpace float64

	HeadSize, HeadPad float64
}

func NewLayout(fontPath string) (*layout, error) {
	f, err := ioutil.ReadFile(fontPath)
	if err != nil {
		return nil, err
	}

	l := new(layout)
	l.font, err = truetype.Parse(f)
	if err != nil {
		return nil, err
	}

	l.PageWidth = 612
	l.PageHeight = 792
	l.VerticalMargin = 72
	l.HorizontalMargin = 72
	l.BodySize = 12
	l.BodyPad = l.BodySize * 1.5
	l.LineSpace = l.BodySize * 1.25
	l.HeadSize = 13
	l.HeadPad = l.HeadSize * 3

	return l, nil
}

func (l *layout) FontName() string {
	return l.font.Name(truetype.NameIDPostscriptName)
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

func main() {
	l, err := NewLayout("/Users/smccoy/Library/Fonts/Roboto-Regular.ttf")
	if err != nil {
		panic(err)
	}

	err = l.Print(os.Stdout)
	if err != nil {
		panic(err)
	}
}

var layoutPreamble = `%!
% Letter = 8.5 x 11 in² = 612 x 792 pt²
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

/{{.FontName}} body_size selectfont

`
