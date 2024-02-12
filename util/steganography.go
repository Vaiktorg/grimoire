package util

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strconv"
	"strings"
)

const (
	NeighborhoodSize = 3
	Threshold        = 15
)

func encode(message string) error {
	imgFile, _ := os.Open("input.png")
	defer imgFile.Close()
	img, _, _ := image.Decode(imgFile)

	binaryMessage := stringToBinary(message)

	output := image.NewRGBA(img.Bounds())

	idx := 0
	for y := 0; y < img.Bounds().Max.Y; y++ {
		for x := 0; x < img.Bounds().Max.X; x++ {
			if idx >= len(binaryMessage) {
				break
			}
			originalColor := img.At(x, y)
			r, g, b, a := originalColor.RGBA()

			// Simple complexity analysis (e.g., variance in RGB values)
			if isComplex(img, x, y, NeighborhoodSize, Threshold) {
				// Embed more data in complex areas
				r, g, b = embedData(r, g, b, binaryMessage, &idx)
			}

			output.Set(x, y, color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)})
		}
	}

	outFile, _ := os.Create("output.png")
	defer outFile.Close()

	return png.Encode(outFile, output)
}

func stringToBinary(s string) string {
	var binary strings.Builder
	for _, c := range s {
		binary.WriteString(fmt.Sprintf("%08b", c))
	}
	return binary.String()
}

func isComplex(img image.Image, x, y, neighborhoodSize int, threshhold float64) bool {
	// Placeholder for complexity analysis
	// For example, calculate variance and compare with a threshold
	// This function should be elaborated based on actual complexity criteria
	return rgbVariance(img, x, y, neighborhoodSize) > threshhold
}

func rgbVariance(img image.Image, x, y, neighborhood int) float64 {
	var sumR, sumG, sumB, count float64
	var meanR, meanG, meanB float64

	// Define the bounds of the neighborhood
	minX := max(x-neighborhood/2, 0)
	maxX := min(x+neighborhood/2, img.Bounds().Dx()-1)
	minY := max(y-neighborhood/2, 0)
	maxY := min(y+neighborhood/2, img.Bounds().Dy()-1)

	// Calculate the mean of RGB values in the neighborhood
	for i := minX; i <= maxX; i++ {
		for j := minY; j <= maxY; j++ {
			r, g, b, _ := color.RGBAModel.Convert(img.At(i, j)).RGBA()
			sumR += float64(r)
			sumG += float64(g)
			sumB += float64(b)
			count++
		}
	}
	meanR = sumR / count
	meanG = sumG / count
	meanB = sumB / count

	// Calculate the variance of RGB values in the neighborhood
	var variance float64
	for i := minX; i <= maxX; i++ {
		for j := minY; j <= maxY; j++ {
			r, g, b, _ := color.RGBAModel.Convert(img.At(i, j)).RGBA()
			variance += math.Pow(float64(r)-meanR, 2) + math.Pow(float64(g)-meanG, 2) + math.Pow(float64(b)-meanB, 2)
		}
	}
	variance /= count

	return variance
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func embedData(r, g, b uint32, binaryMessage string, idx *int) (uint32, uint32, uint32) {
	// Embed bits into the least significant bits of the pixel
	if *idx < len(binaryMessage) {
		r = setLSB(r, binaryMessage[*idx])
		*idx++
	}
	if *idx < len(binaryMessage) {
		g = setLSB(g, binaryMessage[*idx])
		*idx++
	}
	if *idx < len(binaryMessage) {
		b = setLSB(b, binaryMessage[*idx])
		*idx++
	}
	return r, g, b
}

func setLSB(value uint32, bit byte) uint32 {
	return (value &^ 1) | uint32(bit-'0')
}

// =================================================

func decode(imagePath string) string {
	// Load the encoded image
	imgFile, _ := os.Open(imagePath)
	defer imgFile.Close()
	img, _, _ := image.Decode(imgFile)

	// Decode the message
	var binaryMessage strings.Builder
	for y := 0; y < img.Bounds().Max.Y; y++ {
		for x := 0; x < img.Bounds().Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()

			// Use the same complexity analysis as in encoding
			if isComplex(img, x, y, NeighborhoodSize, Threshold) {
				// Extract bits from complex areas
				binaryMessage.WriteByte(extractLSB(r))
				binaryMessage.WriteByte(extractLSB(g))
				binaryMessage.WriteByte(extractLSB(b))
			}
		}

	}

	// Convert binary string to text
	return binaryToString(binaryMessage.String())
}

func extractLSB(value uint32) byte {
	// Extract the least significant bit
	return byte(value&1) + '0'
}

func binaryToString(binary string) string {
	bytes := make([]byte, len(binary)/8)
	for i := 0; i < len(binary); i += 8 {
		byteValue, _ := strconv.ParseUint(binary[i:i+8], 2, 8)
		bytes[i/8] = byte(byteValue)
	}
	return string(bytes)
}
