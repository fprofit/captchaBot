package captcha

import (
	_ "embed"
	"fmt"
	"image"
	"math/rand"
	"strconv"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed DejaVuSans-Bold.ttf
var fontBytes []byte

const (
	Width    = 200
	Height   = 60
	FontSize = 20

	NoiseLines  = 15
	NoisePoints = 50
)

type Captcha struct {
	Options       []Option
	Image         image.Image
	correctAnswer string
}

func GenerateCaptcha() (Captcha, error) {
	a := rand.Intn(MaxNumber-MinNumber+1) + MinNumber
	b := rand.Intn(MaxNumber-MinNumber+1) + MinNumber
	img, err := generateImg(a, b, "+")
	if err != nil {
		return Captcha{}, fmt.Errorf("generate CAPTCHA image: %w", err)
	}
	correctAnswer := a + b

	options := make([]Option, CountAnswer)
	usedAnswers := map[int]bool{correctAnswer: true}
	correctIndex := rand.Intn(CountAnswer)

	for i := 0; i < CountAnswer; i++ {
		var answer int

		if i == correctIndex {
			answer = correctAnswer
		} else {
			for {
				answer = rand.Intn(MaxNumber*2-MinNumber*2+1) + MinNumber*2
				if !usedAnswers[answer] {
					usedAnswers[answer] = true
					break
				}
			}
		}

		baseRandom := rand.Uint64()
		maskedRandom := baseRandom ^ uint64(answer)
		token := fmt.Sprintf("%x", maskedRandom)

		options[i] = Option{
			Text:  strconv.Itoa(answer),
			Token: token,
		}
	}

	return Captcha{
		Options:       options,
		Image:         img,
		correctAnswer: options[correctIndex].Token,
	}, nil
}

func generateImg(a, b int, expression string) (image.Image, error) {
	dc := gg.NewContext(Width, Height)

	dc.SetRGB(0.042, 0.042, 0.042)
	dc.Clear()

	ttFont, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("parse embedded font: %w", err)
	}

	face, err := opentype.NewFace(ttFont, &opentype.FaceOptions{
		Size:    FontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("create font face: %w", err)
	}

	dc.SetFontFace(face)

	drawExpression(dc, a, b, expression)
	drawNoise(dc)

	return dc.Image(), nil
}

func drawExpression(dc *gg.Context, a, b int, e string) {
	centerX := float64(Width) / 2
	centerY := float64(Height) / 2

	drawNumber(dc, fmt.Sprint(a), false, centerY)
	drawChar(dc, e, centerX, centerY, 0)
	drawNumber(dc, fmt.Sprint(b), true, centerY)
}

func drawNumber(dc *gg.Context, number string, rightSide bool, y float64) {
	digits := []rune(number)

	if len(digits) == 1 {
		drawChar(dc, string(digits[0]), randomX(rightSide), randomY(y), randomAngle())
		return
	}

	if rightSide {
		drawChar(dc, string(digits[0]), randomRange(118, 125), randomY(y), randomAngle())
		drawChar(dc, string(digits[1]), randomRange(150, 175), randomY(y), randomAngle())
		return
	}

	drawChar(dc, string(digits[0]), randomRange(25, 50), randomY(y), randomAngle())
	drawChar(dc, string(digits[1]), randomRange(75, 82), randomY(y), randomAngle())
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

		dc.DrawLine(randomRange(0, Width), randomRange(0, Height), randomRange(0, Width), randomRange(0, Height))

		dc.Stroke()
	}

	red, green, blue = randomTextColor()
	dc.SetRGBA(red, green, blue, 0.70)

	for i := 0; i < NoisePoints; i++ {
		dc.DrawPoint(randomRange(0, Width), randomRange(0, Height), randomRange(1, 3))

		dc.Fill()
	}
}

func randomTextColor() (float64, float64, float64) {
	red := rand.Intn(110) + 110
	green := rand.Intn(110) + 110
	blue := rand.Intn(110) + 110

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
	return min + rand.Float64()*(max-min)
}
