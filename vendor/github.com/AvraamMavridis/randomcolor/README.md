# randomcolor

Go package to generate random colors

[![Build Status](https://travis-ci.org/AvraamMavridis/randomcolor.svg?branch=master)](https://travis-ci.org/AvraamMavridis/randomcolor)

```go
package main

import (
	"github.com/AvraamMavridis/randomcolor"
)

func main() {
	var colorInHex string = randomcolor.GetRandomColorInHex()
	var colorInRGB randomcolor.RGBColor = randomcolor.GetRandomColorInRgb()
	var colorInHSV randomcolor.HSVColor = randomcolor.GetRandomColorInHSV()
}
```
