package gwt

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/vaiktorg/grimoire/util"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"
)

const (
	hashLen          = 64
	cellSize         = 16
	cells            = hashLen / (cellSize) * 2
	vgwtWidth        = cellSize * cells
	vgwtHeight       = cellSize * cells
	CardTmplFilepath = "res/id_card_tmpl.png"
)

var (
	HashGridPt   = image.Pt(352, 608)
	HashGridRect = image.Rect(HashGridPt.X, HashGridPt.Y, HashGridPt.X+vgwtWidth, HashGridPt.Y+vgwtHeight)
)

type VisualHash[T any] struct {
	mc *util.MultiCoder[T]
}

func NewVisualHash[T any]() (*VisualHash[T], error) {
	mc, err := util.NewMultiCoder[T]()
	if err != nil {
		return nil, err
	}

	return &VisualHash[T]{
		mc: mc,
	}, nil
}

func (vh *VisualHash[T]) CreateHashCard(obj T) ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})
	err := util.EncodeGob(buff, obj)
	if err != nil {
		return nil, err
	}

	data := []byte(fmt.Sprintf("%x", sha256.Sum256(buff.Bytes())))

	img, err := openImage(CardTmplFilepath)
	if err != nil {
		return nil, err
	}

	// Generate HashGrid
	err = vh.GenerateHashGrid(img, data)
	if err != nil {
		return nil, err
	}

	//vh.GenerateTextHash(img, obj)

	return data[:], saveImage(img)
}
func (vh *VisualHash[T]) HashText(obj T) (code string) {
	lut := "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%&*?+-"
	buff := bytes.NewBuffer([]byte{})

	if err := util.EncodeGob(buff, obj); err != nil {
		return ""
	}

	byteToChar := func(b byte, lut string) rune {
		return rune(lut[b%byte(len(lut))]) // Wrap byte into alphabet range (A-Z)
	}

	hash := buff.Bytes()
	for i, b := range hash {
		width := int(math.Sqrt(float64(len(hash))))
		// Convert each byte to a character
		code += string(byteToChar(b, lut)) + " "
		// Break line every 'width' characters
		if (i+1)%width == 0 {
			code += "\n"
		}
	}

	return code
}
func (vh *VisualHash[T]) GenerateHashGrid(img draw.Image, data []byte) error {
	gridWidth := vgwtWidth / cellSize
	gridHeight := vgwtHeight / cellSize

	for y := 0; y < gridHeight; y++ {
		for x := 0; x < gridWidth; x++ {
			colorIndex := (y*gridWidth + x) % hashLen
			if colorIndex >= len(data) {
				return errors.New("data slice is shorter than expected")
			}
			squareRect := image.Rect(x*cellSize, y*cellSize, (x+1)*cellSize, (y+1)*cellSize).Add(HashGridPt)
			draw.Draw(img, squareRect, &image.Uniform{C: colorFromByte(data[colorIndex])}, image.Point{}, draw.Src)
		}
	}

	return nil
}
func (vh *VisualHash[T]) ReadHashCard(filepath string) ([]byte, error) {
	rgba, err := openImage(filepath)
	if err != nil {
		return nil, err
	}

	return hashFromImage(rgba.SubImage(HashGridRect), cellSize), nil
}

func saveImage(img image.Image) error {
	buff := bytes.NewBuffer([]byte{})
	if err := png.Encode(buff, img); err != nil {
		return err
	}

	return os.WriteFile("id_card.png", buff.Bytes(), 0644)
}
func openImage(filepath string) (*image.RGBA, error) {
	f, err := os.Open(filepath)

	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	rgba, ok := img.(*image.RGBA)
	if !ok {
		return nil, errors.New("image is not in RGBA format")
	}

	return rgba, nil
}
func colorFromByte(byteValue byte) color.RGBA {
	// Distribute the byte's bits across the RGB channels
	r := (byteValue & 0b11100000) >> 5 // Top 3 bits for Red
	g := (byteValue & 0b00011000) >> 3 // Middle 2 bits for Green
	b := byteValue & 0b00000111        // Bottom 3 bits for Blue

	// Scale up to utilize the full color range of each channel
	rScaled := r * 85 // Scaling factor to expand 3 bits to 8 bits
	gScaled := g * 85 // Scaling factor to expand 2 bits to 8 bits
	bScaled := b * 36 // Scaling factor to expand 3 bits to 8 bits

	return color.RGBA{R: rScaled, G: gScaled, B: bScaled, A: 255}
}

func byteFromColor(c color.RGBA) byte {
	// Reverse the scaling and retrieve the original bits
	r := c.R / 85 // Reverse scaling for Red channel
	g := c.G / 85 // Reverse scaling for Green channel
	b := c.B / 36 // Reverse scaling for Blue channel

	// Reassemble the byte from the RGB components
	return (r << 5) | (g << 3) | b
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

// Sample from a small region within each cell rather than a single pixel
func hashFromImage(img image.Image, cellSize int) []byte {
	bounds := img.Bounds()
	hash := make([]byte, hashLen)

	cellWidth := bounds.Dx() / cells
	cellHeight := bounds.Dy() / cells

	for i := 0; i < hashLen; i++ {
		cellX := i % cells
		cellY := i / cells

		// Calculate the top-left corner of each cell
		x := bounds.Min.X + cellX*cellWidth
		y := bounds.Min.Y + cellY*cellHeight

		// Check if the cell is within the bounds of HashGridRect
		if y+cellHeight > bounds.Max.Y {
			log.Fatalf("Cell %d is out of image bounds: Cell Coordinates (%d, %d)\n", i, x, y)
		}

		// Average the colors from the determined sample area
		c := averageColor(img, x, y, cellWidth, cellHeight) // Modify to sample the entire cell
		hash[i] = byteFromColor(c)
	}

	return hash
}

func (vh *VisualHash[T]) GenerateTextHash(img *image.RGBA, obj T) error {
	imgWidth := 416
	imgHeight := 352

	// Define the text to be drawn.
	text := vh.HashText(obj)

	// Set the font file.
	fontFile := "cmd/CONSOLA.ttf" // Update with the path to your Consolas font file.

	// Set font size and grid dimensions.
	fontSize := 32.0 // Adjust as needed to fit your image and grid size.
	gridWidth, gridHeight := 8, 8
	cellWidth, cellHeight := imgWidth/gridWidth, imgHeight/gridHeight

	// Draw each character in a grid.
	for i, char := range text {
		x := (i % gridWidth) * cellWidth
		y := (i / gridWidth) * cellHeight

		err := drawText(img, string(char), fontFile, fontSize, x, y)
		if err != nil {
			return err
		}
	}

	return nil
}
func drawText(img *image.RGBA, text string, fontFile string, fontSize float64, x, y int) error {
	// Read the font data.
	fontBytes, err := os.ReadFile(fontFile)
	if err != nil {
		return err
	}

	fnt, err := truetype.Parse(fontBytes)
	if err != nil {
		return err
	}

	// Initialize the context.
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(fnt)
	c.SetFontSize(fontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.Black)

	// Set the starting point for the text.
	pt := freetype.Pt(x, y).Add(freetype.Pt(48, 224))

	// Draw the text.
	_, err = c.DrawString(text, pt)
	return err
}
