package main

import (
	"github.com/Ro5bert/vera"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use: "vera",
}

var ttCmd = &cobra.Command{
	Use: "tt",
	Short: "Generate a truth table for the given logical expression",
	RunE: tt,
	Args: cobra.ExactArgs(1),
}

func init() {
	ttCmd.Flags().Bool("no-color", false, "do not colorize the output")
	ttCmd.Flags().Bool("ascii", false, "use ASCII characters to draw the table")
	rootCmd.AddCommand(ttCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		// _, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func tt(cmd *cobra.Command, args []string) error {
	nocolor, err := cmd.Flags().GetBool("no-color")
	if err != nil {
		panic(err)
	}
	ascii, err := cmd.Flags().GetBool("ascii")
	if err != nil {
		panic(err)
	}
	stmt, truth, err := vera.Parse(args[0])
	if err != nil {
		return err
	}
	var cs *vera.CharSet
	if ascii {
		cs = vera.ASCIIBoxCS
	} else {
		cs = vera.PrettyBoxCS
	}
	return vera.RenderTT(stmt, truth, os.Stdout, cs, !nocolor)
}
