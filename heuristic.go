package rectpack

import (
	"errors"
	"strings"
)

// Heuristic is a bitfield used for configuration of a rectangle packing algorithm, including the
// general type, bin selection method, and strategy for how to split empty areas. Specific
// combinations of values can be XOR'ed together to achieve the desired behavior.
//
// Note that not not all combinations are valid, each constant of this type will indicate what it
// is valid with. If in doubt, simply use a preset.
//
// To test if a value is valid, use the Validate function, which will return an error message
// describing the issue. When an invalid value is used, the algorithm default will be used, but
// otherwise no error will occur.
type Heuristic uint16

const (
	/**********************************************************************************************
	* Algorithm types
	**********************************************************************************************/

	// MaxRects selects the MaxRects algorithm for packing. This generally results in the
	// most efficiently packed results when packing to a static size. It can result in a lot of
	// waste if a sensible size and bin heustic is not chosen for the given inputs, but has the
	// most potential for efficiency.
	//
	// Type: Algorithm
	MaxRects Heuristic = 0x0

	// Skyline selects the Skyline algorithm for packing. Skyline provides a good balance between
	// speed and efficiency, and is good for maintaining the least amount of waste at any given
	// time, making it a good choice for dynamic data and simply using whatever the final size may
	// be.
	//
	// Type: Algorithm
	Skyline = 0x1

	// Guillotine selects the Guillotine algorithm for packing. This algorithm is typically
	// faster, but is much more sensitive to choosing the correct packing/splitting methods for
	// specific inputs. This makes it less "general-purpose", but can still be the best choice
	// in certain situations where the input sizes are predictable.
	//
	// Type: Algorithm
	Guillotine = 0x2

	/**********************************************************************************************
	* Bin-Selection
	**********************************************************************************************/

	// BestShortSideFit (BSSF) positions the rectangle against the short side of a free rectangle
	// into which it fits the best.
	//
	//	* Type: Bin-Selection
	//	* Valid With: MaxRects, Guillotine
	BestShortSideFit = 0x00
	// BestLongSideFit (BLSF) positions the rectangle against the long side of a free rectangle
	// into which it fits the best.
	//
	//	* Type: Bin-Selection
	//	* Valid With: MaxRects, Guillotine
	BestLongSideFit = 0x10
	// BestAreaFit (BAF) positions the rectangle into the smallest free rect into which it fits.
	//
	//	* Type: Bin-Selection
	//	* Valid With: MaxRects, Guillotine
	BestAreaFit = 0x20
	// BottomLeft (BL) does the Tetris placement.
	//
	//	* Type: Bin-Selection
	//	* Valid With: MaxRects, Skyline
	BottomLeft = 0x30
	// ContactPoint (CP) choosest the placement where the rectangle touches other rects as much
	// as possible.
	//
	//	* Type: Bin-Selection
	//	* Valid With: MaxRects
	ContactPoint = 0x40
	// WorstAreaFit (WAF) is the opposite of the BestAreaFit (BAF) heuristic. Contrary to its
	// name, this is not always "worse" with speciifc inputs.
	//
	//	* Type: Bin-Selection
	//	* Valid With: Guillotine
	WorstAreaFit = 0x50
	// WorstShortSideFit (WSSF) is the opposite of the BestShortSideFit (BSSF) heuristic. Contrary
	// to its name, this is not always "worse" with speciifc inputs.
	//
	//	* Type: Bin-Selection
	//	* Valid With: Guillotine
	WorstShortSideFit = 0x60
	// WorstLongSideFit (WLSF) is the opposite of the BestLongSideFit (BLSF) heuristic. Contrary
	// to its name, this is not always "worse" with speciifc inputs.
	//
	//	* Type: Bin-Selection
	//	* Valid With: Guillotine
	WorstLongSideFit = 0x70
	// MinWaste (MW) uses a "waste map" to split empty spaces and determine which placement will
	// result in the least amount of wasted space. This is most effective when flip/rotate is
	// enabled by the packer.
	//
	//	* Type: Bin-Selection
	//	* Valid With: Skyline
	MinWaste = 0x80

	/**********************************************************************************************
	* Splitting algorithms (only used with guillotine algorithms)
	**********************************************************************************************/

	// SplitShorterLeftoverAxis (SLAS)
	//
	//	* Type: Split Method
	//	* Valid With: Guillotine
	SplitShorterLeftoverAxis = 0x0000

	// SplitLongerLeftoverAxis (LLAS)
	//
	//	* Type: Split Method
	//	* Valid With: Guillotine
	SplitLongerLeftoverAxis = 0x0100

	// SplitMinimizeArea (MINAS) try to make a single big rectangle at the expense of making the
	// other small.
	//
	//	* Type: Split Method
	//	* Valid With: Guillotine
	SplitMinimizeArea = 0x0200

	// SplitMaximizeArea (MAXAS) try to make both remaining rectangles as even-sized as possible.
	//
	//	*Type: Split Method
	//	* Valid With: Guillotine
	SplitMaximizeArea = 0x0300

	// SplitShorterAxis (SAS)
	//
	//	* Type: Split Method
	//	* Valid With: Guillotine
	SplitShorterAxis = 0x0400

	// SplitLongerAxis (LAS)
	//
	//	* Type: Split Method
	//	* Valid With: Guillotine
	SplitLongerAxis = 0x0500

	/**********************************************************************************************
	* Masks for extracting relevant bits
	**********************************************************************************************/

	typeMask  = 0x000F
	fitMask   = 0x00F0
	splitMask = 0x0F00

	/**********************************************************************************************
	* Present combinations of valid heuristics
	**********************************************************************************************/

	// MaxRectsBSSF
	//
	//	* Type: Preset
	MaxRectsBSSF = MaxRects | BestShortSideFit

	// MaxRectsBL
	//
	//	* Type: Preset
	MaxRectsBL = MaxRects | BottomLeft

	// MaxRectsCP
	//
	//	* Type: Preset
	MaxRectsCP = MaxRects | ContactPoint

	// MaxRectsBLSF
	//
	//	* Type: Preset
	MaxRectsBLSF = MaxRects | BestLongSideFit

	// MaxRectsBAF
	//
	//	* Type: Preset
	MaxRectsBAF = MaxRects | BestAreaFit

	// GuillotineBAF
	//
	//	* Type: Preset
	GuillotineBAF = Guillotine | BestAreaFit

	// GuillotineBSSF
	//
	//	* Type: Preset
	GuillotineBSSF = Guillotine | BestShortSideFit

	// GuillotineBLSF
	//
	//	* Type: Preset
	GuillotineBLSF = Guillotine | BestLongSideFit

	// GuillotineWAF
	//
	//	* Type: Preset
	GuillotineWAF = Guillotine | WorstAreaFit

	// GuillotineWSSF
	//
	//	* Type: Preset
	GuillotineWSSF = Guillotine | WorstShortSideFit

	// GuillotineWLSF
	//
	//	* Type: Preset
	GuillotineWLSF = Guillotine | WorstLongSideFit

	// SkylineBLF
	//
	//	* Type: Preset
	SkylineBLF = Skyline | BottomLeft

	// SkylineMinWaste
	//
	//	* Type: Preset
	SkylineMinWaste = Skyline | MinWaste
)

// Algorithm returns the algorithm portion of the bitmask.
func (e Heuristic) Algorithm() Heuristic {
	return e & typeMask
}

// Bin returns the bin selection method portion of the bitmask.
func (e Heuristic) Bin() Heuristic {
	return e & fitMask
}

// Split returns the split method portion of the bitmask.
func (e Heuristic) Split() Heuristic {
	return e & splitMask
}

var (
	algoErr  = errors.New("invalid algorithm type specified")
	splitErr = errors.New("split method heuristic is invalid for algorithm type and will be ignored")
	binErr   = errors.New("bin method heuristic is invalid for algorithm type")
)

// Validate tests whether the combination of heuristics are in good form. A value of nil is
// returned upon success, otherwise an error with message explaining the error.
//
// Note that invalid heuristics will silently fail and cause the packer to revert to its default
// for that setting.
func (e Heuristic) Validate() error {
	bin := e & fitMask
	split := e & splitMask

	switch e & typeMask {
	case MaxRects:
		if split != 0 {
			return splitErr
		}
		switch bin {
		case BestShortSideFit, BestAreaFit, BottomLeft, ContactPoint, BestLongSideFit:
		default:
			return binErr
		}
	case Skyline:
		if split != 0 {
			return splitErr
		}
		switch bin {
		case BottomLeft, MinWaste:
		default:
			return binErr
		}
	case Guillotine:
		switch split {
		case SplitShorterLeftoverAxis, SplitLongerLeftoverAxis, SplitMinimizeArea, SplitMaximizeArea, SplitShorterAxis, SplitLongerAxis:
		default:
			return splitErr
		}
		switch bin {
		case BestShortSideFit, BottomLeft, ContactPoint, BestLongSideFit, BestAreaFit:
		default:
			return splitErr
		}
	default:
		return algoErr
	}

	return nil
}

// String returns the string representation of the heuristic.
func (e Heuristic) String() string {
	var sb strings.Builder
	var split, bin string

	switch e & typeMask {
	case MaxRects:
		sb.WriteString("MaxRects")
	case Skyline:
		sb.WriteString("Skyline")
	case Guillotine:
		sb.WriteString("Guillotine")
		switch e & splitMask {
		case SplitShorterLeftoverAxis:
			split = "-SLAS"
		case SplitLongerLeftoverAxis:
			split = "-LLAS"
		case SplitMinimizeArea:
			split = "-MINAS"
		case SplitMaximizeArea:
			split = "-MAXAS"
		case SplitShorterAxis:
			split = "-SAS"
		case SplitLongerAxis:
			split = "-LAS"
		}
	}

	switch e & fitMask {
	case BestShortSideFit:
		bin = "BSSF"
	case BestLongSideFit:
		bin = "BLSF"
	case BestAreaFit:
		bin = "BAF"
	case BottomLeft:
		bin = "BL"
	case ContactPoint:
		bin = "CP"
	case WorstAreaFit:
		bin = "WAF"
	case WorstShortSideFit:
		bin = "WSSF"
	case WorstLongSideFit:
		bin = "WLSF"
	case MinWaste:
		bin = "MW"
	}

	if bin != "" {
		if sb.Len() > 0 {
			sb.WriteRune('-')
		}
		sb.WriteString(bin)
	}

	if split != "" {
		sb.WriteRune('-')
		sb.WriteString(split)
	}
	return sb.String()
}

// vim: ts=4
