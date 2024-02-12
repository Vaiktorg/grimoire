package vhash

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"github.com/vaiktorg/grimoire/gwt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

type VTokenConfig struct {
	HashKey []byte

	// Authentication Icon GridSquare
	IconGridSrc  image.Point
	IconGridRect image.Rectangle

	TemplateImg string
	PatternImg  string
	ExportPath  string
}

var GridConfig = VTokenConfig{
	IconGridSrc:  image.Pt(8, 8),
	IconGridRect: image.Rect(0, 0, 256, 512),

	TemplateImg: "res/vtkn_tmpl.png",
	PatternImg:  "res/vtkn_mask.png",

	ExportPath: "dist/id_card.png",
	HashKey:    gwt.HashKey,
}

const hashLen = 128

// ====================================================================================================

func CreateTokenCard(hash []byte) (randHash []byte, randMask *image.RGBA, err error) {
	if hash == nil || len(hash) < hashLen {
		return nil, nil, errors.New("invalid hash len")
	}

	// Open Template Image
	img := openImage(GridConfig.TemplateImg)

	// Generate HashGrid
	hashToImage(img, hash)

	// Apply Pattern Mask
	XORBlend(img, openImage(GridConfig.PatternImg))

	// Apply Random Mask
	randHash, randMask = RandomMask()
	XORBlend(img, randMask)

	return randHash, randMask, saveImage(img)
}
func ReadTokenCard(filepath string, randMask *image.RGBA) []byte {
	subImg := openImage(filepath).SubImage(GridConfig.IconGridRect.Add(GridConfig.IconGridSrc)).(*image.RGBA)
	//Remove Random Mask
	XORBlend(subImg, randMask)

	// Remove Pattern Mask
	XORBlend(subImg, openImage(GridConfig.PatternImg))

	return imageToHash(subImg)
}

// HashGrid
// ====================================================================================================

func hashToImage(img *image.RGBA, data []byte) {
	gridX := GridConfig.IconGridRect.Dx() / len(data) * 4
	gridY := GridConfig.IconGridRect.Dy() / len(data) * 2

	cellSizeX := gridX * 4
	cellSizeY := gridY * 4

	Grid(data, cellSizeX, cellSizeY, gridX, gridY, func(x, y int, b byte) {
		// Adjust the x and y positions relative to the image bounds
		x += GridConfig.IconGridSrc.X
		y += GridConfig.IconGridSrc.Y

		squareRect := image.Rect(x, y, x+cellSizeX, y+cellSizeY)
		draw.Draw(img, squareRect, &image.Uniform{C: colorFromByte(b)}, GridConfig.IconGridSrc, draw.Src)
	})
}
func imageToHash(img *image.RGBA) []byte {
	bounds := img.Bounds()

	cellSize := 32
	grid := GridConfig.IconGridRect.Dx() / cellSize
	sampleSizeX := cellSize / 4
	sampleSizeY := cellSize / 4
	hash := make([]byte, hashLen)
	GridSquare(hash, cellSize, grid, func(x, y int, b byte) {
		// Adjust the x and y positions relative to the image bounds
		x += GridConfig.IconGridSrc.X
		y += GridConfig.IconGridSrc.Y

		// Average the colors from the determined sample area
		c := averageColor(img, x, y, sampleSizeX, sampleSizeY) // Modify to sample the entire cell
		hash[(y-bounds.Min.Y)/cellSize*grid+(x-bounds.Min.X)/cellSize] = byteFromColor(c)
	})

	return hash
}

// ====================================================================================================

// Colors
func colorFromByte(byteValue byte) color.RGBA {
	// Map byteValue to color channels (R, G, B) within the 0-255 range
	// Ensure consistency with your image format and encoding logic
	// Example using simple bit shifts:
	r := byteValue >> 4       // Extract high 4 bits for Red
	g := (byteValue >> 2) & 3 // Extract middle 2 bits for Green
	b := byteValue & 3        // Extract low 2 bits for Blue

	return color.RGBA{R: r, G: g, B: b, A: 255} // Set alpha to 255 for opaque
}
func byteFromColor(c color.RGBA) byte {
	// Combine color channels (R, G, B) back into a byte
	// Ensure consistency with colorFromByte's mapping
	// Example using bitwise OR:
	return (c.R << 4) | (c.G << 2) | c.B // Reconstruct byte from channel bits
}
func averageColor(img image.Image, x, y, width, height int) color.RGBA {
	var r, g, b, a, count uint32
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			// Check if pixel coordinates are within the image bounds
			if x+dx >= img.Bounds().Max.X || y+dy >= img.Bounds().Max.Y {
				continue // Skip this pixel
			}
			pixel := img.At(x+dx, y+dy)
			pr, pg, pb, pa := pixel.RGBA()
			r += pr
			g += pg
			b += pb
			a += pa
			count++
		}
	}
	// Calculate the average color; divide by count and by 0x101 to scale 16-bit color to 8-bit color
	r = (r / count) / 0x101
	g = (g / count) / 0x101
	b = (b / count) / 0x101
	a = (a / count) / 0x101

	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
}

// ====================================================================================================

// GridSquare : I want to use this for every grid in the code
func GridSquare(hash []byte, cellSize, gridSize int, h func(x, y int, b byte)) {
	for i, b := range hash {
		x := (i % gridSize) * cellSize
		y := (i / gridSize) * cellSize

		h(x, y, b)
	}
}
func Grid(hash []byte, cellX, cellY, gridX, gridY int, h func(x, y int, b byte)) {
	for i, b := range hash {
		x := (i % gridX) * cellX
		y := (i / gridY) * cellY

		h(x, y, b)
	}
}

// ====================================================================================================

// Image File  I/O
func saveImage(img image.Image) error {
	buff := bytes.NewBuffer([]byte{})
	if err := png.Encode(buff, img); err != nil {
		return err
	}

	return os.WriteFile(GridConfig.ExportPath, buff.Bytes(), 0644)
}
func openImage(filepath string) *image.RGBA {
	f, err := os.Open(filepath)

	if err != nil {
		return nil
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil
	}

	rgba, ok := img.(*image.RGBA)
	if !ok {
		rgba = image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)
	}

	return rgba
}

// ====================================================================================================
// Watermarks
// ----------------------------------------------------------------------------------------------------

// XORBlend applies the XOR operation between two images
func XORBlend(img, mask *image.RGBA) {
	imgBounds := img.Bounds()

	for y := imgBounds.Min.Y; y < imgBounds.Max.Y; y++ {
		for x := imgBounds.Min.X; x < imgBounds.Max.X; x++ {
			color1 := img.At(x, y).(color.RGBA)
			color2 := mask.At(x-GridConfig.IconGridSrc.X, y-GridConfig.IconGridSrc.Y).(color.RGBA)
			img.Set(x, y, xorColors(color1, color2))
		}
	}
}
func xorColors(c1, c2 color.RGBA) color.RGBA {
	return color.RGBA{
		R: c1.R ^ c2.R,
		G: c1.G ^ c2.G,
		B: c1.B ^ c2.B,
		A: c1.A, // If the mask contains alpha, you may want to XOR this channel as well
	}
}

func RandomMask() (mask []byte, maskImg *image.RGBA) {
	// Map byteValue to a color index within the palette
	mask = make([]byte, hashLen*2)
	img := image.NewRGBA(image.Rect(0, 0, 512, 512))

	b := make([]byte, 512*512)
	_, _ = rand.Read(b)

	for i := range b {
		x := i % 512
		y := i / 512
		squareRect := image.Rect(x, y, x+1, y+1)
		draw.Draw(img, squareRect, &image.Uniform{C: colorFromByte(b[i])}, image.Point{}, draw.Src)
		mask = append(mask, b[i])
	}

	hashMask := sha512.Sum512(mask)
	return hashMask[:], img
}
