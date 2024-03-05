package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jewharton/node-redis-dump-repair/tokenizer"
	"github.com/spf13/cobra"
)

var replacements = map[byte]string{
	'\n': "\\n",
	'\r': "\\r",
	'\\': "\\\\",
	'"':  "\\\"",
	0x0:  "\\x00",
}

var cmd = &cobra.Command{
	Use:   "redis-dump-repair [input-file] [output-file]",
	Short: "Repair malformed dumps produced by the redis-dump NPM package.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		input, err := os.Open(args[0])
		if err != nil {
			fmt.Println("Error opening input file:", err)
			os.Exit(1)
		}
		defer func() { _ = input.Close() }()

		output, err := os.Create(args[1])
		if err != nil {
			fmt.Println("Error creating output file:", err)
			os.Exit(1)
		}
		defer func() { _ = output.Close() }()

		parser := tokenizer.New(bufio.NewReader(input))
		newline := true
		for done := false; !done; {
			token, err := parser.Next()
			if err != nil {
				fmt.Println("Error parsing input file:", err)
				os.Exit(1)
			}

			switch token.Kind() {
			case tokenizer.String:
				if newline {
					// The first string on a line is expected to be a Redis command.
					cmd := fmt.Sprintf("%-8s", token.Value())
					if _, err := output.Write([]byte(cmd)); err != nil {
						fmt.Println("Error writing to output file:", err)
						os.Exit(1)
					}
					newline = false
					continue
				}
				var sb strings.Builder
				sb.WriteString(" \"")
				for _, ch := range []byte(token.Value()) {
					if replacement, ok := replacements[ch]; ok {
						sb.WriteString(replacement)
						continue
					}
					sb.WriteByte(ch)
				}
				sb.WriteByte('"')
				if _, err := output.Write([]byte(sb.String())); err != nil {
					fmt.Println("Error writing to output file:", err)
					os.Exit(1)
				}
			case tokenizer.Newline:
				newline = true
				if _, err := output.Write([]byte(token.Value())); err != nil {
					fmt.Println("Error writing to output file:", err)
					os.Exit(1)
				}
			case tokenizer.EOF:
				done = true
			default:
				panic(fmt.Sprintf("unexpected token kind %d", token.Kind()))
			}
		}
	},
}

func main() {
	_ = cmd.Execute()
}
