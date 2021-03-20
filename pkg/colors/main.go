package colors

import (
	"github.com/gookit/color"
)

type Color interface {
	Sprint(a ...interface{}) string
	Sprintf(format string, a ...interface{}) string
}

var Default Color = color.New(color.FgDefault, color.BgDefault)
var Green Color = color.New(color.FgGreen, color.BgDefault)
var Red Color = color.New(color.FgRed, color.BgDefault)
var Yellow Color = color.New(color.FgYellow, color.BgDefault)
var White Color = color.New(color.FgWhite, color.BgDefault)
var Blue Color = color.New(color.FgBlue, color.BgDefault)
var Cyan Color = color.New(color.FgCyan, color.BgDefault)
var Magenta Color = color.New(color.FgMagenta, color.BgDefault)
var LightWhite Color = color.New(color.FgLightWhite, color.BgDefault)
var LightRed Color = color.New(color.FgLightRed, color.BgDefault)
var Purple Color = color.RGB(186, 85, 211)

// color.RGB(186, 85, 211)

var GreenBold Color = color.New(color.FgGreen, color.BgDefault, color.OpBold)
var RedBold Color = color.New(color.FgRed, color.BgDefault, color.OpBold)
var YellowBold Color = color.New(color.FgYellow, color.BgDefault, color.OpBold)
var WhiteBold Color = color.New(color.FgWhite, color.BgDefault, color.OpBold)
var BlueBold Color = color.New(color.FgBlue, color.BgDefault, color.OpBold)
var CyanBold Color = color.New(color.FgCyan, color.BgDefault, color.OpBold)
var MagentaBold Color = color.New(color.FgMagenta, color.BgDefault, color.OpBold)
var LightWhiteBold Color = color.New(color.FgLightWhite, color.BgDefault, color.OpBold)
var LightRedBold Color = color.New(color.FgLightRed, color.BgDefault, color.OpBold)
var PurpleBold Color = color.NewRGBStyle(color.RGB(186, 85, 211)).SetOpts(color.Opts{color.OpBold})

func InitColors(colored bool) {
	if !colored {
		Green = Default
		Red = Default
		Yellow = Default
		White = Default
		Blue = Default
		Cyan = Default
		Magenta = Default
		LightWhite = Default
		LightRed = Default
		Purple = Default
		GreenBold = Default
		RedBold = Default
		YellowBold = Default
		WhiteBold = Default
		BlueBold = Default
		CyanBold = Default
		MagentaBold = Default
		LightWhiteBold = Default
		LightRedBold = Default
		PurpleBold = Default
	}
}
