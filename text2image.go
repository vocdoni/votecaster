package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

const (
	FontsDir   = "fonts/"
	FreeMono   = "FreeMono.ttf"
	FreeSans   = "FreeSans.ttf"
	UbuntuMono = "UbuntuMono-R.ttf"
	Pixeloid   = "PixeloidSans.ttf"
	Inter      = "Inter.ttf"

	BackgroundsDir         = "images/"
	BackgroundAfterVote    = "aftervote.png"
	BackgroundAlreadyVoted = "alreadyvoted.png"
	BackgroundGeneric      = "generic.png"
	BackgroundInfo         = "info.png"
	BackgroundResults      = "results.png"
	BackgroundNotElegible  = "notelegible.png"
	BackgroundError        = "error.png"

	maxBarLength = 12
)

type background struct {
	img             image.Image
	fgColorHex      string
	fontName        string
	fontSize        float64
	xOffset         int
	yOffset         int
	maxTextLineSize int
}

var backgrounds = map[string]*background{
	BackgroundAfterVote:    {nil, "#33ff33", Inter, 50, 10, 30, 20},
	BackgroundAlreadyVoted: {nil, "#ff3333", Inter, 50, 10, 30, 20},
	BackgroundGeneric:      {nil, "#F2EFE5", Inter, 46, 60, 80, 50},
	BackgroundResults:      {nil, "#F2EFE5", Inter, 46, 60, 80, 50},
	BackgroundNotElegible:  {nil, "#ff3333", Inter, 40, 10, 30, 20},
	BackgroundError:        {nil, "#ff3333", Inter, 20, 10, 200, 100},
	BackgroundInfo:         {nil, "#F2EFE5", Inter, 46, 20, 50, 80},
}

func loadImages() error {
	for name, bg := range backgrounds {
		imgFile, err := os.Open(path.Join(BackgroundsDir, name))
		if err != nil {
			return fmt.Errorf("failed to load image %s: %w", name, err)
		}
		defer imgFile.Close()
		img, _, err := image.Decode(imgFile)
		if err != nil {
			return fmt.Errorf("failed to decode image %s: %w", name, err)
		}
		bg.img = img
	}
	return nil
}

func loadFont(fn string) (*truetype.Font, error) {
	fontFile := fmt.Sprintf("fonts/%s", fn)
	fontBytes, err := os.ReadFile(fontFile)
	if err != nil {
		return nil, err
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}
	return f, nil
}

type textToImageContents struct {
	title   string   // Title of the image
	body    []string // Each string is a line of text
	results []string // Same than body, but uses different styles
}

func textToImage(contents textToImageContents, img *background) ([]byte, error) {
	// image size
	// const w = 1685
	// const h = 882
	w := img.img.Bounds().Dx()
	h := img.img.Bounds().Dy()
	// text padding
	const p = 100
	// line spacing
	const ls = 1.5

	// create image
	dc := gg.NewContext(w, h)
	dc.SetRGB(0, 0, 0)
	dc.Clear()
	dc.DrawImage(img.img, 0, 0)

	// load font
	lfont, err := loadFont(img.fontName)
	if err != nil {
		return nil, err
	}

	// text colors
	fgColor := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff} // White by default
	if len(img.fgColorHex) == 7 {
		_, err := fmt.Sscanf(img.fgColorHex, "#%02x%02x%02x", &fgColor.R, &fgColor.G, &fgColor.B)
		if err != nil {
			return nil, err
		}
	}
	dc.SetColor(fgColor)

	// title
	tsize := calculateFontSize(len(contents.title), 100, 40, 250)
	tfont := truetype.NewFace(lfont, &truetype.Options{Size: tsize})
	dc.SetFontFace(tfont)
	dc.DrawStringWrapped(contents.title, p, p, 0, 0, float64(w-(p*2)), ls, gg.AlignLeft)

	// calculate title height
	_, lh := dc.MeasureMultilineString(contents.title, ls)
	tl := dc.WordWrap(contents.title, float64(w-200))
	height := lh*float64(len(tl)) + p*ls

	// body
	if len(contents.body) > 0 {
		bsize := calculateFontSize(len(contents.body), 60, 40, 300)
		bfont := truetype.NewFace(lfont, &truetype.Options{Size: bsize})
		dc.SetFontFace(bfont)
		dc.DrawStringWrapped(strings.Join(contents.body, "\n"), p, p+height, 0, 0, float64(w-(p*2)), ls, gg.AlignLeft)
	}

	// results
	if len(contents.results) > 0 {
		rsize := calculateFontSize(len(contents.body), 50, 20, 400)
		rfont := truetype.NewFace(lfont, &truetype.Options{Size: rsize})
		dc.SetFontFace(rfont)
		dc.DrawStringWrapped(strings.Join(contents.results, "\n"), p, (p/2)+height, 0, 0, float64(w-(p*2)), ls, gg.AlignLeft)
	}

	// return as []byte
	buf := new(bytes.Buffer)
	err = png.Encode(buf, dc.Image())
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func errorImage(err string) ([]byte, error) {
	contents := textToImageContents{
		title: "Error",
		body:  []string{err},
	}
	return textToImage(contents, backgrounds[BackgroundError])
}

func calculateFontSize(stringLength int, maxFontSize, minFontSize, maxStringLength int) float64 {
	// Calculate the scale factor based on the range of font sizes and string lengths
	scaleFactor := float64(maxFontSize-minFontSize) / float64(maxStringLength-1)

	// Calculate the font size using a linear relationship
	fontSize := float64(maxFontSize) - scaleFactor*float64(stringLength-1)

	// Ensure the font size is within the specified bounds
	if fontSize < float64(minFontSize) {
		fontSize = float64(minFontSize)
	} else if fontSize > float64(maxFontSize) {
		fontSize = float64(maxFontSize)
	}

	return fontSize
}

// generateProgressBar generates a progress bar string for the given percentage.
// The progress bar uses '⣿' to represent filled portions.
func generateProgressBar(percentage float64) string {
	filledLength := (int(percentage) * maxBarLength) / 100
	// Generate the filled portion of the progress bar
	filledBar := strings.Repeat("⣿", filledLength)
	// Generate the empty portion of the progress bar
	emptyBar := strings.Repeat(" ", maxBarLength-filledLength)
	return fmt.Sprintf("  %.2f%% %s%s", percentage, filledBar, emptyBar)
}
