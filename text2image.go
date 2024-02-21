package main

import (
	"bytes"
	"fmt"
	"image"
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

type fontSize struct {
	min             int
	max             int
	maxStringLength int
}

type section struct {
	fg       string
	font     string
	fontSize *fontSize
	x        float64
	y        float64
}

type theme struct {
	bg      image.Image
	bgColor string
	padding float64
	title   *section
	body    *section
}

var frames = map[string]*theme{
	BackgroundAfterVote:    {},
	BackgroundAlreadyVoted: {},
	BackgroundGeneric: {
		padding: 100,
		title: &section{
			font:     Inter,
			fontSize: &fontSize{min: 40, max: 100, maxStringLength: 250},
		},
		body: &section{
			font:     Inter,
			fontSize: &fontSize{min: 40, max: 60, maxStringLength: 220},
		},
	},
	BackgroundResults: {
		padding: 100,
		title: &section{
			font:     Inter,
			fontSize: &fontSize{min: 40, max: 80, maxStringLength: 250},
		},
		body: &section{
			font:     Inter,
			fontSize: &fontSize{min: 40, max: 50, maxStringLength: 400},
		},
	},
	BackgroundNotElegible: {},
	BackgroundError: {
		title: &section{
			fg:       "#ff3333",
			fontSize: &fontSize{min: 20, max: 30, maxStringLength: 200},
			x:        20,
			y:        200,
		},
	},
	BackgroundInfo: {
		title: &section{
			fg:       "#F2EFE5",
			fontSize: &fontSize{min: 40, max: 60, maxStringLength: 250},
		},
		body: &section{
			fg:       "#F2EFE5",
			fontSize: &fontSize{min: 40, max: 60, maxStringLength: 400},
		},
	},
}

func loadImages() error {
	for name, bg := range frames {
		imgFile, err := os.Open(path.Join(BackgroundsDir, name))
		if err != nil {
			return fmt.Errorf("failed to load image %s: %w", name, err)
		}
		defer imgFile.Close()
		img, _, err := image.Decode(imgFile)
		if err != nil {
			return fmt.Errorf("failed to decode image %s: %w", name, err)
		}
		bg.bg = img
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
	title string   // Title of the image
	body  []string // Each string is a line of text
}

func textToImage(contents textToImageContents, img *theme) ([]byte, error) {
	// image size
	w := 1685
	h := 882
	if img.bg != nil {
		w = img.bg.Bounds().Dx()
		h = img.bg.Bounds().Dy()
	}
	// text padding
	p := float64(100)
	if img.padding > 0 {
		p = img.padding
	}
	// line spacing
	const ls = 1.5

	// create image
	dc := gg.NewContext(w, h)
	if img.bgColor != "" {
		dc.SetHexColor(img.bgColor)
	} else {
		dc.SetRGB(0, 0, 0)
	}
	dc.Clear()
	dc.DrawImage(img.bg, 0, 0)

	// title
	height := float64(0)
	if len(contents.title) > 0 {
		x := img.title.x
		y := img.title.y
		if x == 0 {
			x = p
		}
		if y == 0 {
			y = p
		}
		writeSection(dc, img.title, contents.title, x, y, float64(w)-x*2, ls)
		// calculate title height
		_, lh := dc.MeasureMultilineString(contents.title, ls)
		tl := dc.WordWrap(contents.title, float64(w)-(p*2))
		height = lh*float64(len(tl))*ls + y
	}

	// body
	if len(contents.body) > 0 {
		x := img.body.x
		y := img.body.y
		if x == 0 {
			x = p
		}
		if y == 0 {
			y = height + (ls * 30)
		} else {
			y += height
		}
		writeSection(dc, img.body, strings.Join(contents.body, "\n"), x, y, float64(w)-x*2, ls)
	}

	// return as []byte
	buf := new(bytes.Buffer)
	err := png.Encode(buf, dc.Image())
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writeSection(dc *gg.Context, section *section, contents string, x, y, w, ls float64) {
	// load font
	f := section.font
	if f == "" {
		f = Inter
	}
	font, err := loadFont(f)
	if err != nil {
		fmt.Println(err)
		return
	}

	// text colors
	fgColor := "#ffffff" // White by default
	if len(section.fg) == 7 {
		fgColor = section.fg
	}
	dc.SetHexColor(fgColor)

	// write text
	size := calculateFontSize(len(contents), section.fontSize.max, section.fontSize.min, section.fontSize.maxStringLength)
	face := truetype.NewFace(font, &truetype.Options{Size: size})
	dc.SetFontFace(face)
	dc.DrawStringWrapped(contents, x, y, 0, 0, w, ls, gg.AlignLeft)

}

func errorImage(err error) ([]byte, error) {
	contents := textToImageContents{
		title: err.Error(),
	}
	return textToImage(contents, frames[BackgroundError])
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
