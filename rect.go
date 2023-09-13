package rectpack

import "fmt"

// Point describes a location in 2D space.
type Point struct {
	// X is the location on the horizontal x-axis.
	X int `json:"x"`
	// Y is the location on the vertical y-axis.
	Y int `json:"y"`
}

// NewPoint initializes a new point with the specified coordinates.
func NewPoint(x, y int) Point {
	return Point{X: x, Y: y}
}

// Eq tests whether the receiver and another point have equal values.
func (p *Point) Eq(point Point) bool {
	return p.X == point.X && p.Y == point.Y
}

// String returns a string representation of the point.
func (p *Point) String() string {
	return fmt.Sprintf("<%v, %v>", p.X, p.Y)
}

// Move will move the location of the receiver to the specified absolute coordinates.
func (p *Point) Move(x, y int) {
	p.X = x
	p.Y = y
}

// Offset will move the location of receiver by the specified relative amount.
func (p *Point) Offset(x, y int) {
	p.X += x
	p.Y += y
}

// Size describes dimensions of an entity in 2D space.
type Size struct {
	// Width is the dimension on the horizontal x-axis.
	Width int `json:"width"`
	// Height is the dimensions on the vertical y-axis.
	Height int `json:"height"`
	// ID is a user-defined identifier that can be used to differentiate this instance from others.
	ID int `json:"-"`
}

// NewSize creates a new size with specified dimensions.
func NewSize(width, height int) Size {
	return Size{Width: width, Height: height}
}

// NewSizeID creates a new size with specified dimensions and unique identifier.
func NewSizeID(id, width, height int) Size {
	return Size{ID: id, Width: width, Height: height}
}

// Eq tests whether the receiver and another size have equal values. The ID field is ignored.
func (sz *Size) Eq(size Size) bool {
	return sz.Width == size.Width && sz.Height == size.Height
}

// String returns a string representation of the size.
func (p *Size) String() string {
	return fmt.Sprintf("<%v, %v>", p.Width, p.Height)
}

// Area returns the total area (width * height).
func (sz *Size) Area() int {
	return sz.Width * sz.Height
}

// Perimeter returns the sum length of all sides.
func (sz *Size) Perimeter() int {
	return (sz.Width + sz.Height) << 1
}

// MaxSide returns the value of the greater side.
func (sz *Size) MaxSide() int {
	return max(sz.Width, sz.Height)
}

// MinSide returns the value of the lesser side.
func (sz *Size) MinSide() int {
	return min(sz.Width, sz.Height)
}

// Ratio compute the ratio between the width/height.
func (sz *Size) Ratio() float64 {
	return float64(sz.Width) / float64(sz.Height)
}

// Rect describes a location (top-left corner) and size in 2D space.
type Rect struct {
	// Point is the location of the rectangle.
	Point
	// Size is the dimensions of the rectangle.
	Size
	// Flipped indicates if a rectangle has been flipped to achieve a better fit while
	// being packed. Only relevant when the packer has AllowFlip enabled.
	Flipped bool `json:"flipped,omitempty"`
}

// NewRect initialzies a new rectangle using the specified point and size values.
func NewRect(x, y, w, h int) Rect {
	return Rect{
		Point: Point{X: x, Y: y},
		Size:  Size{Width: w, Height: h},
	}
}

// NewRectLTRB initializes a new rectangle using  the specified left/top/right/bottom values.
func NewRectLTRB(l, t, r, b int) Rect {
	return Rect{
		Point: Point{X: l, Y: r},
		Size:  Size{Width: r - l, Height: b - t},
	}
}

// Eq compares two rectangles to determine if the location and size is equal.
func (r *Rect) Eq(rect Rect) bool {
	return r.Point.Eq(rect.Point) && r.Size.Eq(rect.Size)
}

// String returns a string describing the rectangle.
func (r *Rect) String() string {
	return fmt.Sprintf("<%v, %v, %v, %v>", r.X, r.Y, r.Width, r.Height)
}

// Left returns the coordinate of the left-edge of the rectangle on the x-axis.
func (r *Rect) Left() int {
	return r.X
}

// Top returns the coordinate of the top-edge of the rectangle on the y-axis.
func (r *Rect) Top() int {
	return r.Y
}

// Right returns the coordinate of the right-edge of the rectangle on the x-axis.
func (r *Rect) Right() int {
	return r.X + r.Width
}

// Bottom returns the coordinate of the bottom-edge of the rectangle on the y-axis.
func (r *Rect) Bottom() int {
	return r.Y + r.Height
}

// TopLeft returns a point representing the top-left corner of the rectangle.
func (r *Rect) TopLeft() Point {
	return Point{X: r.Left(), Y: r.Top()}
}

// TopRight returns a point representing the top-right corner of the rectangle.
func (r *Rect) TopRight() Point {
	return Point{X: r.Right(), Y: r.Top()}
}

// BottomLeft returns a point representing the bottom-left corner of the rectangle.
func (r *Rect) BottomLeft() Point {
	return Point{X: r.Left(), Y: r.Bottom()}
}

// BottomRight returns a point representing the bottom-right corner of the rectangle.
func (r *Rect) BottomRight() Point {
	return Point{X: r.Right(), Y: r.Bottom()}
}

// Centers returns a point representing the center of the rectangle. For rectangles that are a
// power of two, the coordinate is floored.
func (r *Rect) Center() Point {
	return Point{X: r.X + (r.Width >> 1), Y: r.Y + (r.Height >> 1)}
}

// ContainsRect tests whether the specified rectangle is contained within the bounds of the
// current receiver.
func (r *Rect) ContainsRect(rect Rect) bool {
	return r.X <= rect.X &&
		rect.X+rect.Width <= r.X+r.Width &&
		r.Y <= rect.Y &&
		rect.Y+rect.Height <= r.Y+r.Height
}

// Contains tests whether the specified coordinates are within the bounds of the receiver.
func (r *Rect) Contains(x, y int) bool {
	return r.X <= x && x < r.X+r.Width && r.Y <= y && y < r.Y+r.Height
}

// IsEmpty tests whether the width or size of the rectangle is less than 1.
func (r *Rect) IsEmpty() bool {
	return r.Width <= 0 || r.Height <= 0
}

// Inflate pushes each edge of the rectangle out from the center by the specified relative
// amount on each axis.
func (r *Rect) Inflate(width, height int) {
	r.X -= width
	r.Y -= height
	r.Width += (width << 1)
	r.Height += (height << 1)
}

// Intersects tests whether the receiver has any overlap with the specified rectangle.
func (r *Rect) Intersects(rect Rect) bool {
	return rect.X < r.X+r.Width &&
		r.X < rect.X+rect.Width &&
		rect.Y < r.Y+r.Height &&
		r.Y < rect.Y+rect.Height
}

// Intersect returns a rectangle representing only the overlapping area of this rectangle and
// another, or an empty recatangle when no overlap is present.
func (r *Rect) Intersect(rect Rect) (result Rect) {
	x1 := max(r.X, rect.X)
	x2 := min(r.X+r.Width, rect.X+rect.Width)
	y1 := max(r.Y, rect.Y)
	y2 := min(r.Y+r.Height, rect.Y+rect.Height)

	if x2 >= x1 && y2 >= y1 {
		result.Point = Point{X: x1, Y: y1}
		result.Size = Size{Width: x2 - x1, Height: y2 - y1}
	}
	return
}

// Union returns a minimum rectangle required to contain the receiver and another rectangle.
func (r *Rect) Union(rect Rect) Rect {
	x1 := min(r.X, rect.X)
	x2 := max(r.X+r.Width, rect.X+rect.Width)
	y1 := min(r.Y, rect.Y)
	y2 := max(r.Y+r.Height, rect.Y+rect.Height)
	return NewRect(x1, y1, x2-x1, y2-y1)
}

// vim: ts=4
