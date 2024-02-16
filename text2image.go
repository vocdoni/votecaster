package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
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

	maxBarLength = 16
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
	BackgroundGeneric:      {nil, "#F2EFE5", Inter, 46, 60, 80, 60},
	BackgroundResults:      {nil, "#F2EFE5", Inter, 46, 60, 80, 60},
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
	title string   // Title of the image
	body  []string // Each string is a line of text
}

func textToImage(contents textToImageContents, img *background) ([]byte, error) {
	// Set foreground color
	fgColor := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff} // Default font color
	if len(img.fgColorHex) == 7 {
		_, err := fmt.Sscanf(img.fgColorHex, "#%02x%02x%02x", &fgColor.R, &fgColor.G, &fgColor.B)
		if err != nil {
			return nil, err
		}
	}

	loadedFont, err := loadFont(img.fontName)
	if err != nil {
		return nil, err
	}

	// Prepare the image canvas based on the background image size
	rgba := image.NewRGBA(img.img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img.img, image.Point{}, draw.Src)

	text := prepareImageText(img, contents.title)

	fg := image.NewUniform(fgColor)
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(loadedFont)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	c.SetHinting(font.HintingNone)
	c.SetFontSize(img.fontSize)

	textXOffset := img.xOffset
	textYOffset := img.yOffset + int(c.PointToFixed(img.fontSize)>>6) // Note shift/truncate 6 bits first
	pt := freetype.Pt(textXOffset, textYOffset)

	// draw title
	pt, err = drawTextOnImage(c, img, text, img.fontSize*1.4, pt)
	if err != nil {
		return nil, err
	}
	// draw questions
	for _, q := range contents.body {
		text = prepareImageText(img, q)
		pt, err = drawTextOnImage(c, img, text, img.fontSize, pt)
		if err != nil {
			return nil, err
		}
	}

	b := new(bytes.Buffer)
	if err := png.Encode(b, rgba); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func drawTextOnImage(c *freetype.Context, img *background, text []string, fontSize float64, pt fixed.Point26_6) (fixed.Point26_6, error) {
	c.SetFontSize(fontSize)

	for _, s := range text {
		_, err := c.DrawString(strings.Replace(s, "\r", "", -1), pt)
		if err != nil {
			return pt, err
		}
		pt.Y += c.PointToFixed(img.fontSize * 1.5)
	}
	return pt, nil
}

func prepareImageText(img *background, t string) []string {
	code := strings.Replace(t, "\t", "    ", -1) // convert tabs into spaces
	text := strings.Split(code, "\n")            // split newlines into arrays

	// Check if the text is too long and needs to be split.
	if img.maxTextLineSize > 0 {
		var newText []string
		for _, s := range text {
			if len(s) <= img.maxTextLineSize {
				newText = append(newText, s)
				continue
			}
			// Split the string by words and respect maxTextLineSize
			words := strings.Fields(s)
			line := ""
			for _, w := range words {
				// Check if adding the next word exceeds line length
				if len(line)+len(w) > img.maxTextLineSize {
					if line != "" {
						newText = append(newText, strings.TrimSpace(line)) // Append the current line if it's not empty
					}
					line = w // Start a new line with the current word
				} else {
					// Add a space before the word if the line is not empty
					if line != "" {
						line += " "
					}
					line += w
				}
			}
			if line != "" {
				newText = append(newText, strings.TrimSpace(line)) // Append any remaining text
			}
		}
		text = newText // Replace original text with the reformatted text
	}

	return text
}

// generateProgressBar generates a progress bar string for the given percentage.
// The progress bar uses '⣿' to represent filled portions.
func generateProgressBar(percentage int) string {
	filledLength := (percentage * maxBarLength) / 100
	// Generate the filled portion of the progress bar
	filledBar := strings.Repeat("⣿", filledLength)
	// Generate the empty portion of the progress bar
	emptyBar := strings.Repeat(" ", maxBarLength-filledLength)
	return fmt.Sprintf("  %d%% %s%s", percentage, filledBar, emptyBar)
}
