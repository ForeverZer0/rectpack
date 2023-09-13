package rectpack

import "slices"

// DefaultSize is the default width/height used as the maximum extent for packing rectangles.
//
// There is based off a maximum texture size for many modern GPUs. If this library is not being
// used for creating a texture atlas, then there is absolutely no significance about this number
// other than providing a sane starting point.
const DefaultSize = 4096

// Packer contains the state of a 2D rectangle packer.
type Packer struct {
	// unpacked contains sizes that have not yet been packed or unable to be packed.
	unpacked []Size
	// algo is the algorithm implementation that performs the actual computation.
	algo packAlgorithm
	// sortFunc contains the function that will be used to determine comparison of sizes
	// when sorting.
	sortFunc SortFunc
	// Padding defines the amount of empty space to place around rectangles. Values of 0 or less
	// indicates that rectangles will be tightly packed.
	//
	// Default: 0
	Padding int
	// sortRev is flag indicating if reverse-ordering of rectangles during sorting should be
	// enabled.
	//
	// Default: false
	sortRev bool
	// Online indicates if rectangles should be packed as they are inserted (online), or simply
	// collected until Pack is called.
	//
	// There is a trade-off to online/offline packing.
	//
	// * Online packing is faster to pack due to a lack of sorting or comparing to
	//	 other rectangles, but results in significantly less optimized results.
	// * Offline packing can be significantly slower, but allows the algorithm to achieve its
	//   maximum potential by having all sizes known ahead of time and sorting for efficiency.
	//
	// Unless you are packing and using the results in real-time, it is recommended to use
	// offline mode (default). For tasks such creating a texture atlas, spending the extra time
	// to prepare the atlas in the most efficient manner is well worth the extra milliseconds
	// of computation.
	//
	// Default: false
	Online bool
}

// Size computes the size of the current packing. The returned value is the minimum size required
// to contain all packed rectangles.
func (p *Packer) Size() Size {
	var size Size
	for _, rect := range p.algo.Rects() {
		size.Width = max(size.Width, rect.Right()+p.Padding)
		size.Height = max(size.Height, rect.Bottom()+p.Padding)
	}
	return size
}

// Insert adds to rectangles to the packer.
//
// When online mode is enabled, the rectangle(s) are immediately packed. The return value will
// contain any values that could not be packed due to size limitations, or an empty slice upon
// success.
//
// When online mode is disabled, the rectangles(s) are simply staged to be packed with the
// next call to Pack. The return value will contain a slice of all rectangles that are currently
// staged.
func (p *Packer) Insert(sizes ...Size) []Size {
	if p.Online {
		return p.algo.Insert(p.Padding, sizes...)
	}

	p.unpacked = append(p.unpacked, sizes...)
	return p.unpacked
}

// Insert adds to rectangles to the packer.
//
// When online mode is enabled, the rectangle(s) are immediately packed. The return value will
// contain any values that could not be packed due to size limitations, or an empty slice upon
// success.
//
// When online mode is disabled, the rectangles(s) are simply staged to be packed with the
// next call to Pack. The return value will contain a slice of all rectangles that are currently
// staged.
func (p *Packer) InsertSize(id, width, height int) bool {
	result := p.Insert(NewSizeID(id, width, height))
	if p.Online && len(result) != 0 {
		return false
	}
	return true
}

// Sorter sets the comparer function used for pre-sorting sizes before packing. Depending on
// the algorithm and the input data, this can provide a significant improvement on efficiency.
//
// Default: SortArea
func (p *Packer) Sorter(compare SortFunc, reverse bool) {
	p.sortFunc = compare
	p.sortRev = reverse
}

// Rects returns a slice of rectangles that are currently packed.
//
// The backing memory is owned by the packer, and a copy should be made if modification or
// persistence is required.
func (p *Packer) Rects() []Rect {
	return p.algo.Rects()
}

// Unpacked returns a slice of rectangles that are currently staged to be packed.
//
// The backing memory is owned by the packer, and a copy should be made if modification or
// persistence is required.
func (p *Packer) Unpacked() []Size {
	return p.unpacked
}

// Used computes the ratio of used surface area to the available area, in the range of
// 0.0 and 1.0.
//
// When current is set to true, the ratio will reflect the ratio of used surface area relative
// to the current size required by the packer, otherwise it is the ratio of the maximum
// possible area.
func (p *Packer) Used(current bool) float64 {
	if current {
		size := p.Size()
		return float64(p.algo.UsedArea()) / float64(size.Width * size.Height)
	}
	return p.algo.Used()
}

// Map creates and returns a map where each key is an ID, and the value is the rectangle it
// pertains to.
func (p *Packer) Map() map[int]Rect {
	rects := p.algo.Rects()
	mapping := make(map[int]Rect, len(rects))
	for _, rect := range rects {
		mapping[rect.ID] = rect
	}

	return mapping
}

// Clear resets the internal state of the packer without changing its current configuration. All
// currently packed and pending rectangles are removed.
func (p *Packer) Clear() {
	size := p.algo.MaxSize()
	p.algo.Reset(size.Width, size.Height)
	p.unpacked = p.unpacked[:0]
}

// Pack will sort and pack all rectangles that are currently staged.
//
// The return value indicates if all staged rectangles were successfully packed. When false,
// Unpacked can be used to retrieve the sizes that failed.
func (p *Packer) Pack() bool {
	if len(p.unpacked) == 0 {
		return true
	}

	if p.sortFunc != nil {
		if p.sortRev {
			slices.SortFunc(p.unpacked, func(a, b Size) int {
				return p.sortFunc(b, a)
			})
		} else {
			slices.SortFunc(p.unpacked, p.sortFunc)
		}
	} else if p.sortRev {
		slices.Reverse(p.unpacked)
	}

	failed := p.algo.Insert(p.Padding, p.unpacked...)
	if len(failed) == 0 {
		p.unpacked = p.unpacked[:0]
		return true
	}

	p.unpacked = failed
	return false
}

// RepackAll clears the internal packed rectangles, and repacks them all with one operation. This
// can be useful to optimize the packing when/if it was previously performed in multiple pack
// operations, or to reflect settings for the packer that have been modified. 
func (p *Packer) RepackAll() bool {
	rects := p.algo.Rects()
	for _, rect := range rects {
		p.unpacked = append(p.unpacked, rect.Size)
	}
	
	size := p.Size()
	p.algo.Reset(size.Width, size.Height)
	return p.Pack()
}

// AllowFlip indicates if rectangles can be flipped/rotated to provide better placement.
//
// Default: false
func (p *Packer) AllowFlip(enabled bool) {
	p.algo.AllowFlip(enabled)
}

// NewPacker initializes a new Packer using the specified maximum size and heustistics for
// packing rectangles.
func NewPacker(maxWidth, maxHeight int, heuristic Heuristic) *Packer {
	p := &Packer{
		Online:   false,
		sortFunc: SortArea,
		sortRev:  false,
	}

	switch heuristic & typeMask {
	case MaxRects:
		p.algo = newMaxRects(maxWidth, maxHeight, heuristic)
	case Skyline:
		p.algo = newSkyline(maxWidth, maxHeight, heuristic)
	case Guillotine:
		p.algo = newGuillotine(maxWidth, maxHeight, heuristic)
	default:
		panic("heuristics specify invalid argorithm")
	}

	return p
}

// NewDefaultPacker initializes a new Packer with sensible default settings suitable for
// general-purpose rectangle packing.
func NewDefaultPacker() *Packer {
	return NewPacker(DefaultSize, DefaultSize, MaxRectsBSSF)
}

// vim: ts=4
