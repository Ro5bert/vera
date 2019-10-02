package vera

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"io"
	"strings"
)

type CharSet struct {
	RowSep   string
	ColSep   string
	Center   string
	TopT     string
	BottomT  string
	LeftT    string
	RightT   string
	TLCorner string
	TRCorner string
	BLCorner string
	BRCorner string
}

var PrettyBoxCS = &CharSet{
	RowSep:   "─",
	ColSep:   "│",
	Center:   "┼",
	TopT:     "┬",
	BottomT:  "┴",
	LeftT:    "├",
	RightT:   "┤",
	TLCorner: "┌",
	TRCorner: "┐",
	BLCorner: "└",
	BRCorner: "┘",
}

var ASCIIBoxCS = &CharSet{
	RowSep:   "-",
	ColSep:   "|",
	Center:   "+",
	TopT:     "+",
	BottomT:  "+",
	LeftT:    "+",
	RightT:   "+",
	TLCorner: "+",
	TRCorner: "+",
	BLCorner: "+",
	BRCorner: "+",
}

// TODO: improve customizability
func RenderTT(stmt Stmt, truth Truth, out io.Writer, cs *CharSet, colorize bool) error {
	color.NoColor = !colorize
	if len(truth.Names) == 0 {
		return errors.New("cannot make a truth table with no atomics")
	}
	stmtStr := stmt.String()
	if err := printTopLine(len(truth.Names), len(stmtStr), out, cs); err != nil {
		return err
	}
	if err := printHeader(truth.Names, stmtStr, out, cs); err != nil {
		return err
	}
	if err := printHeaderLine(len(truth.Names), len(stmtStr), out, cs); err != nil {
		return err
	}
	n := 1 << len(truth.Names)
	for i := 0; i < n; i++ {
		if err := printData(truth.Val, len(truth.Names), stmt.Eval(truth), len(stmtStr), out, cs); err != nil {
			return err
		}
		truth.Val++
	}
	if err := printBottomLine(len(truth.Names), len(stmtStr), out, cs); err != nil {
		return err
	}
	return nil
}

func printTopLine(nAtomics int, outputWidth int, out io.Writer, cs *CharSet) error {
	return printLine(nAtomics, outputWidth, out, cs.RowSep, cs.TLCorner, cs.TopT, cs.TRCorner)
}

func printHeaderLine(nAtomics int, outputWidth int, out io.Writer, cs *CharSet) error {
	return printLine(nAtomics, outputWidth, out, cs.RowSep, cs.LeftT, cs.Center, cs.RightT)
}

func printBottomLine(nAtomics int, outputWidth int, out io.Writer, cs *CharSet) error {
	return printLine(nAtomics, outputWidth, out, cs.RowSep, cs.BLCorner, cs.BottomT, cs.BRCorner)
}

func calcInputWidth(nAtomics int) int {
	return nAtomics + 2*(nAtomics-1)
}

func printLine(nAtomics int, outputWidth int, out io.Writer, rowSep string, l string, m string, r string) error {
	_, err := fmt.Fprintf(out, "%s%s%s%s%s\n",
		l,
		strings.Repeat(rowSep, calcInputWidth(nAtomics)),
		m,
		strings.Repeat(rowSep, outputWidth),
		r,
	)
	return err
}

func printHeader(atomics []byte, stmt string, out io.Writer, cs *CharSet) error {
	var sb strings.Builder
	sb.Grow(calcInputWidth(len(atomics)))
	for i := len(atomics) - 1; i >= 0; i-- {
		sb.WriteByte(atomics[i])
		if i > 0 {
			sb.WriteString("  ")
		}
	}
	return printRow(sb.String(), stmt, out, cs)
}

func centerText(text string, width int) string {
	// Assumes text is ASCII.
	return fmt.Sprintf("%[1]*s", -width, fmt.Sprintf("%[1]*s", (width+len(text))/2, text))
}

func printData(truth uint64, nAtomics int, output bool, outputWidth int, out io.Writer, cs *CharSet) error {
	var sb strings.Builder
	for i := nAtomics - 1; i >= 0; i-- {
		if truth&(1<<i) > 0 {
			sb.WriteString(color.GreenString("1"))
		} else {
			sb.WriteString(color.RedString("0"))
		}
		if i > 0 {
			sb.WriteString("  ")
		}
	}
	var outputStr string
	if output {
		outputStr = color.GreenString(centerText("1", outputWidth))
	} else {
		outputStr = color.RedString(centerText("0", outputWidth))
	}
	return printRow(sb.String(), outputStr, out, cs)
}

func printRow(input string, output string, out io.Writer, cs *CharSet) error {
	var sb strings.Builder
	sb.Grow(3 + len(input) + len(output))
	sb.WriteString(cs.ColSep)
	sb.WriteString(input)
	sb.WriteString(cs.ColSep)
	sb.WriteString(output)
	sb.WriteString(cs.ColSep)
	_, err := fmt.Fprintln(out, sb.String())
	return err
}
