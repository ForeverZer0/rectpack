package rectpack

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func createAtlas(p *Packer, paths []string) (*image.RGBA, map[string]Rect, error) {
	// Reset the packer to its initial state
	p.Clear()

	// Enumerate each given path and decode its header to retrieve the size
	for i, path := range paths {
		var cfg image.Config
		if file, err := os.Open(path); err != nil {
			return nil, nil, err
		} else {
			cfg, _, err = image.DecodeConfig(file)
			file.Close()
			if err != nil {
				return nil, nil, err
			}
		}

		// Insert the size into the packer, using the index as the ID
		if !p.InsertSize(i, cfg.Width, cfg.Height) && p.Online {
			// If packing in online mode, ensure each image fits as it is inserted.
			size := p.Size()
			return nil, nil, fmt.Errorf("cannot fit all images into size of %s", size.String())
		}
	}

	// If packing in offline mode, perform tha packing and ensure all images were packed.
	if !p.Online && !p.Pack() {
		size := p.Size()
		return nil, nil, fmt.Errorf("cannot fit all images into size of %s", size.String())
	}

	// Get the final required size of the atlas (includes any configured padding), and create
	// a new image of that size to draw onto.
	size := p.Size()
	mapping := make(map[string]Rect, len(paths))
	dst := image.NewRGBA(image.Rect(0, 0, size.Width, size.Height))
	var zero image.Point

	// Iterate through each packed rectangle.
	for _, rect := range p.Rects() {
		// The ID of the rectangle is the index of the path (assigned above).
		path := paths[rect.ID]
		file, err := os.Open(path)
		if err != nil {
			return nil, nil, err
		}

		// Decode the image at that path
		src, _, err := image.Decode(file)
		file.Close()
		if err != nil {
			return nil, nil, err
		}

		// Draw the image onto the destination at the rectangles location.
		bounds := image.Rect(rect.X, rect.Y, rect.Right(), rect.Bottom())
		draw.Draw(dst, bounds, src, zero, draw.Src)

		// Map the path to the rectangle where the image is drawn at.
		mapping[path] = rect
	}

	// Return the results
	return dst, mapping, nil
}

// randomSize returns a size within the given minimum and maximum sizes.
func randomSize(id int, minSize, maxSize Size) Size {
	w := rand.Intn(maxSize.Width-minSize.Width) + minSize.Width
	h := rand.Intn(maxSize.Height-minSize.Height) + minSize.Height
	return NewSizeID(id, w, h)
}

// randomColor (surprise!) returns a random color.
func randomColor() color.RGBA {
	// Offset to use a minimum value so it is never pure black.
	return color.RGBA{
		R: uint8(rand.Intn(240)) + 15,
		G: uint8(rand.Intn(240)) + 15,
		B: uint8(rand.Intn(240)) + 15,
		A: 255,
	}
}

// createImage colorizes and creates an image from packed rectangles to provide
// a visual representation.
func createImage(t *testing.T, path string, packer *Packer) {
	black := color.RGBA{0, 0, 0, 255}
	size := packer.Size()
	img := image.NewRGBA(image.Rect(0, 0, size.Width, size.Height))
	draw.Draw(img, img.Bounds(), &image.Uniform{black}, image.Point{}, draw.Src)

	for _, rect := range packer.Rects() {
		color := randomColor()
		r := image.Rect(rect.X, rect.Y, rect.Right(), rect.Bottom())
		draw.Draw(img, r, &image.Uniform{color}, image.Point{0, 0}, draw.Src)
	}

	if file, err := os.Create(path); err == nil {
		defer file.Close()
		png.Encode(file, img)
	} else {
		t.Fatal(err)
	}
}

func TestAtlas(t *testing.T) {
	return
	paths, _ := filepath.Glob("/usr/share/icons/Adwaita/32x32/devices/*.png")
	packer := NewPacker(512, 512, MaxRectsBAF)

	img, mapping, err := createAtlas(packer, paths)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	for k, v := range mapping {
		fmt.Printf("%v: %s\n", k, v.String())
	}

	file, _ := os.Create("atlas.png")
	defer file.Close()
	png.Encode(file, img)
}

func TestRandom(t *testing.T) {
	const count = 1024
	minSize := NewSize(32, 32)
	maxSize := NewSize(96, 96)

	packer := NewPacker(1024, 8192, MaxRectsBSSF)
	packer.AllowFlip(true)
	packer.Online = false
	packer.Padding = 2
	packer.Sorter(SortArea, false)

	sizes := make([]Size, count)
	for i := 0; i < count; i++ {
		sizes[i] = randomSize(i, minSize, maxSize)
	}

	packer.Insert(sizes...)
	if !packer.Pack() {
		t.Fatal("cannot fit all rectangles in the given dimensions")
	}

	packer.Pack()
	// Compare every rectangle to every other and test for intersection
	rects := packer.Rects()
	for i := 0; i < len(rects)-1; i++ {
		for j := i + 1; j < len(rects); j++ {
			if rects[i].Intersects(rects[j]) {
				t.Errorf("%s and %s intersect\n", rects[i].String(), rects[j].String())
			}
		}
	}

	createImage(t, "packed.png", packer)
}

// vim: ts=4
