// Code generated by fyne-theme-generator

package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type Theme struct{}

func (Theme) Color(c fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch c {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}
	case theme.ColorNameButton:
		return color.Alpha16{A: 0x0}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 0x26, G: 0x26, B: 0x26, A: 0xff}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x42}
	case theme.ColorNameError:
		return color.NRGBA{R: 0xf4, G: 0x43, B: 0x36, A: 0xff}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0x21, G: 0x96, B: 0xf3, A: 0x7f}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	case theme.ColorNameHover:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xf}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x19}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0xb2, G: 0xb2, B: 0xb2, A: 0xff}
	case theme.ColorNamePressed:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x66}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0x21, G: 0x96, B: 0xf3, A: 0xff}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x99}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x66}
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0x0, G: 0xf0, B: 0x0, A: 0x66}
	default:
		return theme.DefaultTheme().Color(c, v)
	}
}

func (Theme) Font(s fyne.TextStyle) fyne.Resource {
	if s.Monospace {
		return theme.DefaultTheme().Font(s)
	}
	if s.Bold {
		if s.Italic {
			return theme.DefaultTheme().Font(s)
		}
		return theme.DefaultTheme().Font(s)
	}
	if s.Italic {
		return theme.DefaultTheme().Font(s)
	}
	return theme.DefaultTheme().Font(s)
}

func (Theme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (Theme) Size(s fyne.ThemeSizeName) float32 {
	switch s {
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNamePadding:
		return 2
	case theme.SizeNameScrollBar:
		return 16
	case theme.SizeNameScrollBarSmall:
		return 6
	case theme.SizeNameSeparatorThickness:
		return 0
	case theme.SizeNameText:
		return 12
	case theme.SizeNameInputBorder:
		return 2
	case theme.SizeNameInnerPadding:
		return 2
	default:
		return theme.DefaultTheme().Size(s)
	}
}
