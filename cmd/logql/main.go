package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/spf13/cobra"

	"go.linka.cloud/logql/pkg/logql"
)

const (
	red    = 31
	yellow = 33
	blue   = 36
	white  = 39
	gray   = 90
)

var internalLabels = labels.Labels{{Name: "__internal__"}}

var (
	skipLabels bool
	cmd        = &cobra.Command{
		Use:   "logql (file) <logql query>",
		Short: "A simple LogQL query processor that reads log lines from stdin and outputs processed log lines to stdout.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			q := args[0]
			if len(args) == 2 {
				q = args[1]
			}
			if i := strings.IndexRune(q, '}'); q[0] == '{' && i > 0 {
				q = q[i+1:]
			}
			if q[0] != '{' {
				q = internalLabels.String() + q
			}
			e, err := logql.ParseLogSelector(q)
			if err != nil {
				return err
			}
			p, err := e.Pipeline()
			if err != nil {
				return err
			}
			in := os.Stdin
			if len(args) == 2 {
				f, err := os.Open(args[0])
				if err != nil {
					return err
				}
				defer f.Close()
				in = f
			}
			s := bufio.NewScanner(in)
			for s.Scan() {
				line, l, ok := p.ForStream(internalLabels).Process(s.Bytes())
				if !ok || len(line) == 0 {
					continue
				}
				if !skipLabels {
					fmt.Printf("%s %s\n", formatLabels(l.Labels().WithoutLabels("__internal__")), string(line))
				} else {
					fmt.Println(string(line))
				}
			}
			return nil
		},
	}
)

func formatLabels(ls labels.Labels) string {
	var b bytes.Buffer

	b.WriteString(color.New(red).Sprint("{"))
	for i, l := range ls {
		if i > 0 {
			b.WriteByte(',')
			b.WriteByte(' ')
		}
		b.WriteString(color.New(gray).Sprint(l.Name))
		b.WriteByte('=')
		b.WriteString(color.New(blue).Sprint(strconv.Quote(l.Value)))
	}
	b.WriteString(color.New(red).Sprint("}"))
	return b.String()
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cmd.Flags().BoolVar(&skipLabels, "skip-labels", false, "Skip printing labels in output")
}
