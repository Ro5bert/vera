package vera

import (
	"testing"
)

func TestParse(t *testing.T) {
	type testCase struct {
		input    string
		expected []bool
	}
	for _, c := range []testCase{
		{"a", []bool{false, true}},
		{"(a)", []bool{false, true}},
		{"a&b", []bool{false, false, false, true}},
		{"a|b", []bool{false, true, true, true}},
		{"a^b", []bool{false, true, true, false}},
		{"a>b", []bool{true, false, true, true}},
		{"a=b", []bool{true, false, false, true}},
		{"0", []bool{false}},
		{"1", []bool{true}},
		{"a|!0", []bool{true, true}},
		{"(a & b)", []bool{false, false, false, true}},
		{"!(a > b)", []bool{false, true, false, false}},
		{"!!(a = b)", []bool{true, false, false, true}},
		{"(a = b) | b", []bool{true, false, true, true}},
	} {
		stmt, truth, err := parse(c.input)
		if err != nil {
			t.Fatalf("error occurred while parsing: %v (input: %s)", err, c.input)
		}
		for _, exp := range c.expected {
			if stmt.eval(truth) != exp {
				t.Fatalf("expected %t for evalation of '%s' at %s", exp, c.input, truth)
			}
			truth.val++
		}
	}
}