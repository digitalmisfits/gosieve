package rfc5228

// Tree is the representation of a single parsed template.
type Tree struct {
	Name      string    // name of the template represented by the tree.
	ParseName string    // name of the top-level template during parsing, for error messages.
	Root      *ListNode // top-level root of the tree.
	Mode      Mode      // parsing mode.
	input     string    // input
	// Parsing only; cleared after parse.
	//funcs     []map[string]any
	lex       *lexer
	token     [3]item // three-token lookahead for parser.
	peekCount int
	//vars       []string // variables defined at the moment.
	treeSet map[string]*Tree
	//actionLine int // line of left delim starting action
	rangeDepth int
}

// A mode value is a set of flags (or 0). Modes control parser behavior.
type Mode uint

// Copy returns a copy of the Tree. Any parsing state is discarded.
func (t *Tree) Copy() *Tree {
	if t == nil {
		return nil
	}
	return &Tree{
		Name:      t.Name,
		ParseName: t.ParseName,
		Root:      t.Root.CopyList(),
		input:     t.input,
	}
}
