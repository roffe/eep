package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

//go:embed pcb.jpg
var pcbBytes []byte

//go:embed eeprom.jpg
var eepromBytes []byte

//go:embed overview.jpg
var overviewBytes []byte

var (
	PCB = &canvas.Image{
		ScaleMode: canvas.ImageScalePixels,
		FillMode:  canvas.ImageFillOriginal,
		Resource: &fyne.StaticResource{
			StaticName:    "pcb.jpg",
			StaticContent: pcbBytes},
	}
	EEPROM = &canvas.Image{
		ScaleMode: canvas.ImageScalePixels,
		FillMode:  canvas.ImageFillOriginal,
		Resource: &fyne.StaticResource{
			StaticName:    "eeprom.jpg",
			StaticContent: eepromBytes},
	}
	OVERVIEW = &canvas.Image{
		ScaleMode: canvas.ImageScalePixels,
		FillMode:  canvas.ImageFillOriginal,
		Resource: &fyne.StaticResource{
			StaticName:    "overview.jpg",
			StaticContent: overviewBytes},
	}
)
