package rectpack

import (
	"math"
	"slices"
)

type scoreFunc func(width, height int, freeRect *Rect) int

type guillotinePack struct {
	algorithmBase
	Merge       bool
	splitMethod Heuristic

	scoreRect scoreFunc
	freeRects []Rect
}

func newGuillotine(width, height int, heuristic Heuristic) *guillotinePack {
	var packer guillotinePack
	packer.Merge = true
	packer.splitMethod = SplitMinimizeArea

	switch heuristic & fitMask {
	case BestShortSideFit:
		packer.scoreRect = scoreBestShort
	case BestLongSideFit:
		packer.scoreRect = scoreBestLong
	case WorstAreaFit:
		packer.scoreRect = func(w, h int, r *Rect) int { return -scoreBestArea(w, h, r) }
	case WorstShortSideFit:
		packer.scoreRect = func(w, h int, r *Rect) int { return -scoreBestShort(w, h, r) }
	case WorstLongSideFit:
		packer.scoreRect = func(w, h int, r *Rect) int { return -scoreBestLong(w, h, r) }
	default: // BestAreaFit
		packer.scoreRect = scoreBestArea
	}

	packer.splitMethod = heuristic & splitMask
	packer.Reset(width, height)
	return &packer
}

func (p *guillotinePack) Reset(width, height int) {
	p.algorithmBase.Reset(width, height)
	p.freeRects = p.freeRects[:0]
	p.freeRects = append(p.freeRects, NewRect(0, 0, p.maxWidth, p.maxHeight))
}

func (p *guillotinePack) Insert(padding int, sizes ...Size) []Size {
	// Remember variables about the best packing choice we have made so far during the iteration process.
	bestFreeRect := 0
	bestRect := 0
	bestFlipped := false

	// Pack rectangles one at a time until we have cleared the rects array of all rectangles.
	// rects will get destroyed in the process.
	for len(sizes) > 0 {
		// Stores the penalty score of the best rectangle placement - bigger=worse, smaller=better.
		bestScore := math.MaxInt

		for i, freeRect := range p.freeRects {
			for j, size := range sizes {

				padSize(&size, padding)

				// If this rectangle is a perfect match, we pick it instantly.
				if size.Width == freeRect.Width && size.Height == freeRect.Height {
					bestFreeRect = i
					bestRect = j
					bestFlipped = false
					bestScore = math.MinInt
					i = len(p.freeRects) // Force a jump out of the outer loop as well - we got an instant fit.
					break
				} else if p.allowFlip && size.Height == freeRect.Width && size.Width == freeRect.Height {
					// If flipping this rectangle is a perfect match, pick that then.
					bestFreeRect = i
					bestRect = j
					bestFlipped = true
					bestScore = math.MinInt
					i = len(p.freeRects) // Force a jump out of the outer loop as well - we got an instant fit.
					break
				} else if size.Width <= freeRect.Width && size.Height <= freeRect.Height {
					// Try if we can fit the rectangle upright.
					score := p.scoreRect(size.Width, size.Height, &freeRect)
					if score < bestScore {
						bestFreeRect = i
						bestRect = j
						bestFlipped = false
						bestScore = score
					}
				} else if p.allowFlip && size.Height <= freeRect.Width && size.Width <= freeRect.Height {
					// If not, then perhaps flipping sideways will make it fit?
					score := p.scoreRect(size.Height, size.Width, &freeRect)
					if score < bestScore {
						bestFreeRect = i
						bestRect = j
						bestFlipped = true
						bestScore = score
					}
				}
			}
		}

		// If we didn't manage to find any rectangle to pack, abort.
		if bestScore == math.MaxInt {
			break
		}

		// Otherwise, we're good to go and do the actual packing.
		newNode := Rect{
			Point: p.freeRects[bestFreeRect].Point,
			Size:  sizes[bestRect],
		}

		if bestFlipped {
			newNode.Width, newNode.Height = newNode.Height, newNode.Width
			newNode.Flipped = true
		}

		// Remove the free space we lost in the bin.
		p.splitByHeuristic(&p.freeRects[bestFreeRect], &newNode)
		p.freeRects = slices.Delete(p.freeRects, bestFreeRect, bestFreeRect+1)

		// Remove the rectangle we just packed from the input list.
		sizes = slices.Delete(sizes, bestRect, bestRect+1)

		// Perform a Rectangle Merge step if desired.
		if p.Merge {
			p.mergeFreeList()
		}

		// Remember the new used rectangle.
		p.usedArea += newNode.Area()

		unpadRect(&newNode, padding)
		p.packed = append(p.packed, newNode)
	}

	return sizes
}

func scoreBestArea(width, height int, freeRect *Rect) int {
	return freeRect.Width*freeRect.Height - width*height
}

func scoreBestShort(width, height int, freeRect *Rect) int {
	leftoverHoriz := abs(freeRect.Width - width)
	leftoverVert := abs(freeRect.Height - height)
	return min(leftoverHoriz, leftoverVert)
}

func scoreBestLong(width, height int, freeRect *Rect) int {
	leftoverHoriz := abs(freeRect.Width - width)
	leftoverVert := abs(freeRect.Height - height)
	return max(leftoverHoriz, leftoverVert)
}

func (p *guillotinePack) splitAlongAxis(freeRect, placedRect *Rect, splitHorizontal bool) {
	// Form the two new rectangles.
	var bottom Rect
	bottom.X = freeRect.X
	bottom.Y = freeRect.Y + placedRect.Height
	bottom.Height = freeRect.Height - placedRect.Height

	var right Rect
	right.X = freeRect.X + placedRect.Width
	right.Y = freeRect.Y
	right.Width = freeRect.Width - placedRect.Width

	if splitHorizontal {
		bottom.Width = freeRect.Width
		right.Height = placedRect.Height
	} else { // Split vertically
		bottom.Width = placedRect.Width
		right.Height = freeRect.Height
	}

	// Add the new rectangles into the free rectangle pool if they weren't degenerate.
	if bottom.Width > 0 && bottom.Height > 0 {
		p.freeRects = append(p.freeRects, bottom)
	}
	if right.Width > 0 && right.Height > 0 {
		p.freeRects = append(p.freeRects, right)
	}
}

func (p *guillotinePack) findPosition(width, height int, nodeIndex *int) Rect {
	var bestNode Rect

	bestScore := math.MaxInt

	/// Try each free rectangle to find the best one for placement.
	for i, freeRect := range p.freeRects {
		// If this is a perfect fit upright, choose it immediately.
		if width == freeRect.Width && height == freeRect.Height {
			bestNode.X = freeRect.X
			bestNode.Y = freeRect.Y
			bestNode.Width = width
			bestNode.Height = height
			bestScore = math.MinInt
			*nodeIndex = i
			break
		} else if p.allowFlip && height == freeRect.Width && width == freeRect.Height {
			// If this is a perfect fit sideways, choose it.
			bestNode.X = freeRect.X
			bestNode.Y = freeRect.Y
			bestNode.Width = height
			bestNode.Height = width
			bestScore = math.MinInt
			*nodeIndex = i
			break
		} else if width <= freeRect.Width && height <= freeRect.Height {
			// Does the rectangle fit upright?
			score := p.scoreRect(width, height, &freeRect)
			if score < bestScore {
				bestNode.X = freeRect.X
				bestNode.Y = freeRect.Y
				bestNode.Width = width
				bestNode.Height = height
				bestScore = score
				*nodeIndex = i
			}
		} else if p.allowFlip && height <= freeRect.Width && width <= freeRect.Height {
			// Does the rectangle fit sideways?
			score := p.scoreRect(height, width, &freeRect)
			if score < bestScore {
				bestNode.X = freeRect.X
				bestNode.Y = freeRect.Y
				bestNode.Width = height
				bestNode.Height = width
				bestScore = score
				*nodeIndex = i
			}
		}
	}
	return bestNode
}

func (p *guillotinePack) splitByHeuristic(freeRect, placedRect *Rect) {
	// Compute the lengths of the leftover area.
	w := freeRect.Width - placedRect.Width
	h := freeRect.Height - placedRect.Height

	// Placing placedRect into freeRect results in an L-shaped free area, which must be split into
	// two disjoint rectangles. This can be achieved with by splitting the L-shape using a single
	// line. We have two choices: horizontal or vertical.

	// Use the current heuristic to decide which choice to make.

	var splitHorizontal bool
	switch p.splitMethod {
	case SplitShorterLeftoverAxis:
		// Split along the shorter leftover axis.
		splitHorizontal = w <= h
	case SplitLongerLeftoverAxis:
		// Split along the longer leftover axis.
		splitHorizontal = w > h
	case SplitMinimizeArea:
		// Maximize the larger area == minimize the smaller area.
		// Tries to make the single bigger rectangle.
		splitHorizontal = placedRect.Width*h > w*placedRect.Height
	case SplitMaximizeArea:
		// Maximize the smaller area == minimize the larger area.
		// Tries to make the rectangles more even-sized.
		splitHorizontal = placedRect.Width*h <= w*placedRect.Height
	case SplitShorterAxis:
		// Split along the shorter total axis.
		splitHorizontal = freeRect.Width <= freeRect.Height
	case SplitLongerAxis:
		// Split along the longer total axis.
		splitHorizontal = freeRect.Width > freeRect.Height
	default:
		splitHorizontal = true
	}

	// Perform the actual split.
	p.splitAlongAxis(freeRect, placedRect, splitHorizontal)
}

func (p *guillotinePack) mergeFreeList() {
	// Do a Theta(n^2) loop to see if any pair of free rectangles could me merged into one.
	// Note that we miss any opportunities to merge three rectangles into one. (should call this function again to detect that)

	for i := 0; i < len(p.freeRects); i++ {
		for j := i + 1; j < len(p.freeRects); j++ {
			if p.freeRects[i].Width == p.freeRects[i].Width && p.freeRects[i].X == p.freeRects[i].X {
				if p.freeRects[i].Y == p.freeRects[i].Y+p.freeRects[i].Height {
					p.freeRects[i].Y -= p.freeRects[i].Height
					p.freeRects[i].Height += p.freeRects[i].Height
					p.freeRects = slices.Delete(p.freeRects, j, j+1)
					j--
				} else if p.freeRects[i].Y+p.freeRects[i].Height == p.freeRects[i].Y {
					p.freeRects[i].Height += p.freeRects[i].Height
					p.freeRects = slices.Delete(p.freeRects, j, j+1)
					j--
				}
			} else if p.freeRects[i].Height == p.freeRects[i].Height && p.freeRects[i].Y == p.freeRects[i].Y {
				if p.freeRects[i].X == p.freeRects[i].X+p.freeRects[i].Width {
					p.freeRects[i].X -= p.freeRects[i].Width
					p.freeRects[i].Width += p.freeRects[i].Width
					p.freeRects = slices.Delete(p.freeRects, j, j+1)
					j--
				} else if p.freeRects[i].X+p.freeRects[i].Width == p.freeRects[i].X {
					p.freeRects[i].Width += p.freeRects[i].Width
					p.freeRects = slices.Delete(p.freeRects, j, j+1)
					j--
				}
			}
		}
	}
}

// vim: ts=4
