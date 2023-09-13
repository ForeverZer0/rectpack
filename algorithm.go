package rectpack

type packAlgorithm interface {
	// Reset returns the packer to its initial configured state with the specified maximum extents.
	// This function will panic if width or height is less than 1.
	Reset(width, height int)
	// Used computes the ratio of used surface area to the maximum possible area, in the range of
	// 0.0 (empty) and 1.0 (perfectly packed with no waste).
	Used() float64
	// Insert pushes new rectangles into packer, finding the best placement for them based on
	// its configuration and current state.
	//
	// Returns a slice of sizes that could not be packed.
	Insert(sizes ...Size) []Size
	// Rects returns a slice of rectangles that have been packed.
	Rects() []Rect
	// AllowFlip indicates if rectangles can be flipped/rotated to provide better placement.
	//
	// Default: false
	AllowFlip(enabled bool)
	// Padding defines the padding to place around rectangles. Because the padding applies to all
	// sides equally, it will always end up being a power of 2 between rectangles.
	Padding(padding int)
	// MaxSize returns the maximum size the algorithm can pack into.
	MaxSize() Size
	// UsedArea returns the total area that is occupied.
	UsedArea() int
}

type algorithmBase struct {
	packed    []Rect
	maxWidth  int
	maxHeight int
	usedArea  int
	allowFlip bool
	padding   int
}

func (p *algorithmBase) Used() float64 {
	return float64(p.usedArea) / float64(p.maxWidth*p.maxHeight)
}

func (p *algorithmBase) Reset(width, height int) {
	if width <= 0 || height <= 0 {
		panic("width and height must be greater than 0")
	}

	p.maxWidth = width
	p.maxHeight = height
	p.usedArea = 0
	p.packed = p.packed[:0]
}

func (p *algorithmBase) Rects() []Rect {
	return p.packed
}

func (p *algorithmBase) AllowFlip(enabled bool) {
	p.allowFlip = enabled
}

func (p *algorithmBase) Padding(padding int) {
	p.padding = int(padding)
}

func (p *algorithmBase) MaxSize() Size {
	return NewSize(p.maxWidth, p.maxHeight)
}

func (p *algorithmBase) UsedArea() int {
	return p.usedArea
}

func abs(x int) int {
	if x >= 0 {
		return x
	}
	return -x
}

func padSize(size *Size, padding int) {
	if padding <= 0 {
		return
	}
	size.Width += padding
	size.Height += padding
}

func unpadRect(rect *Rect, padding int) {
	if padding <= 0 {
		return
	}

	if rect.X == 0 {
		rect.X += padding
		rect.Width -= padding * 2
	} else {
		rect.Width -= padding
	}

	if rect.Y == 0 {
		rect.Y += padding
		rect.Height -= padding * 2
	} else {
		rect.Height -= padding
	}
}

// vim: ts=4
