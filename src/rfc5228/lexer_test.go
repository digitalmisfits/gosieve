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
