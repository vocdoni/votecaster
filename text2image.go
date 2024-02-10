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
	FreeMono   = "FreeMono.ttf"
	FreeSans   = "FreeSans.ttf"
	UbuntuMono = "UbuntuMono-R.ttf"
	Pixeloid   = "PixeloidSans.ttf"
	FontsDir   = "fonts/"
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

func textToImage(textContent string, fgColorHex string, bgColorHex string, fontName string, fontSize float64) ([]byte, error) {
	fgColor := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff} // Default font color
	if len(fgColorHex) == 7 {
		_, err := fmt.Sscanf(fgColorHex, "#%02x%02x%02x", &fgColor.R, &fgColor.G, &fgColor.B)
		if err != nil {
			return nil, err
		}
	}

	bgColor := color.RGBA{R: 0x30, G: 0x0a, B: 0x24, A: 0xff} // Default background color
	if len(bgColorHex) == 7 {
		_, err := fmt.Sscanf(bgColorHex, "#%02x%02x%02x", &bgColor.R, &bgColor.G, &bgColor.B)
		if err != nil {
			return nil, err
		}
	}

	loadedFont, err := loadFont(fontName)
	if err != nil {
		return nil, err
	}

	code := strings.Replace(textContent, "\t", "    ", -1) // convert tabs into spaces
	text := strings.Split(code, "\n")                      // split newlines into arrays

	fg := image.NewUniform(fgColor)
	bg := image.NewUniform(bgColor)
	rgba := image.NewRGBA(image.Rect(0, 0, 1200, 630))
	draw.Draw(rgba, rgba.Bounds(), bg, image.Pt(0, 0), draw.Src)
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(loadedFont)
	c.SetFontSize(fontSize)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	c.SetHinting(font.HintingNone)

	textXOffset := 50
	textYOffset := 40 + int(c.PointToFixed(fontSize)>>6) // Note shift/truncate 6 bits first

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
