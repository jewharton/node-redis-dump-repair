package tokenizer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Tokenizer reads tokens from a bufio.Reader.
type Tokenizer struct {
	reader         *byteReader
	err            error
	justReadString bool
}

// New returns a new Tokenizer.
func New(reader *bufio.Reader) *Tokenizer {
	return &Tokenizer{reader: &byteReader{reader: reader}}
}

// Next returns the next token, whether it was the last in its line,
// and the error that occurred when reading it (if any).
//
// A token can be a string, a newline character, or an end-of-file marker.
// String literals may be expressed as either of the following:
//   - A sequence of alphanumeric characters in addition to '_', ':', or '-'.
//   - A sequence of bytes surrounded by single quotation marks.
//     If a backslash or single quote is contained in the sequence between
//     the surrounding single quotes, it must be escaped with a backslash.
//
// Next should not be called after it has returned an error or an EOF token.
func (t *Tokenizer) Next() (token Token, err error) {
	if t.err != nil {
		return invalidToken, t.err
	}

	defer func() {
		if err != nil {
			t.err = err
		}
	}()

	for {
		ch, err := t.reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				t.err = err
				return eofToken, nil
			}
			return invalidToken, err
		}

		switch ch {
		case ' ':
			t.justReadString = false
			continue
		case '\n':
			t.justReadString = false
			return newlineToken, nil
		case '\'':
			if t.justReadString {
				return invalidToken, fmt.Errorf("expected space or newline before string at position %d", t.reader.pos-1)
			}
			t.justReadString = true

			var sb strings.Builder
			var escaping bool
			for {
				ch, err := t.reader.Read()
				if err != nil {
					if errors.Is(err, io.EOF) {
						return invalidToken, io.ErrUnexpectedEOF
					}
					return invalidToken, err
				}
				if escaping {
					if ch == '\\' || ch == '\'' {
						sb.WriteByte(ch)
						escaping = false
						continue
					}
					return invalidToken, fmt.Errorf("unescaped backslash or invalid escape sequence at position %d", t.reader.pos-2)
				}
				if ch == '\\' {
					escaping = true
					continue
				}
				if ch == '\'' {
					break
				}
				sb.WriteByte(ch)
			}
			t.justReadString = true
			return Token{value: sb.String(), kind: String}, nil
		}

		if isUnquotedStringCharacter(ch) {
			if t.justReadString {
				return invalidToken, fmt.Errorf("expected space or newline before string at position %d", t.reader.pos)
			}
			t.justReadString = true

			var sb strings.Builder
			sb.WriteByte(ch)
			for {
				ch, err := t.reader.Read()
				if err != nil {
					if errors.Is(err, io.EOF) {
						return Token{value: sb.String(), kind: String}, nil
					}
					return invalidToken, err
				}
				if isUnquotedStringCharacter(ch) {
					sb.WriteByte(ch)
					continue
				}
				if ch == ' ' || ch == '\n' {
					if err := t.reader.Unread(); err != nil {
						return invalidToken, err
					}
					return Token{value: sb.String(), kind: String}, nil
				}
				return invalidToken, fmt.Errorf("unexpected character 0x%x at position %d", ch, t.reader.pos-1)
			}
		}
		return invalidToken, fmt.Errorf("unexpected character 0x%x at position %d", ch, t.reader.pos-1)
	}
}

// byteReader tracks the amount of bytes read.
type byteReader struct {
	pos    int
	reader *bufio.Reader
}

// Read reads and returns a single byte. If no byte is available, it returns an error.
func (b *byteReader) Read() (byte, error) {
	v, err := b.reader.ReadByte()
	if err == nil {
		b.pos++
	}
	return v, err
}

// Unread unreads the last byte. Only the most recently read byte can be unread.
func (b *byteReader) Unread() error {
	err := b.reader.UnreadByte()
	if err == nil {
		b.pos--
	}
	return err
}

// isUnquotedStringCharacter returns whether the given character is valid for unquoted string literals.
func isUnquotedStringCharacter(ch byte) bool {
	return ('a' <= ch && ch <= 'z') ||
		('A' <= ch && ch <= 'Z') ||
		('0' <= ch && ch <= '9') ||
		ch == '_' || ch == ':' || ch == '-'
}
