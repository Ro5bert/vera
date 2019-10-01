package vera

import (
	"fmt"
	"strings"
)

func alphaToIdx(ascii byte) byte {
	if ascii >= 'a' {
		ascii -= 6
	}
	return ascii - 65
}

func idxToAlpha(idx byte) byte {
	if idx >= 26 {
		idx += 6
	}
	return idx + 65
}

type truth struct {
	val      uint64
	shiftMap *[52]byte
	names    []byte
}

func (t truth) get(stmt byte) bool {
	return t.val&(1<<t.shiftMap[alphaToIdx(stmt)]) > 0
}

func (t truth) String() string {
	var sb strings.Builder
	sb.WriteByte('{')
	for i, b := range t.names {
		sb.WriteByte(b)
		sb.WriteByte(':')
		if t.get(b) {
			sb.WriteByte('1')
		} else {
			sb.WriteByte('0')
		}
		if i < len(t.names)-1 {
			sb.WriteByte(',')
		}
	}
	sb.WriteByte('}')
	return sb.String()
}

func newTruth(atomics uint64) truth {
	var shiftMap [52]byte
	names := make([]byte, 0, 52)
	var shift byte
	for i := byte(0); i < 52; i++ {
		if atomics&(1<<i) > 0 {
			shiftMap[i] = shift
			shift++
			// Use append even though we pre-allocated so that the slice length represents the number of atomics.
			names = append(names, idxToAlpha(i))
		}
	}
	return truth{0, &shiftMap, names}
}

type operator func(bool, bool) bool

type stmt interface {
	fmt.Stringer
	eval(truth) bool
}

func surroundIfBinary(s stmt) string {
	if _, ok := s.(binaryStmt); ok {
		return "(" + s.String() + ")"
	}
	return s.String()
}

type falseStmt struct{}

func (falseStmt) eval(truth) bool {
	return false
}

func (falseStmt) String() string {
	return "0"
}

type trueStmt struct{}

func (trueStmt) eval(truth) bool {
	return true
}

func (trueStmt) String() string {
	return "1"
}

type negatedStmt struct {
	stmt
}

func (s negatedStmt) eval(t truth) bool {
	return !s.stmt.eval(t)
}

func (s negatedStmt) String() string {
	return "!" + surroundIfBinary(s.stmt)
}

type atomicStmt byte

func (s atomicStmt) eval(t truth) bool {
	return t.get(byte(s))
}

func (s atomicStmt) String() string {
	return string(s)
}

type binaryStmt struct {
	left  stmt
	op    operator
	right stmt
	opSym string
}

func (s binaryStmt) eval(t truth) bool {
	return s.op(s.left.eval(t), s.right.eval(t))
}

func (s binaryStmt) String() string {
	return surroundIfBinary(s.left) + s.opSym + surroundIfBinary(s.right)
}

func and(left bool, right bool) bool {
	return left && right
}

func or(left bool, right bool) bool {
	return left || right
}

func xor(left bool, right bool) bool {
	return (left && !right) || (!left && right)
}

func cond(left bool, right bool) bool {
	return !left || right
}

func bicond(left bool, right bool) bool {
	return left == right
}

func byteToOp(b byte) operator {
	switch b {
	case andSym:
		return and
	case orSym:
		return or
	case xorSym:
		return xor
	case condSym:
		return cond
	case bicondSym:
		return bicond
	default:
		// Lexer should guarantee this never happens.
		panic(fmt.Sprintf("invalid op byte '%c'", b))
	}
}

func parse(input string) (stmt, truth, error) {
	stmt, atomics, err := parseRecursive(lex(input))
	return stmt, newTruth(atomics), err
}

type stmtBuilder struct {
	inner   stmt
	negated bool
}

func (sb *stmtBuilder) negate() {
	sb.negated = !sb.negated
}

func (sb *stmtBuilder) build() stmt {
	if sb.negated {
		return negatedStmt{sb.inner}
	}
	return sb.inner
}

func parseRecursive(c chan lexerResult) (stmt, uint64, error) {
	const (
		expStmt = iota
		expOpOrClose
		expClose
	)
	state := expStmt
	left := &stmtBuilder{}
	right := &stmtBuilder{}
	// op contains the binary operator if this invocation of parseRecursive is parsing a binary statement. If this
	// invocation of parseRecursive is parsing a single statement, op is nil. For example, for the input 'a & (b)', the
	// op in the outer invocation of parseRecursive will be set to the AND operator, and the op in the inner invocation
	// of parseRecursive (i.e. when parsing '(b)') will be nil.
	var op operator
	var opSym string
	// atomics is a bit field where a set bit indicates that the corresponding atomic statement (i.e. an ascii letter)
	// appeared in the input. Whether 'A' occurred in the input is 'atomics & 1', and whether 'z' occurred in the input
	// is 'atomics & 1 << 51', for example.
	var atomics uint64
	// pick is to prevent cluttering below with nil checks on op.
	pick := func() *stmtBuilder {
		// "pick" the left or right statement.
		if op == nil {
			return left
		}
		return right
	}
forLoop:
	for lr := range c {
		if lr.err != nil {
			return nil, 0, lr.err
		}
		switch state {
		case expStmt:
			switch lr.l.t {
			case LTFalse:
				pick().inner = falseStmt{}
			case LTTrue:
				pick().inner = trueStmt{}
			case LTNegate:
				// TODO: try to preserve original statement as faithfully as possible: increment negate counter instead?
				pick().negate()
				// continue so state is not set below the switch statement.
				continue
			case LTOpenParen:
				var err error
				var a uint64
				pick().inner, a, err = parseRecursive(c)
				if err != nil {
					return nil, 0, err
				}
				atomics |= a
			case LTStatement:
				atomics |= 1 << alphaToIdx(lr.l.v)
				pick().inner = atomicStmt(lr.l.v)
			default:
				// Lexer should guarantee this never happens.
				panic(fmt.Sprintf("expected False, True, Negate, OpenParen, or Statement, not %s", lr.l.t))
			}
			if op == nil {
				state = expOpOrClose
			} else {
				state = expClose
			}
		case expOpOrClose:
			switch lr.l.t {
			case LTOperator:
				op = byteToOp(lr.l.v)
				opSym = " " + string(lr.l.v) + " "
				state = expStmt
			case LTCloseParen:
				break forLoop
			default:
				// Lexer should guarantee this never happens.
				panic(fmt.Sprintf("expected Operator or CloseParen/EOF, not %s", lr.l.t))
			}
		case expClose:
			if lr.l.t != LTCloseParen {
				// This should only ever happen if the lexeme is of type LTOperator, since the lexer does not understand
				// that multiple operators chained together without parentheses is ambiguous.
				return nil, 0, fmt.Errorf("expected CloseParen/EOF, not %s", lr.l.t)
			}
			break forLoop
		}
	}
	if op == nil {
		return left.build(), atomics, nil
	}
	return binaryStmt{left.build(), op, right.build(), opSym}, atomics, nil
}
