package rfc5228

import (
	"fmt"
	"os"
	"testing"
)

func TestLexer(t *testing.T) {
	dat, _ := os.ReadFile("../../input/comment.sieve")
	lexer := lex("test", string(dat))

	for {
		switch i := lexer.nextItem(); {
		case i.typ == itemError:
			fmt.Printf("error = [%s]\n", i.val)
			return
		case i.typ == itemEOF:
			fmt.Printf("EOF = [%s]\n", i.val)
			return
		default:
			fmt.Printf("%s\n", i)
		}
	}
}
