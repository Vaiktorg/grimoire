package main

import (
	"github.com/vaiktorg/grimoire/uid"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

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
	pt := freetype.Pt(x+16, y+32)

	// Draw the text.
	_, err = c.DrawString(text, pt)
	return err
}

func main() {
	// Create a new image.
	const imgWidth, imgHeight = 416, 352
	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)

	// Define the text to be drawn.
	text := uid.NewUIDSrc(32, uid.UPPERCASEAlphaNumeric)
	if len(text) > 64 {
		log.Fatalf("Text exceeds 64 characters: %d", len(text))
	}

	// Set the font file.
	fontFile := "Consolas.ttf" // Update with the path to your Consolas font file.

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
			log.Println(err)
			continue
		}
	}

	// Save the image to a file.
	outFile, err := os.Create("output.png")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outFile.Close()
	png.Encode(outFile, img)
}
