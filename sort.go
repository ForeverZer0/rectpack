package rectpack

import "cmp"

// SortFunc is a prototype for a funcion that compares two rectangle sizes, returning standard
// comparer result of -1 for less-than, 1 for greater-than, or 0 for equal to.
type SortFunc func(a, b Size) int

// SortArea sorts two rectangle sizes in descending order (greatest to least) by comparing the
// total area of each.
func SortArea(a, b Size) int {
	return cmp.Compare(b.Area(), a.Area())
}

// SortPerimeter sorts two rectangle sizes in descending order (greatest to least) by comparing
// the perimeter of each.
func SortPerimeter(a, b Size) int {
	return cmp.Compare(b.Perimeter(), a.Perimeter())
}

// SortDiff sorts two rectangle sizes in descending order (greatest to least) by comparing the
// difference between the width/height of each.
func SortDiff(a, b Size) int {
	return cmp.Compare(abs(b.Width-b.Height), abs(a.Width-a.Height))
}

// SortMinSide sorts two rectangle sizes in descending order (greatest to least) by comparing the
// shortest side of each.
func SortMinSide(a, b Size) int {
	return cmp.Compare(b.MinSide(), a.MinSide())
}

// SortMaxSide sorts two rectangle sizes in descending order (greatest to least) by comparing the
// longest side of each.
func SortMaxSide(a, b Size) int {
	return cmp.Compare(b.MaxSide(), a.MaxSide())
}

// SortRatio sorts two rectangle sizes in descending order (greatest to least) by comparing the
// ratio between the width/height of each.
func SortRatio(a, b Size) int {
	return cmp.Compare(b.Ratio(), a.Ratio())
}

// vim: ts=4
