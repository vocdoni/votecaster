package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

const (
	FreeMono            = "FreeMono.ttf"
	FreeSans            = "FreeSans.ttf"
	UbuntuMono          = "UbuntuMono-R.ttf"
	Pixeloid            = "PixeloidSans.ttf"
	FontsDir            = "fonts/"
	BackgroundImagePath = FontsDir + "background.png"
)

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

func textToImage(textContent string, fgColorHex string, bgImagePath string, fontName string, fontSize float64) ([]byte, error) {
	// Load the background image
	bgImgFile, err := os.Open(bgImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load background image: %w", err)
	}
	defer bgImgFile.Close()

	bgImg, _, err := image.Decode(bgImgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode background image: %w", err)
	}

	// Set foreground color
	fgColor := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff} // Default font color
	if len(fgColorHex) == 7 {
		_, err := fmt.Sscanf(fgColorHex, "#%02x%02x%02x", &fgColor.R, &fgColor.G, &fgColor.B)
		if err != nil {
			return nil, err
		}
	}

	loadedFont, err := loadFont(fontName)
	if err != nil {
		return nil, err
	}

	// Prepare the image canvas based on the background image size
	rgba := image.NewRGBA(bgImg.Bounds())
	draw.Draw(rgba, rgba.Bounds(), bgImg, image.Point{}, draw.Src)

	code := strings.Replace(textContent, "\t", "    ", -1) // convert tabs into spaces
	text := strings.Split(code, "\n")                      // split newlines into arrays

	fg := image.NewUniform(fgColor)
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(loadedFont)
	c.SetFontSize(fontSize)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	c.SetHinting(font.HintingNone)

	textXOffset := 140
	textYOffset := 180 + int(c.PointToFixed(fontSize)>>6) // Note shift/truncate 6 bits first

	pt := freetype.Pt(textXOffset, textYOffset)
	for _, s := range text {
		_, err = c.DrawString(strings.Replace(s, "\r", "", -1), pt)
		if err != nil {
			return nil, err
		}
		pt.Y += c.PointToFixed(fontSize * 1.5)
	}

	b := new(bytes.Buffer)
	if err := png.Encode(b, rgba); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
