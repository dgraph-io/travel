package randomcolor

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RGBColor RBG Color Type
type RGBColor struct {
	Red   int
	Green int
	Blue  int
}

// HSVColor HSV Color Type
type HSVColor struct {
	Hue        float64
	Saturation float64
	Value      float64
}

// GetHex Converts a decimal number to hex representations
func getHex(num int) string {
	hex := fmt.Sprintf("%x", num)
	if len(hex) == 1 {
		hex = "0" + hex
	}
	return hex
}

func rgpToHSV(color RGBColor) HSVColor {
	max := math.Max(float64(color.Red), float64(color.Green))
	max = math.Max(max, float64(color.Blue))
	min := math.Min(float64(color.Red), float64(color.Green))
	min = math.Min(min, float64(color.Blue))
	delta := max - min

	red := float64(color.Red)
	green := float64(color.Green)
	blue := float64(color.Blue)

	var hue, saturation, value float64

	if delta == 0 {
		hue = 0
	} else if red == max {
		hue = float64(int(((green - blue) / delta)) % 6)
	} else if green == max {
		hue = (blue-red)/delta + 2
	} else if blue == max {
		hue = (red-green)/delta + 4
	}

	hue = math.Round(hue * 60)
	if hue < 0 {
		hue += 360
	}

	if max == 0 {
		saturation = 0
	} else {
		saturation = (delta / max) * 100
	}

	value = max / 255 * 100

	return HSVColor{hue, saturation, value}
}

// GetRandomColorInRgb Returns a random RGBColor
func GetRandomColorInRgb() RGBColor {
	rand.Seed(time.Now().UnixNano())
	Red := rand.Intn(255)
	Green := rand.Intn(255)
	blue := rand.Intn(255)
	c := RGBColor{Red, Green, blue}
	return c
}

// GetRandomColorInHex returns a random color in HEX format
func GetRandomColorInHex() string {
	color := GetRandomColorInRgb()
	hex := "#" + getHex(color.Red) + getHex(color.Green) + getHex(color.Blue)
	return hex
}

// GetRandomColorInHSV returns a random color in HSV format
func GetRandomColorInHSV() HSVColor {
	color := GetRandomColorInRgb()
	return rgpToHSV(color)
}
