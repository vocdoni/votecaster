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
)

const (
	FontsDir   = "fonts/"
	FreeMono   = "FreeMono.ttf"
	FreeSans   = "FreeSans.ttf"
	UbuntuMono = "UbuntuMono-R.ttf"
	Pixeloid   = "PixeloidSans.ttf"

	BackgroundsDir         = "images/"
	BackgroundAfterVote    = "aftervote.png"
	BackgroundAlreadyVoted = "alreadyvoted.png"
	BackgroundGeneric      = "generic.png"
	BackgroundInfo         = "info.png"
	BackgroundResults      = "results.png"
	BackgroundNotElegible  = "notelegible.png"
	BackgroundError        = "error.png"
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
	BackgroundAfterVote:    {nil, "#33ff33", Pixeloid, 50, 10, 30, 20},
	BackgroundAlreadyVoted: {nil, "#ff3333", Pixeloid, 50, 10, 30, 20},
	BackgroundGeneric:      {nil, "#33ff33", Pixeloid, 50, 30, 30, 60},
	BackgroundResults:      {nil, "#33ff33", Pixeloid, 50, 30, 30, 60},
	BackgroundNotElegible:  {nil, "#ff3333", Pixeloid, 40, 10, 30, 20},
	BackgroundError:        {nil, "#ff3333", Pixeloid, 30, 10, 200, 80},
	BackgroundInfo:         {nil, "#33ff33", Pixeloid, 50, 10, 30, 80},
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

func textToImage(textContent string, img *background) ([]byte, error) {
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

	code := strings.Replace(textContent, "\t", "    ", -1) // convert tabs into spaces
	text := strings.Split(code, "\n")                      // split newlines into arrays

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

	fg := image.NewUniform(fgColor)
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(loadedFont)
	c.SetFontSize(img.fontSize)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	c.SetHinting(font.HintingNone)

	textXOffset := img.xOffset
	textYOffset := img.yOffset + int(c.PointToFixed(img.fontSize)>>6) // Note shift/truncate 6 bits first

	pt := freetype.Pt(textXOffset, textYOffset)
	for _, s := range text {
		_, err = c.DrawString(strings.Replace(s, "\r", "", -1), pt)
		if err != nil {
			return nil, err
		}
		pt.Y += c.PointToFixed(img.fontSize * 1.5)
	}

	b := new(bytes.Buffer)
	if err := png.Encode(b, rgba); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
