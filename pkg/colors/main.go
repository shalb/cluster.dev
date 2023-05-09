package colors

import (
	"github.com/gookit/color"
)

// ColoredFmt colored formater interface with base functions.
type ColoredFmt interface {
	Sprint(a ...interface{}) string
	Sprintf(format string, a ...interface{}) string
	Print(a ...interface{})
	Println(a ...interface{})
	Printf(format string, a ...interface{})
}

// Color implementation.
type Color uint8

const (
	Default Color = iota
	Green
	Red
	Yellow
	White
	Blue
	Cyan
	Magenta
	LightWhite
	LightRed
	Purple
	DefaultBold
	GreenBold
	RedBold
	YellowBold
	WhiteBold
	BlueBold
	CyanBold
	MagentaBold
	LightWhiteBold
	LightRedBold
	PurpleBold
)

var colored = true
var colorsMap map[Color]ColoredFmt = map[Color]ColoredFmt{
	Default:        color.New(color.FgDefault, color.BgDefault),
	DefaultBold:    color.New(color.FgDefault, color.BgDefault, color.OpBold),
	Blue:           color.New(color.FgBlue, color.BgDefault),
	BlueBold:       color.New(color.FgBlue, color.BgDefault, color.OpBold),
	Cyan:           color.New(color.FgCyan, color.BgDefault),
	CyanBold:       color.New(color.FgCyan, color.BgDefault, color.OpBold),
	Green:          color.New(color.FgGreen, color.BgDefault),
	GreenBold:      color.New(color.FgGreen, color.BgDefault, color.OpBold),
	LightRed:       color.New(color.FgLightRed, color.BgDefault),
	LightRedBold:   color.New(color.FgLightRed, color.BgDefault, color.OpBold),
	LightWhite:     color.New(color.FgLightWhite, color.BgDefault),
	LightWhiteBold: color.New(color.FgLightWhite, color.BgDefault, color.OpBold),
	Magenta:        color.New(color.FgMagenta, color.BgDefault),
	MagentaBold:    color.New(color.FgMagenta, color.BgDefault, color.OpBold),
	Purple:         color.RGB(186, 85, 211),
	PurpleBold:     color.NewRGBStyle(color.RGB(186, 85, 211)).SetOpts(color.Opts{color.OpBold}),
	Red:            color.New(color.FgRed, color.BgDefault),
	RedBold:        color.New(color.FgRed, color.BgDefault, color.OpBold),
	White:          color.New(color.FgWhite, color.BgDefault),
	WhiteBold:      color.New(color.FgWhite, color.BgDefault, color.OpBold),
	Yellow:         color.New(color.FgYellow, color.BgDefault),
	YellowBold:     color.New(color.FgYellow, color.BgDefault, color.OpBold),
}

// SetColored set all colors to default, if colored == false.
func SetColored(isColored bool) {
	colored = isColored
}

// Fmt return colored formater.
func Fmt(c Color) ColoredFmt {
	if !colored {
		color.Enable = false
		return colorsMap[Default]
	}
	return colorsMap[c]
}
