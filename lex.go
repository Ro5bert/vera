package vera

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

const (
	negateSym = '!'
	andSym    = '&'
	orSym     = '|'
	xorSym    = '^'
	condSym   = '>'
	bicondSym = '='
)

type lexemeType byte

const (
	LTFalse lexemeType = iota
	LTTrue
	LTNegate
	LTOperator
	LTOpenParen
	LTCloseParen
	LTStatement
)

type lexeme struct {
	t lexemeType
	v byte
}

type lexerResult struct {
	l   lexeme
	err error
}

type statefn func(byte, *lexer) (lexemeType, statefn, error)

type lexer struct {
	input    string
	c        chan lexerResult
	nextIdx  int
	nestCnt  int
	allowEOF bool
}

func removeAllWS(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func lex(input string) chan lexerResult {
	l := &lexer{
		input: removeAllWS(input),
		// Arbitrary buffer size.
		c:        make(chan lexerResult, 10),
		allowEOF: false,
	}
	go l.run()
	return l.c
}

func (l *lexer) run() {
	for sfn := lexStatement; sfn != nil; {
		n, eof := l.next()
		if eof {
			if !l.allowEOF {
				l.c <- lexerResult{err: errors.New("unexpected EOF")}
			}
			break
		}
		var lt lexemeType
		var err error
		lt, sfn, err = sfn(n, l)
		if err != nil {
			l.c <- lexerResult{err: err}
			break
		}
		l.c <- lexerResult{l: lexeme{lt, n}}
	}
	// Closing the channel without any errors implies EOF.
	close(l.c)
}

func (l *lexer) next() (byte, bool) {
	if l.nextIdx == len(l.input) {
		return 0, true
	}
	// Indexing into string => we are not expecting UTF-8 chars with width > 1 byte.
	next := l.input[l.nextIdx]
	l.nextIdx++
	return next, false
}

func (l *lexer) nest() {
	l.nestCnt++
	l.allowEOF = false
}

func (l *lexer) denest() bool {
	if l.nestCnt == 0 {
		// false return indicates failure.
		return false
	}
	l.nestCnt--
	l.allowEOF = l.nestCnt == 0
	return true
}

func lexStatement(n byte, l *lexer) (lexemeType, statefn, error) {
	// By default, allow EOF if there are no unmatched parentheses.
	// Some branches in the below switch set the allowEOF flag based on other conditions.
	l.allowEOF = l.nestCnt == 0
	switch n {
	case negateSym:
		l.allowEOF = false
		return LTNegate, lexStatement, nil
	case '(':
		l.nest()
		return LTOpenParen, lexStatement, nil
	case '0':
		return LTFalse, lexOperator, nil
	case '1':
		return LTTrue, lexOperator, nil
	}
	if ('a' <= n && n <= 'z') || ('A' <= n && n <= 'Z') {
		return LTStatement, lexOperator, nil
	}
	return 0, nil, fmt.Errorf("unexpected char '%c'; expected '%c', '(', '0', '1', or a statement", n, negateSym)
}

func lexOperator(n byte, l *lexer) (lexemeType, statefn, error) {
	switch n {
	case ')':
		if !l.denest() {
			return 0, nil, errors.New("unexpected closing parenthesis: no corresponding opening parenthesis")
		}
		return LTCloseParen, lexOperator, nil
	case andSym, orSym, xorSym, condSym, bicondSym:
		l.allowEOF = false
		return LTOperator, lexStatement, nil
	}
	return 0, nil, fmt.Errorf("unexpected char '%c'; expected ')', '%c', '%c', '%c', '%c', or '%c'",
		n, andSym, orSym, xorSym, condSym, bicondSym)
}
