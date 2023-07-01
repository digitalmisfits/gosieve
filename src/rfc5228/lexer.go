/*
 * MIT License
 *
 * Copyright (c) 2023 Erik-Paul Dittmer (epdittmer@s114.nl)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NON INFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR
 * ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
 * TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package rfc5228

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// item represents a token or input string returned from the scanner.
type item struct {
	typ itemType // The type of this item.
	pos Pos      // The starting position, in bytes, of this item in the input string.
	val string   // The value of this item.
}

func (i item) String() string {
	return fmt.Sprintf("type = [%d], pos = [%d], value = [%s]", i.typ, i.pos, i.val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError itemType = iota // error occurred; value is input of error
	itemEOF
	itemComment
	itemIdentifier
	itemEnd
	itemString
	itemNumeric
	itemStringListOpen
	itemStringListClose
	itemTestListOpen
	itemTestListClose
	itemBlockOpen
	itemBlockClose
)

const textMarker = "input:"

const EOF = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name  string // name of the lexer; used for error reporting
	input string // the string being scanned
	start Pos    // start position of this token
	pos   Pos    // current position in the input
	atEOF bool   // we have hit the end of input and returned EOF
	width int    // width of the last rune read
	item  item   // item to return to parser
}

// thisItem returns the item at the current input point with the specified type
// and advances the input.
func (l *lexer) thisItem(t itemType) item {
	i := item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
	return i
}

func (l *lexer) emitItem(i item) stateFn {
	l.item = i
	return nil
}

// emit passes the trailing input as an item back to the parser.
func (l *lexer) emit(t itemType) stateFn {
	return l.emitItem(l.thisItem(t))
}

// next advances the position past the decoded rune
func (l *lexer) next() rune {

	// if we read past the end of the input we've reached the end of the file
	if l.pos >= Pos(len(l.input)) {
		l.width = 0
		return EOF
	}

	// decode the string into a rune (utf-8 code points) and advance the position
	r, size := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = size
	l.pos += Pos(size)
	return r
}

func (l *lexer) acceptExact(r rune) bool {
	if next := l.next(); next == r {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptAny(runes []rune) bool {
	r := l.next()
	for _, v := range runes {
		if r == v {
			return true
		}
	}
	l.backup()
	return false
}

func (l *lexer) acceptRunStringSequence(s string) bool {
	return l.acceptRunSequence([]rune(s))
}

// acceptRunStringSequence consumes the next n items from the input
func (l *lexer) acceptRunSequence(runes []rune) bool {
	if l.isExactPrefix(runes) {
		for i := 0; i < len(runes); i++ {
			l.next()
		}
		return true
	}
	return false
}

// acceptRunStringSequence acceptRunStringSequence a run of runes from the valid set
func (l *lexer) acceptRunAny(valid string) {
	for strings.ContainsRune(valid, l.next()) {
		// consumed
	}
	l.backup()
}

// backup steps back one rune.
func (l *lexer) backup() stateFn {
	if !l.atEOF && l.pos > 0 {
		_, w := utf8.DecodeLastRuneInString(l.input[:l.pos])
		l.pos -= Pos(w)
	}
	return nil
}

// ignore skips over the pending input before this point
func (l *lexer) ignore() {
	l.start = l.pos
}

// peek does return but does not accept a rune from the input
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// isExactPrefix tests if a run of runes equals a given prefix; this method does not accept any tokens (peek only)
func (l *lexer) isExactPrefix(prefix []rune) bool {
	offset := int(l.pos)
	width := 0
	for _, p := range prefix {
		r, size := utf8.DecodeRuneInString(l.input[offset+width:])
		width += size
		if r != p {
			return false
		}
	}
	return true
}

// isNotExactPrefix is the inverse of isExactPrefix
func (l *lexer) isNotExactPrefix(prefix []rune) bool {
	return !l.isExactPrefix(prefix)
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	l.item = item{itemEOF, l.pos, "EOF"}

	state := lexStart
	for {
		state = state(l)
		if state == nil {
			return l.item
		}
	}
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.next.
func (l *lexer) errorf(format string, args ...any) stateFn {
	l.item = item{itemError, l.start, fmt.Sprintf(format, args...)}
	l.start = 0
	l.pos = 0
	l.input = l.input[:0]
	return nil
}

// isWhitespace tests if a rune is a (part of a) whitespace character
//
// Whitespace is used to separate items.  Whitespace is made up of
// tabs, newlines (CRLF, never just '\r' or '\n'), and the space character.
//
// Comments are semantically equivalent to whitespace and can be used anyplace that whitespace is
// (with one exception in multi-line strings, as described in the grammar).
func isWhitespace(r rune) bool {
	return r == ' ' ||
		r == '\t' ||
		r == '\r' ||
		r == '\n' ||
		r == '/' ||
		r == '#'
}

func isOctetFiltered(r rune, filters ...rune) bool {
	if r < 0x01 || r > 0xFF {
		return false
	}
	for _, f := range filters {
		if f == r {
			return false
		}
	}
	return true
}

// isAlpha reports whether r is an alphabetic, digit, or underscore.
func isAlpha(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

func isAlphaNumeric(r rune) bool {
	return isAlpha(r) || isDigit(r)
}

func lex(name, input string) *lexer {
	return &lexer{
		name:  name,
		input: input,
		start: 0,
		pos:   0,
		width: 0,
	}
}

func lexStart(l *lexer) stateFn {
	for {
		switch r := l.peek(); {
		case r == EOF:
			return nil
		case isWhitespace(r):
			return lexWhitespace
		case isAlpha(r) && l.isNotExactPrefix([]rune(textMarker)):
			return lexIdentifier
		case r == '"':
			return lexQuotedString
		case r == 't' && l.isExactPrefix([]rune(textMarker)):
			return lexMultiline
		case r == '[':
			return lexStringList
		case r == ',':
			l.next()   // we only peeked `r`, so we need to absorb it
			l.ignore() // and we ignore it because we don't want it
		case r == ']':
			return lexStringList
		case r == ':':
			return lexTag
		case isDigit(r):
			return lexNumeric
		case r == '(':
			return lexTestList
		case r == ')':
			return lexTestList
		case r == ';':
			l.next() // we only peeked `r`, so we need to absorb it
			return l.emit(itemEnd)
		case r == '{':
			return lexBlock
		case r == '}':
			return lexBlock
		default:
			return l.errorf("unexpected rune")
		}
	}
}

// lexWhitespace scans whitespace.
func lexWhitespace(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == EOF:
			return nil
		case r == ' ' || r == '\t':
			l.ignore()
		case r == '\r':
			if next := l.next(); next != '\n' {
				return l.errorf("unexpected carriage return")
			}
			l.ignore()
		case r == '\n':
			return l.errorf("dangling line feed")
		case r == '#':
			return lexHashComment
		case r == '/':
			if next := l.next(); next != '*' {
				return l.errorf("unexpected bracket comment")
			}
			return lexBracketComment
		default:
			l.backup() // restore non matching rune
			return lexStart
		}
	}
}

// lexBracketComment scans a bracket-comment.
func lexBracketComment(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == EOF:
			return nil
		case isOctetFiltered(r, '\r', '\n', '*'):
			// absorb.
		case r == '\r':
			if next := l.next(); next != '\n' {
				return l.errorf("unexpected carriage return")
			}
		case r == '*':
			if next := l.next(); next != '/' {
				l.backup() // restore rune
			} else {
				return l.emit(itemComment)
			}
		default:
			return l.errorf("unexpected rune")
		}
	}
}

// lexHashComment scans a hash comment
func lexHashComment(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == EOF:
			return nil
		case isOctetFiltered(r, '\r', '\n'):
			// absorb.
		case r == '\r':
			// we don't want to include the trailing CRLF in the token value
			// as we already consume '\r' and we don't know if the next character is going to be '\n',
			// we need to peek '\n', if the next char is indeed '\n', we can backup the token stream
			// and let the whitespace state absorb the CRLF
			if l.isExactPrefix([]rune{'\n'}) {
				l.backup()
				return l.emit(itemComment)
			} else {
				return l.errorf("unexpected carriage return")
			}
		default:
			return l.errorf("unexpected rune")
		}
	}
}

// lexIdentifier scans an identifier
func lexIdentifier(l *lexer) stateFn {
	if r := l.next(); !isAlpha(r) {
		return l.errorf("expected alpha rune as first character")
	}

	for {
		switch r := l.next(); {
		case r == EOF:
			return nil
		case isAlphaNumeric(r):
			// absorb.
		default:
			l.backup()
			return l.emit(itemIdentifier)
		}
	}
}

// lexQuotedString scans a quoted string
func lexQuotedString(l *lexer) stateFn {
	if l.acceptExact('"') == false {
		return l.errorf("quoted-string opening quote expected")
	}

	for {
		switch r := l.next(); {
		case r == EOF:
			return nil
		case isOctetFiltered(r, '\r', '\n', '"', '\\'):
			// absorb
		case r == '\\':
			{
				// quoted-special
				switch next := l.next(); {
				case next == '"':
					// absorb
				case next == '\\':
					// absorb
				default:
					/*
						TODO
							Scripts SHOULD NOT escape other characters with a backslash.
							An undefined escape sequence (such as "\a" in a context where "a" has
							no special meaning) is interpreted as if there were no backslash (in
							this case, "\a" is just "a"), though that may be changed by
							extensions.
					*/
					return l.errorf("quoted-other not supported")
				}
			}
		case r == '\r':
			if l.acceptExact('\n') == false {
				return l.errorf("dangling carriage return")
			}
		case r == '"':
			return l.emit(itemString)
		default:
			return l.errorf("unexpected character")
		}
	}
}

// lexMultiline scans a multi-line string
func lexMultiline(l *lexer) stateFn {
	var endSequence = []rune{'.', '\r', '\n'}

	// input:
	if l.acceptRunStringSequence(textMarker) == false {
		return l.errorf("missing input marker")
	}

	// *(SP / '\t)
	l.acceptRunAny(" \t")

	// *(hash-comment)
	if l.acceptExact('#') {
		for {
			switch r := l.next(); {
			case r == EOF:
				return nil
			case isOctetFiltered(r, '\r', '\n'):
				// absorb
			default:
				l.backup()
				break
			}
		}
	}

	// CRLF
	if l.acceptRunStringSequence("\r\n") == false {
		return l.errorf("CRLF expected")
	}

	// prematurely check if the end sequence was found
	// this is equivalent to an empty multi-line string
	if l.acceptRunSequence(endSequence) {
		return l.emit(itemString)
	}

	for {
		switch r := l.next(); {
		case r == EOF:
			return nil
		case isOctetFiltered(r, '\r', '\n'):
			// absorb
		case r == '\r':
			if l.acceptExact('\n') == false {
				return l.errorf("unexpected carriage return")
			}
			if l.acceptRunSequence(endSequence) {
				return l.emit(itemString)
			}
		default:
			return l.errorf("unexpected rune")
		}
	}
}

// lexStringList scans a string-list open or close tag
func lexStringList(l *lexer) stateFn {
	switch r := l.next(); {
	case r == '[':
		return l.emit(itemStringListOpen)
	case r == ']':
		return l.emit(itemStringListClose)
	}
	return l.errorf("string list open/close expected")
}

// lexTag scans a tag
func lexTag(l *lexer) stateFn {
	if !l.acceptExact(':') {
		return l.errorf("colon expected")
	}
	return lexIdentifier
}

// lexNumeric scans a numerical value (digit w/ optional quantifier)
func lexNumeric(l *lexer) stateFn {
	//    number             = 1*DIGIT [ QUANTIFIER ]

	if !isDigit(l.peek()) {
		return l.errorf("digit expected")
	}

iter:
	for {
		switch r := l.next(); {
		case r == EOF:
			return nil
		case isDigit(r):
			//absorb
		default:
			l.backup()
			break iter
		}
	}

	// accept optional QUANTIFIER
	l.acceptAny([]rune{'K', 'M', 'G'})
	return l.emit(itemNumeric)
}

// lexTestList scans an test-list open and closing tag
func lexTestList(l *lexer) stateFn {
	switch r := l.next(); {
	case r == '(':
		return l.emit(itemTestListOpen)
	case r == ')':
		return l.emit(itemTestListClose)
	}
	return l.errorf("test-list open/close expected")
}

// lexBlock scans an test-list open and closing tag
func lexBlock(l *lexer) stateFn {
	switch r := l.next(); {
	case r == '{':
		return l.emit(itemBlockOpen)
	case r == '}':
		return l.emit(itemBlockClose)
	}
	return l.errorf("block open/close expected")
}
