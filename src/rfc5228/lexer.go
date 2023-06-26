package rfc5228

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// item represents a token or text string returned from the scanner.
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
	itemError itemType = iota // error occurred; value is text of error
	itemEOF
	itemComment
	itemIdentifier
)

const (
	CR              = '\r'
	LF              = '\n'
	STAR            = '*'
	TAB             = '\t'
	SPACE           = ' '
	HASH            = '#'
	SLASH           = '/'
	COLON           = ':'
	StringListOpen  = '['
	StringListClose = ']'
)

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name  string // name of the lexer; used for error reporting
	input string // the string being scanned
	start Pos    // start position of this token
	pos   Pos    // current position in the input
	atEOF bool   // we have hit the end of input and returned eof
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

// emit passes the trailing text as an item back to the parser.
func (l *lexer) emit(t itemType) stateFn {
	return l.emitItem(l.thisItem(t))
}

// emitItem passes the specified item to the parser.

func (l *lexer) emitItem(i item) stateFn {
	l.item = i
	return nil
}

// next advances the position past the decoded rune
func (l *lexer) next() rune {

	// if we read past the end of the input we've reached the end of the file
	if l.pos >= Pos(len(l.input)) {
		l.width = 0
		return eof
	}

	// decode the string into a rune (utf-8 code points) and advance the position
	r, size := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = size
	l.pos += Pos(size)
	return r
}

// peek does return but does not consume a rune from the input
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune.
func (l *lexer) backup() {
	if !l.atEOF && l.pos > 0 {
		_, w := utf8.DecodeLastRuneInString(l.input[:l.pos])
		l.pos -= Pos(w)
	}
}

// ignore skips over the pending input before this point
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consume the next rune if its from the valid set
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// accept consume a run of runes from the valid set
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
		// consumed
	}
	l.backup()
}

//type RuneRange struct {
//	min, max uint32
//}
//
//func (l *lexer) acceptRanges(rrange ...RuneRange) {
//	for {
//		r := l.next()
//		for _, minMax := range rrange {
//			if r < rune(minMax.min) && r > rune(minMax.max) {
//				l.backup()
//				return
//			}
//		}
//	}
//}

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
// back a nil pointer that will be the next state, terminating l.nextItem.
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
// tabs, newlines (CRLF, never just CR or LF), and the space character.
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

// isOctetNotCRLF tests if a rune is contained within the set %x01-09 / %x0B-0C / %x0E-FF
func isOctetNotCRLF(r rune) bool {
	return (r >= 0x01 && r <= 0x09) ||
		(r >= 0x0B && r <= 0x0C) ||
		(r >= 0x0E && r <= 0xFF)
}

// isOctetNotStarCRLF tests if a rune is contained within the set %x01-09 / %x0B-0C / %x0E-29 / %x2B-FF
func isOctetNotStarCRLF(r rune) bool {
	return (r >= 0x01 && r <= 0x09) ||
		(r >= 0x0B && r <= 0x0C) ||
		(r >= 0x0E && r <= 0x29) ||
		(r >= 0x2B && r <= 0xFF)
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isDigits(r rune) bool {
	return r >= 0x30 && r <= 0x39
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
		switch r := l.next(); {
		case r == eof:
			return l.errorf("end of file reached")
		case isWhitespace(r):
			l.backup() // restore the whitespace rune; the proper action is taken in the state function
			return lexWhitespace
		case isAlphaNumeric(r):
			return lexIdentifier
		default:
			return l.errorf("unexpected rune")
		}
	}
}

// lexWhitespace scans whitespace.
func lexWhitespace(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			return l.errorf("end of file reached")
		case r == SPACE || r == TAB:
			l.ignore()
		case r == CR:
			if next := l.next(); next != LF {
				return l.errorf("unexpected carriage return")
			}
			l.ignore()
		case r == LF:
			return l.errorf("dangling line feed")
		case r == HASH:
			return lexHashComment
		case r == SLASH:
			if next := l.next(); next != STAR {
				return l.errorf("unexpected bracket comment")
			}
			return lexBracketComment
		default:
			l.backup() // restore non matching rune
			return nil
		}
	}
}

// lexBracketComment scans a bracket-comment.
func lexBracketComment(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			return l.errorf("end of file reached")
		case isOctetNotStarCRLF(r):
			// absorb.
		case r == CR:
			if next := l.next(); next != LF {
				return l.errorf("unexpected carriage return")
			}
		case r == STAR:
			if next := l.next(); next != SLASH {
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
		case r == eof:
			return l.errorf("end of file reached")
		case isOctetNotCRLF(r):
			// absorb.
		case r == CR:
			if next := l.next(); next != LF {
				return l.errorf("unexpected carriage return")
			}
			return l.emit(itemComment)
		default:
			return l.errorf("unexpected rune")
		}
	}
}

// lexIdentifier scans an identifier
// (ALPHA / "_") *(ALPHA / DIGIT / "_")
func lexIdentifier(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			return l.errorf("end of file reached")
		case isAlphaNumeric(r) || isDigits(r):
			// absorb.
		default:
			l.backup()
			return l.emit(itemIdentifier)
		}
	}
}

func lexArgument(l *lexer) stateFn {
	/*

		argument     = string-list / number / tag
		string-list  = "[" string *("," string) "]" / string
			                    ; if there is only a single string, the brackets
			                    ; are optional

			number             = 1*DIGIT [ QUANTIFIER ]

			tag                = ":" identifier
	*/

	for {
		switch r := l.next(); {
		case r == eof:
			return l.errorf("end of file reached")
		case r == StringListOpen:
		case isDigits(r):
			break
		case r == COLON:

		}
	}
}

func lexString(l *lexer) stateFn {

	// string       		= quoted-string / multi-line
	// quoted-string      	= DQUOTE quoted-text DQUOTE
	// quoted-text        = *(quoted-safe / quoted-special)

	return nil
}
