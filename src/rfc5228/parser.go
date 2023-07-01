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

import "fmt"

// Tree is the representation of a sieve script
type Tree struct {
	Start
}

func newTree() *Tree {
	return &Tree{}
}

type Start []*CommandNode

func (s *Start) append(node *CommandNode) {
	*s = append(*s, node)
}

// Parser is an eager token stream
type Parser struct {
	Pos
	tokens []item
}

// next advances the position in the token stream
func (p *Parser) next() item {
	// if we read past the end of the input we've reached the end of the file
	if p.isAtEOF() {
		return item{typ: itemEOF, pos: p.Pos, val: "EOF"}
	}

	// advance the pointer after we returned the token @ pos
	defer func() {
		p.Pos += Pos(1)
	}()

	return p.tokens[p.Pos]
}

// peek returns the next token without advancing the position in the token stream
func (p *Parser) peek() item {
	defer func() {
		p.backup()
	}()
	return p.next()
}

func (p *Parser) isAtEOF() bool {
	return p.Pos >= Pos(len(p.tokens))
}
func (p *Parser) advance() {
	_ = p.next()
}

// backup steps back one token
func (p *Parser) backup() {
	if !p.isAtEOF() && p.Pos > 0 {
		p.Pos -= Pos(1)
	}
}

func (p *Parser) accept(typ itemType) bool {
	if token := p.next(); token.typ == typ {
		return true
	}
	p.backup()
	return false
}

// newTokenStream creates a token stream
func newParser(l *lexer) (*Parser, error) {
	var tokens []item

iter:
	for {
		switch token := l.nextItem(); {
		case token.typ == itemError:
			return nil, fmt.Errorf("syntax error: `%s`", token.val)
		case token.typ == itemEOF:
			break iter
		default:
			tokens = append(tokens, token)
		}
	}

	return &Parser{tokens: tokens, Pos: Pos(0)}, nil
}

func (p *Parser) Parse() (*Tree, error) {
	tree := newTree()
	for {
		switch token := p.peek(); token.typ {
		case itemEOF:
			return tree, nil
		case itemComment:
			p.advance() // absorb the peeked token
		case itemIdentifier:
			node, err := p.parseCommand(tree)
			if err != nil {
				return nil, err
			}
			tree.Start.append(&node)
		default:
			return nil, fmt.Errorf("unexpected token")
		}
	}
}

type keyword int

const (
	require keyword = iota
	stop
	keep
	discard
	redirect
)

const (
	IF       = "if"
	REQUIRE  = "require"
	STOP     = "stop"
	KEEP     = "keep"
	DISCARD  = "discard"
	REDIRECT = "redirect"
)

func (p *Parser) parseCommand(tree *Tree) (CommandNode, error) {
	switch token := p.next(); token.typ {
	case itemEOF:
		return nil, nil
	case itemIdentifier:
		var node CommandNode

		switch token.val {
		case IF:
			return p.parseIf(tree)
		case REQUIRE: // require <capabilities: string-list>
			return p.parseRequire(tree)
		case STOP: // stop
			node = tree.newStop(p.Pos)
		case KEEP: // keep
			node = tree.newKeep(p.Pos)
		case DISCARD: // discard
			node = tree.newDiscard(p.Pos)
		case REDIRECT: //  redirect <address: string>
			return p.parseRequire(tree)
		default:
			return nil, fmt.Errorf("uknown identifier %s", token)
		}

		// expect inline handled commands (stop/keep/discard) to end with a ;
		if !p.accept(itemEnd) {
			return nil, fmt.Errorf("expected end `;`")
		}

		return node, nil
	default:
		return nil, fmt.Errorf("unexpected start token %s", token)
	}
}

func (p *Parser) parseRequire(tree *Tree) (CommandNode, error) {
	return nil, fmt.Errorf("not implemented")
}

func (p *Parser) parseRedirect(tree *Tree) (CommandNode, error) {
	return nil, fmt.Errorf("not implemented")
}

func (p *Parser) parseIf(tree *Tree) (CommandNode, error) {
	return nil, fmt.Errorf("not implemented")
}
