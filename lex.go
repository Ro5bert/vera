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

func (lt lexemeType) String() string {
	switch lt {
	case ltFalse:
		return "False"
	case ltTrue:
		return "True"
	case ltNegate:
		return "Negate"
	case ltOperator:
		return "Operator"
	case ltOpenParen:
		return "OpenParen"
	case ltCloseParen:
		return "CloseParen"
	case ltStatement:
		return "Statement"
	default:
		panic("lexemeType not added to String method!")
	}
}

const (
	ltFalse lexemeType = iota
	ltTrue
	ltNegate
	ltOperator
	ltOpenParen
	ltCloseParen
	ltStatement
)

type lexeme struct {
	t lexemeType
	v byte
}

type lexerResult struct {
	l   lexeme
	err error
}

// statefn is a state combined with an associated action. See Rob Pike's talk on lexical scanning.
type statefn func(byte, *lexer) (lexemeType, statefn, error)

type lexer struct {
	input    string
	c        chan lexerResult
	nextIdx  int
	nestCnt  int
	allowEOF bool
}

// removeAllWS removes all the whitespace from a string and returns the new string, where whitespace is identified
// according to unicode.IsSpace.
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

// lex lexes the given string in a separate goroutine and outputs the resultant lexerResults over the returned channel.
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

// run is the main loop for a lexer. It should be called in a separate goroutine.
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

// next returns the next byte in the input string. The boolean return value indicates if the end of the string was
// reached (i.e. EOF); if it is true, the byte return value should be disregarded.
func (l *lexer) next() (byte, bool) {
	if l.nextIdx == len(l.input) {
		return 0, true
	}
	// Indexing into string => we are not expecting UTF-8 chars with width > 1 byte.
	next := l.input[l.nextIdx]
	l.nextIdx++
	return next, false
}

// nest increments nestCnt and sets allowEOF as appropriate.
func (l *lexer) nest() {
	l.nestCnt++
	l.allowEOF = false
}

// denest decrements nestCnt and sets allowEOF as appropriate. The boolean return value indicates success; false is
// returned if nestCnt was already zero when denest was called (i.e. parentheses not matched).
func (l *lexer) denest() bool {
	if l.nestCnt == 0 {
		// false return indicates failure.
		return false
	}
	l.nestCnt--
	l.allowEOF = l.nestCnt == 0
	return true
}

// lexStatement is a statefn for parsing the start of a statement (this includes opening parentheses, "0", "1", and
// letters) or negation.
func lexStatement(n byte, l *lexer) (lexemeType, statefn, error) {
	// By default, allow EOF if there are no unmatched parentheses.
	// Some branches in the below switch set the allowEOF flag based on other conditions.
	l.allowEOF = l.nestCnt == 0
	switch n {
	case negateSym:
		l.allowEOF = false
		return ltNegate, lexStatement, nil
	case '(':
		l.nest()
		return ltOpenParen, lexStatement, nil
	case '0':
		return ltFalse, lexOperator, nil
	case '1':
		return ltTrue, lexOperator, nil
	}
	if ('a' <= n && n <= 'z') || ('A' <= n && n <= 'Z') {
		return ltStatement, lexOperator, nil
	}
	return 0, nil, fmt.Errorf("unexpected char '%c'; expected '%c', '(', '0', '1', or a statement", n, negateSym)
}

// lexOperator is a statefn for parsing a binary operator or a closing parenthesis.
func lexOperator(n byte, l *lexer) (lexemeType, statefn, error) {
	switch n {
	case ')':
		if !l.denest() {
			return 0, nil, errors.New("unexpected closing parenthesis: no corresponding opening parenthesis")
		}
		return ltCloseParen, lexOperator, nil
	case andSym, orSym, xorSym, condSym, bicondSym:
		l.allowEOF = false
		return ltOperator, lexStatement, nil
	}
	return 0, nil, fmt.Errorf("unexpected char '%c'; expected ')', '%c', '%c', '%c', '%c', or '%c'",
		n, andSym, orSym, xorSym, condSym, bicondSym)
}
