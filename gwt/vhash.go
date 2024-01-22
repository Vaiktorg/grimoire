package gwt

import (
	"bytes"
	"errors"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

const (
	hashLen    = 64
	cellSize   = 16
	grid       = cellSize / 2
	sampleSize = cellSize / 4

	vTokWidth  = cellSize * grid
	vTokHeight = cellSize * grid
)

var DefaultConfig = VisualTokenConfig{
	HashGridSrc:  image.Pt(352, 608),
	HashGridRect: image.Rect(0, 0, 128, 128),

	FontFile: "res/Consolas.ttf", // Update with the path to your Consolas font file.,

	HashFontSize: 32.0,
	HashTextSrc:  image.Pt(32, 208),
	HashTextRect: image.Rect(0, 0, 448, 400),

	InfoBlockSrc:  image.Pt(32, 608),
	InfoBlockRect: image.Rect(0, 0, 304, 128),

	MaskImg:     "res/vtkn_mask.png",
	TemplateImg: "res/id_card_tmpl.png",
	HashKey:     HashKey,
}
var SmallConfig = VisualTokenConfig{
	HashGridSrc:  image.Pt(8, 8),
	HashGridRect: image.Rect(0, 0, vTokWidth, vTokHeight),

	FontFile: "res/Consolas.ttf", // Update with the path to your Consolas font file.,

	HashFontSize: 16.0,
	HashTextSrc:  image.Pt(-4, 4),
	HashTextRect: image.Rect(0, 0, vTokWidth, vTokHeight),

	//InfoBlockSrc:  image.Pt(16, 16),
	//InfoBlockRect: image.Rect(0, 0, 128, 128),
	TemplateImg: "res/vtkn_tmpl.png",
	MaskImg:     "res/vtkn_mask.png",
	HashKey:     HashKey,
}

var currCfg = DefaultConfig

type VisualTokenConfig struct {
	HashKey []byte

	HashGridSrc  image.Point
	HashGridRect image.Rectangle

	FontFile     string
	HashFontSize float64

	HashTextSrc  image.Point
	HashTextRect image.Rectangle

	InfoBlockSrc  image.Point
	InfoBlockRect image.Rectangle

	TemplateImg string
	MaskImg     string
}

func SetVTokenConfig(cfg VisualTokenConfig) {
	currCfg = cfg
}

// ====================================================================================================

func (tok *Token) CreateTokenCard() ([]byte, error) {
	hash := XORText(tok.Signature, HashKey)
	if hash == nil || len(hash) < hashLen {
		return nil, errors.New("invalid hash len")
	}

	img := openImage(currCfg.TemplateImg)

	// Generate HashGrid
	err := hashToImage(img, hash)
	if err != nil {
		return nil, err
	}

	// Apply Mask
	img = XORBlend(img, openImage(currCfg.MaskImg))

	//if !currCfg.HashTextRect.Empty() {
	//	err = colorTextHashToImage(img, hash)
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	//Generate XORText
	//if !currCfg.InfoBlockRect.Empty() {
	//err = tok.generateInfoBlock(img, [8]string{
	//	"Issuer: " + string(tok.Header.Issuer),
	//	"Recipient: " + string(tok.Header.Recipient),
	//	"Expire: " + string(tok.Header.Expires.Format("20060102030405")),
	//	"Issuer: " + string(tok.Header.Issuer),
	//	"Recipient: " + string(tok.Header.Recipient),
	//	"Expire: " + string(tok.Header.Expires.Format("20060102030405")),
	//	"Issuer: " + string(tok.Header.Issuer),
	//	"Recipient: " + string(tok.Header.Recipient),
	//})
	//if err != nil {
	//	return nil, err
	//}
	//}

	return hash, saveImage(img)
}

// HashGrid
func hashToImage(img *image.RGBA, data []byte) error {
	return Grid(data, grid, cellSize, func(x, y int, b byte) error {
		squareRect := image.Rect(x, y, x+cellSize, y+cellSize).Add(currCfg.HashGridSrc)
		draw.Draw(img, squareRect, &image.Uniform{C: colorFromByte(b)}, image.Point{}, draw.Src)
		return nil
	})
}

// HashGrid Text
func textHashToImage(img *image.RGBA, hashText []byte, color color.Color) error {
	// Draw each character in a grid.
	return Grid(hashText, grid, cellSize, func(x, y int, b byte) error {
		err := drawText(img, string(b), currCfg.FontFile, currCfg.HashFontSize, color, currCfg.HashTextSrc, x+cellSize, y+cellSize+1)
		if err != nil {
			return err
		}

		return nil
	})
}
func colorTextHashToImage(img *image.RGBA, hashText []byte) error {
	// Draw each character in a grid.
	return Grid(hashText, grid, cellSize, func(x, y int, b byte) error {
		err := drawText(img, string(b), currCfg.FontFile, currCfg.HashFontSize, colorFromByte(b), currCfg.HashTextSrc, x+cellSize, y+cellSize)
		if err != nil {
			return err
		}

		return nil
	})
}

// ====================================================================================================

func (tok *Token) ReadTokenCard(filepath string) ([]byte, error) {
	rgba := openImage(filepath)
	return imageToHash(rgba.SubImage(currCfg.HashGridRect.Add(currCfg.HashGridSrc)).(*image.RGBA))
}
func imageToHash(img *image.RGBA) ([]byte, error) {
	bounds := img.Bounds()

	img = XORBlend(img, openImage(currCfg.MaskImg))
	hash := make([]byte, hashLen)

	return hash, Grid(hash, grid, cellSize, func(x, y int, b byte) error {
		// Adjust the x and y positions relative to the image bounds
		x += bounds.Min.X
		y += bounds.Min.Y

		// Average the colors from the determined sample area
		c := averageColor(img, x, y, sampleSize, sampleSize) // Modify to sample the entire cell
		hash[(y-bounds.Min.Y)/cellSize*grid+(x-bounds.Min.X)/cellSize] = byteFromColor(c)

		return nil
	})
}

// ====================================================================================================

type InfoConfig struct {
	Issuer    string
	Recipient string
	Expires   string
}

func generateInfoBlock(img *image.RGBA, infoLines [8]string) error {
	for i, datum := range infoLines {
		// Calculate the top-left corner of each cell
		x := cellSize / 4
		y := cellSize + (cellSize * i) - x

		// Text, Font, Size, Color, Src, X, Y
		err := drawText(img, datum, currCfg.FontFile, float64(cellSize-x), color.Black, currCfg.InfoBlockSrc, x, y)
		if err != nil {
			return err
		}
	}

	return nil
}
func drawText(img *image.RGBA, text string, fontFile string, FontSize float64, color color.Color, PtSrc image.Point, x, y int) error {
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
	c.SetFontSize(FontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(color))

	// Set the starting point for the text.
	pt := freetype.Pt(x, y).Add(freetype.Pt(PtSrc.X, PtSrc.Y))

	// Draw the text.
	_, err = c.DrawString(text, pt)
	return err
}

// ====================================================================================================

// Image File  I/O
func saveImage(img image.Image) error {
	buff := bytes.NewBuffer([]byte{})
	if err := png.Encode(buff, img); err != nil {
		return err
	}

	return os.WriteFile("id_card.png", buff.Bytes(), 0644)
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

func grayColorFromByte(byteValue byte) color.RGBA {
	return color.RGBA{R: byteValue&0b11111111 - 85, G: byteValue&0b11111111 - 85, B: byteValue&0b11111111 - 85, A: 255}
}
func byteFromGrayColor(c color.RGBA) byte {
	return c.R&0b11111111 + 85
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

// Watermarks
// ----------------------------------------------------------------------------------------------------

// XORBlend applies the XOR operation between two images
func XORBlend(img, mask *image.RGBA) *image.RGBA {
	imgBounds := img.Bounds()
	result := image.NewRGBA(imgBounds)

	for y := imgBounds.Min.Y; y < imgBounds.Max.Y; y++ {
		for x := imgBounds.Min.X; x < imgBounds.Max.X; x++ {
			color1 := img.At(x, y).(color.RGBA)
			color2 := mask.At(x-currCfg.HashGridSrc.X, y-currCfg.HashGridSrc.Y).(color.RGBA)
			result.Set(x, y, xorColors(color1, color2))
		}
	}

	return result
}
func XORText(hash []byte, secretKey []byte) []byte {
	encoded := make([]byte, len(hash))

	for i := range hash {
		encoded[i] = hash[i] ^ secretKey[i%len(secretKey)]
	}

	return encoded
}

// Helper function to apply XOR operation to two colors and return the result
func xorColors(c1, c2 color.RGBA) color.RGBA {
	return color.RGBA{
		R: c1.R ^ c2.R,
		G: c1.G ^ c2.G,
		B: c1.B ^ c2.B,
		A: c1.A, // If the mask contains alpha, you may want to XOR this channel as well
	}
}

// ====================================================================================================

// Grid : I want to use this for every grid in the code
func Grid(hash []byte, cells, cellSize int, h func(x, y int, b byte) error) error {
	for i, b := range hash {
		x := (i % cells) * cellSize
		y := (i / cells) * cellSize

		err := h(x, y, b)
		if err != nil {
			return err
		}
	}

	return nil
}
