package rfc5228

// Pos represents a byte position in the original input text
type Pos int

func (p Pos) Position() Pos {
	return p
}
