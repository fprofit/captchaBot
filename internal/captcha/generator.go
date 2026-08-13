package captcha

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	mathrand "math/rand"

	"github.com/fogleman/gg"
)

const (
	Width       = 200
	Height      = 60
	FontSize    = 20
	MinNumber   = 4
	MaxNumber   = 38
	NoiseLines  = 15
	NoisePoints = 50

	MinAnswer = 8
	MaxAnswer = 76
)

type GeneratedChallenge struct {
	Correct string
	Options []Option
	Image   image.Image
}

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(count int) (GeneratedChallenge, error) {
	if count < 1 || count > MaxAnswer-MinAnswer+1 {
		return GeneratedChallenge{}, fmt.Errorf(
			"invalid CAPTCHA option count: %d",
			count,
		)
	}

	challenge, err := generateImage()
	if err != nil {
		return GeneratedChallenge{}, fmt.Errorf(
			"generate CAPTCHA image: %w",
			err,
		)
	}

	correctIndex := mathrand.Intn(count)

	used := map[int]bool{
		challenge.correct: true,
	}

	options := make([]Option, count)

	for i := 0; i < count; i++ {
		answer := challenge.correct

		if i != correctIndex {
			for {
				answer = mathrand.Intn(MaxAnswer-MinAnswer+1) + MinAnswer

				if !used[answer] {
					used[answer] = true
					break
				}
			}
		}

		token, err := randomToken()
		if err != nil {
			return GeneratedChallenge{}, fmt.Errorf(
				"generate CAPTCHA token: %w",
				err,
			)
		}

		options[i] = Option{
			Text:  fmt.Sprintf("%d", answer),
			Token: token,
		}
	}

	return GeneratedChallenge{
		Correct: options[correctIndex].Token,
		Options: options,
		Image:   challenge.Image,
	}, nil
}

type generatedImage struct {
	correct int
	Image   image.Image
}

func generateImage() (generatedImage, error) {
	a := mathrand.Intn(MaxNumber-MinNumber+1) + MinNumber
	b := mathrand.Intn(MaxNumber-MinNumber+1) + MinNumber

	dc := gg.NewContext(Width, Height)

	dc.SetRGB(0.042, 0.042, 0.042)
	dc.Clear()

	if err := dc.LoadFontFace("DejaVuSans-Bold.ttf", FontSize); err != nil {
		return generatedImage{}, fmt.Errorf(
			"load CAPTCHA font: %w",
			err,
		)
	}

	drawExpression(dc, a, b)
	drawNoise(dc)

	return generatedImage{
		correct: a + b,
		Image:   dc.Image(),
	}, nil
}

func randomToken() (string, error) {
	var data [16]byte

	if _, err := cryptorand.Read(data[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(data[:]), nil
}

func drawExpression(dc *gg.Context, a, b int) {
	centerX := float64(Width) / 2
	centerY := float64(Height) / 2

	drawNumber(dc, fmt.Sprint(a), false, centerY)
	drawChar(dc, "+", centerX, centerY, 0)
	drawNumber(dc, fmt.Sprint(b), true, centerY)
}

func drawNumber(dc *gg.Context, number string, rightSide bool, y float64) {
	digits := []rune(number)

	if len(digits) == 1 {
		drawChar(
			dc,
			string(digits[0]),
			randomX(rightSide),
			randomY(y),
			randomAngle(),
		)

		return
	}

	if rightSide {
		drawChar(
			dc,
			string(digits[0]),
			randomRange(118, 125),
			randomY(y),
			randomAngle(),
		)

		drawChar(
			dc,
			string(digits[1]),
			randomRange(150, 175),
			randomY(y),
			randomAngle(),
		)

		return
	}

	drawChar(
		dc,
		string(digits[0]),
		randomRange(25, 50),
		randomY(y),
		randomAngle(),
	)

	drawChar(
		dc,
		string(digits[1]),
		randomRange(75, 82),
		randomY(y),
		randomAngle(),
	)
}

func drawChar(dc *gg.Context, text string, x, y, angle float64) {
	red, green, blue := randomTextColor()

	dc.Push()
	defer dc.Pop()

	dc.Translate(x, y)
	dc.Rotate(angle)

	dc.SetRGB(red, green, blue)
	dc.DrawStringAnchored(text, 0, 0, 0.9, 0.9)
}

func drawNoise(dc *gg.Context) {
	red, green, blue := randomTextColor()
	dc.SetRGBA(red, green, blue, 0.70)

	for i := 0; i < NoiseLines; i++ {
		dc.SetLineWidth(2)

		dc.DrawLine(
			randomRange(0, Width),
			randomRange(0, Height),
			randomRange(0, Width),
			randomRange(0, Height),
		)

		dc.Stroke()
	}

	red, green, blue = randomTextColor()
	dc.SetRGBA(red, green, blue, 0.70)

	for i := 0; i < NoisePoints; i++ {
		dc.DrawPoint(
			randomRange(0, Width),
			randomRange(0, Height),
			randomRange(1, 3),
		)

		dc.Fill()
	}
}

func randomTextColor() (float64, float64, float64) {
	red := mathrand.Intn(110) + 110
	green := mathrand.Intn(110) + 110
	blue := mathrand.Intn(110) + 110

	return float64(red) / 255,
		float64(green) / 255,
		float64(blue) / 255
}

func randomX(right bool) float64 {
	if right {
		return randomRange(118, 175)
	}
	return randomRange(25, 85)
}

func randomY(center float64) float64 {
	return randomRange(center-12, center+12)
}

func randomAngle() float64 {
	return randomRange(-0.8, 0.8)
}

func randomRange(min, max float64) float64 {
	return min + mathrand.Float64()*(max-min)
}
