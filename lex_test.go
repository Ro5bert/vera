package vera

import "testing"

func TestLex(t *testing.T) {
	type testCase struct {
		input    string
		expected []lexeme
	}
	for _, c := range []testCase{
		{"", []lexeme{}},
		{"a", []lexeme{{LTStatement, 'a'}}},
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
				t.Fatalf("error occurred while lexing: %v", r.err)
			case idx == len(c.expected):
				t.Fatal("too many lexemes")
			case r.l != c.expected[idx]:
				t.Fatalf("expected %v; got %v", c.expected[idx], r.l)
			}
			idx++
		}
	}
}
