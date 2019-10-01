package vera

import "testing"

func TestLex(t *testing.T) {
	type testCase struct {
		input    string
		expected []lexeme
	}
	for _, c := range []testCase{
		{"a", []lexeme{{LTStatement, 'a'}}},
		{"(a)", []lexeme{
			{LTOpenParen, '('},
			{LTStatement, 'a'},
			{LTCloseParen, ')'},
		}},
		{"a > b", []lexeme{
			{LTStatement, 'a'},
			{LTOperator, '>'},
			{LTStatement, 'b'},
		}},
		{"  (  a       >b)   &   1 ", []lexeme{
			{LTOpenParen, '('},
			{LTStatement, 'a'},
			{LTOperator, '>'},
			{LTStatement, 'b'},
			{LTCloseParen, ')'},
			{LTOperator, '&'},
			{LTTrue, '1'},
		}},
		{"!(!(a = b) | !0) > (c ^ d)", []lexeme{
			{LTNegate, '!'},
			{LTOpenParen, '('},
			{LTNegate, '!'},
			{LTOpenParen, '('},
			{LTStatement, 'a'},
			{LTOperator, '='},
			{LTStatement, 'b'},
			{LTCloseParen, ')'},
			{LTOperator, '|'},
			{LTNegate, '!'},
			{LTFalse, '0'},
			{LTCloseParen, ')'},
			{LTOperator, '>'},
			{LTOpenParen, '('},
			{LTStatement, 'c'},
			{LTOperator, '^'},
			{LTStatement, 'd'},
			{LTCloseParen, ')'},
		}},
	} {
		idx := 0
		for r := range lex(c.input) {
			switch {
			case r.err != nil:
				t.Fatalf("error occurred while lexing: %v (input: %s)", r.err, c.input)
			case idx == len(c.expected):
				t.Fatalf("too many lexemes (input: %s)", c.input)
			case r.l != c.expected[idx]:
				t.Fatalf("expected %v; got %v (input: %s)", c.expected[idx], r.l, c.input)
			}
			idx++
		}
	}
}

func TestLexError(t *testing.T) {
	type testCase struct {
		input string
		error bool
	}
	for _, c := range []testCase{
		{"", true},
		{"()", true},
		{">", true},
		{"(=)", true},
		{"(", true},
		{")", true},
		{"a >", true},
		{"^ a", true},
		{"(a & b > c)", false},
	} {
		err := false
		for r := range lex(c.input) {
			if r.err != nil {
				err = true
			}
		}
		if err != c.error {
			fill := ""
			if !c.error {
				fill = "not "
			}
			t.Fatalf("expected '%s' to %serror", c.input, fill)
		}
	}
}
