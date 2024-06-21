package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"math/rand"

	// "image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter the word or phrase: ")
	word, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read input: %v", err)
	}
	word = strings.TrimSpace(word)

	width := 600
	height := 600
	fontSize := 64.0
	outputFile := "output.png"

	_, err = createImage(width, height, word, "OpenSans-SemiBold.ttf", fontSize, outputFile)
	if err != nil {
		log.Fatalf("Failed to create image: %v", err)
	}

	fmt.Printf("Image created successfully as %s\n", outputFile)
}

func createImage(width, height int, initials string, fontPath string, fontSize float64, outputPath string) (*image.RGBA, error) {
	randomNum := rand.Intn(9) + 1
	filename := fmt.Sprintf("images/%02d.jpg", randomNum)

	f, err := os.Open(filename)
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
	// draw.Draw(background, background.Bounds(), bg, image.Point{}, draw.Src)
	draw.ApproxBiLinear.Scale(background, background.Bounds(), bg, bg.Bounds(), draw.Over, nil)

	face, err := loadFont(fontPath, fontSize)
	if err != nil {
		return nil, fmt.Errorf("failed to load font: %v", err)
	}
	defer face.Close()

	addLabel(background, width/2, height, initials, face)

	outFile, err := os.Create(outputPath)
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
	bgColor := color.RGBA{0, 0, 0, 128}

	col := color.RGBA{255, 255, 255, 255}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}

	d.Dot.X -= d.MeasureString(label) / 2
	d.Dot.Y -= fixed.I(30)

	padding := fixed.I(30)
	bgRect := image.Rect(
		0,
		fixed.Int26_6(d.Dot.Y-face.Metrics().Height/2-padding).Floor(),
		x*2,
		fixed.Int26_6(d.Dot.Y+padding).Ceil(),
	)
	draw.Draw(img, bgRect, &image.Uniform{bgColor}, image.Point{}, draw.Over)

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
