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
		ScaleMode: canvas.ImageScaleSmooth,
		FillMode:  canvas.ImageFillOriginal,
		Resource: &fyne.StaticResource{
			StaticName:    "pcb.png",
			StaticContent: pcbBytes},
	}
	EEPROM = &canvas.Image{
		ScaleMode: canvas.ImageScaleSmooth,
		FillMode:  canvas.ImageFillOriginal,
		Resource: &fyne.StaticResource{
			StaticName:    "eeprom.png",
			StaticContent: eepromBytes},
	}
	OVERVIEW = &canvas.Image{
		ScaleMode: canvas.ImageScaleSmooth,
		FillMode:  canvas.ImageFillOriginal,
		Resource: &fyne.StaticResource{
			StaticName:    "overview.png",
			StaticContent: overviewBytes},
	}
)
