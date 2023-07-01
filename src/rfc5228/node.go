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

// A Node is an element in the parse tree. The interface is trivial.
type Node interface {
	Type() NodeType
	Position() Pos
}

// TestCommandNode represents a test parseCommand
//
// A test parseCommand is used as part of a control parseCommand.  It is used to
// specify whether or not the block of code given to the control parseCommand
// is executed.
//
// Since the test parseCommand is part of a control parseCommand,
// we do not consider it an actual parseCommand
type TestCommandNode interface {
	Node
}

// ActionCommandNode represents an action parseCommand
//
// An action parseCommand is an
// identifier followed by zero or more arguments, terminated by a
// semicolon.
type ActionCommandNode interface {
	Node
}

// ControlCommandNode represents a control parseCommand
//
// A control parseCommand is a parseCommand that affects the parsing or the flow
// of execution of the Sieve script in some way.  A control structure is
// da control parseCommand that ends with a block instead of a semicolon.
type ControlCommandNode interface {
	Node
}

// CommandNode represents a node that may exist by itself
type CommandNode interface {
	ControlCommandNode
	ActionCommandNode
}

// NodeType identifies the type of a parse tree node.
type NodeType int

// Type returns itself and provides an easy default implementation
// for embedding in a Node. Embedded in all non-trivial Nodes.
func (t NodeType) Type() NodeType {
	return t
}

const (
	NodeList = iota // A list of Nodes.
	NodeControlRequire
	nodeControlStop
	NodeControlIf
	NodeControlIfElse
	NodeControlElse
	nodeTest
	nodeKeep
	nodeDiscard
	nodeRedirect
	NodeString
	NodeStringList
)

// Pos represents a byte position in the original input input
type Pos int

func (p Pos) Position() Pos {
	return p
}

// // CommandsNode holds a sequence of nodes.
type CommandsNode struct {
	NodeType
	Pos
	Nodes []CommandNode // The element nodes in lexical order.
}

func (t *Tree) newCommands(pos Pos) *CommandsNode {
	return &CommandsNode{NodeType: NodeList, Pos: pos}
}

func (l *CommandsNode) append(n CommandNode) {
	l.Nodes = append(l.Nodes, n)
}

type StopNode struct {
	ActionCommandNode
	NodeType
	Pos
}

func (t *Tree) newStop(pos Pos) *StopNode {
	return &StopNode{NodeType: nodeControlStop, Pos: pos}
}

func (n *StopNode) Type() NodeType {
	return n.NodeType
}

func (n *StopNode) Position() Pos {
	return n.Pos
}

type RequireNode struct {
	ActionCommandNode
	NodeType
	Pos
	Capabilities []string
}

func (t *Tree) newRequire(pos Pos) *RequireNode {
	return &RequireNode{NodeType: NodeControlRequire, Pos: pos}
}

func (n *RequireNode) Type() NodeType {
	return n.NodeType
}

func (n *RequireNode) Position() Pos {
	return n.Pos
}

type KeepNode struct {
	ActionCommandNode
	NodeType
	Pos
}

func (t *Tree) newKeep(pos Pos) *KeepNode {
	return &KeepNode{NodeType: nodeKeep, Pos: pos}
}

func (n *KeepNode) Type() NodeType {
	return n.NodeType
}

func (n *KeepNode) Position() Pos {
	return n.Pos
}

type DiscardNode struct {
	ActionCommandNode
	NodeType
	Pos
}

func (t *Tree) newDiscard(pos Pos) *DiscardNode {
	return &DiscardNode{NodeType: nodeDiscard, Pos: pos}
}

func (n *DiscardNode) Type() NodeType {
	return n.NodeType
}

func (n *DiscardNode) Position() Pos {
	return n.Pos
}

type RedirectNode struct {
	ActionCommandNode
	NodeType
	Pos
	Address string
}

func (t *Tree) newRedirect(pos Pos) *RedirectNode {
	return &RedirectNode{NodeType: nodeRedirect, Pos: pos}
}

func (n *RedirectNode) Type() NodeType {
	return n.NodeType
}

func (n *RedirectNode) Position() Pos {
	return n.Pos
}

type TestNode struct {
	TestCommandNode
	NodeType
	Pos
}

func (t *Tree) newTest(pos Pos) *TestNode {
	return &TestNode{NodeType: nodeTest, Pos: pos}
}

func (n *TestNode) Type() NodeType {
	return n.NodeType
}

func (n *TestNode) Position() Pos {
	return n.Pos
}

type IfNode struct {
	CommandNode
	// fields
	NodeType
	Pos
	Tests   []*TestNode
	Body    *CommandsNode
	ElseIfs []*ElseIfNode
	Else    *ElseNode
}

func (t *Tree) newIf(pos Pos) *IfNode {
	return &IfNode{NodeType: NodeControlIf, Pos: pos}
}

func (n *IfNode) Type() NodeType {
	return n.NodeType
}

func (n *IfNode) Position() Pos {
	return n.Pos
}

type ElseIfNode struct {
	CommandNode

	// fields
	NodeType
	Pos
	Test []*TestNode
	Body *CommandsNode
}

func (t *Tree) newElseIf(pos Pos) *ElseIfNode {
	return &ElseIfNode{NodeType: NodeControlIfElse, Pos: pos}
}

func (n *ElseIfNode) Type() NodeType {
	return n.NodeType
}

func (n *ElseIfNode) Position() Pos {
	return n.Pos
}

type ElseNode struct {
	CommandNode

	// fields
	NodeType
	Pos
	Body []*CommandsNode
}

func (t *Tree) newElse(pos Pos) *ElseNode {
	return &ElseNode{NodeType: NodeControlElse, Pos: pos}
}

func (n *ElseNode) Type() NodeType {
	return n.NodeType
}

func (n *ElseNode) Position() Pos {
	return n.Pos
}
