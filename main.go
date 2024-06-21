package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func main() {
	word := "Vlad"
	width := 736
	height := 736

	_, err := createImage(width, height, word, "OpenSans-SemiBold.ttf", 48)
	if err != nil {
		log.Fatalf("Failed to create image: %v", err)
	}
}

func createImage(width, height int, initials string, fontPath string, fontSize float64) (*image.RGBA, error) {
	f, err := os.Open("bg.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to open background image: %w", err)
	}
	defer f.Close()

	buf := make([]byte, 512)
	_, err = f.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read background image: %v", err)
	}
	contentType := http.DetectContentType(buf)

	_, err = f.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to beginning of file: %v", err)
	}

	var bg image.Image
	switch contentType {
	case "image/jpeg":
		bg, err = jpeg.Decode(f)
	case "image/png":
		bg, err = png.Decode(f)
	default:
		return nil, fmt.Errorf("unsupported image format: %s", contentType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decode background image: %v", err)
	}

	background := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(background, background.Bounds(), bg, image.Point{}, draw.Src)

	face, err := loadFont(fontPath, fontSize)
	if err != nil {
		return nil, fmt.Errorf("failed to load font: %v", err)
	}
	defer face.Close()

	addLabel(background, width/2, height-50, initials, face)

	outFile, err := os.Create("output.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	err = png.Encode(outFile, background)
	if err != nil {
		return nil, fmt.Errorf("failed to encode image to file: %w", err)
	}

	return background, nil
}

func addLabel(img *image.RGBA, x, y int, label string, face font.Face) {
	col := color.RGBA{255, 255, 255, 255}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}

	d.Dot.X -= d.MeasureString(label) / 2
	d.Dot.Y += d.Face.Metrics().Height / 2

	d.DrawString(label)
}

func loadFont(path string, size float64) (font.Face, error) {
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read font file: %v", err)
	}

	fnt, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %v", err)
	}

	face, err := opentype.NewFace(fnt, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create font face: %v", err)
	}

	return face, nil
}
