package vera

import "testing"

func TestLex(t *testing.T) {
	type testCase struct {
		input    string
		expected []lexeme
	}
	for _, c := range []testCase{
		{"", []lexeme{}},
		{"a", []lexeme{{Statement, 'a'}}},
		{"a > b", []lexeme{
			{Statement, 'a'},
			{Operator, '>'},
			{Statement, 'b'},
		}},
		{"  (  a       >b)   &   1 ", []lexeme{
			{OpenParen, '('},
			{Statement, 'a'},
			{Operator, '>'},
			{Statement, 'b'},
			{CloseParen, ')'},
			{Operator, '&'},
			{True, '1'},
		}},
		{"!(!(a = b) | !0) > (c ^ d)", []lexeme{
			{Negate, '!'},
			{OpenParen, '('},
			{Negate, '!'},
			{OpenParen, '('},
			{Statement, 'a'},
			{Operator, '='},
			{Statement, 'b'},
			{CloseParen, ')'},
			{Operator, '|'},
			{Negate, '!'},
			{False, '0'},
			{CloseParen, ')'},
			{Operator, '>'},
			{OpenParen, '('},
			{Statement, 'c'},
			{Operator, '^'},
			{Statement, 'd'},
			{CloseParen, ')'},
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
