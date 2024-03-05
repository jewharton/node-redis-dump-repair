package tokenizer_test

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/jewharton/node-redis-dump-repair/tokenizer"
	"github.com/stretchr/testify/require"
)

func TestTokenizer(t *testing.T) {
	parser := tokenizer.New(bufio.NewReader(strings.NewReader("")))
	token, err := parser.Next()
	require.NoError(t, err)
	require.Equal(t, tokenizer.EOF, token.Kind())
	require.Panics(t, func() { token.Value() })

	type expectedToken struct {
		kind    tokenizer.Kind
		value   string
		err     error
		errText string
	}

	newlineToken := expectedToken{kind: tokenizer.Newline, value: "\n"}
	eofToken := expectedToken{kind: tokenizer.EOF}
	eofErrorToken := expectedToken{kind: tokenizer.Invalid, err: io.EOF}

	for _, tt := range []struct {
		name           string
		input          string
		expectedTokens []expectedToken
	}{
		{
			name: "Empty string",
			expectedTokens: []expectedToken{
				eofToken,
				eofErrorToken,
			},
		}, {
			name:  "Plain string",
			input: "foo_:-",
			expectedTokens: []expectedToken{
				{kind: tokenizer.String, value: "foo"},
				eofToken,
				eofErrorToken,
			},
		}, {
			name:  "Newline-separated strings",
			input: "foo\nbar\nbaz\n",
			expectedTokens: []expectedToken{
				{kind: tokenizer.String, value: "foo"},
				newlineToken,
				{kind: tokenizer.String, value: "bar"},
				newlineToken,
				{kind: tokenizer.String, value: "baz"},
				newlineToken,
				eofToken,
				eofErrorToken,
			},
		}, {
			name:  "Quoted string with special and escaped characters",
			input: "'foo' 'b\na\\\\r\\''\n",
			expectedTokens: []expectedToken{
				{kind: tokenizer.String, value: "foo"},
				{kind: tokenizer.String, value: "b\na\\r\\'"},
				newlineToken,
				eofToken,
				eofErrorToken,
			},
		}, {
			name:  "Unterminated quoted string",
			input: "foo 'bar",
			expectedTokens: []expectedToken{
				{kind: tokenizer.String, value: "foo"},
				{kind: tokenizer.Invalid, err: io.ErrUnexpectedEOF},
			},
		}, {
			name:  "Invalid character in unquoted string",
			input: "foo b@r",
			expectedTokens: []expectedToken{
				{kind: tokenizer.String, value: "foo"},
				{kind: tokenizer.Invalid, errText: "unexpected character 0x40 at position 5"},
			},
		}, {
			name:  "Unescaped slash in quoted string",
			input: "foo 'b\\ar'",
			expectedTokens: []expectedToken{
				{kind: tokenizer.String, value: "foo"},
				{kind: tokenizer.Invalid, errText: "unescaped backslash or invalid escape sequence at position 6"},
			},
		}, {
			name:  "Missing space between strings",
			input: "'foo'bar",
			expectedTokens: []expectedToken{
				{kind: tokenizer.String, value: "foo"},
				{kind: tokenizer.Invalid, errText: "expected space or newline before string at position 6"},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			parser := tokenizer.New(bufio.NewReader(strings.NewReader(tt.input)))
			for _, expected := range tt.expectedTokens {
				token, err := parser.Next()
				if expected.err != nil {
					require.ErrorIs(t, err, expected.err)
				} else if expected.errText != "" {
					require.EqualError(t, err, expected.errText)
				} else {
					require.NoError(t, err)
				}
				require.Equal(t, expected.kind, token.Kind())
				if expected.value == "" {
					require.Panics(t, func() { token.Value() })
				}
			}
		})
	}
}
