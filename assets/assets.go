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

func PCB() fyne.CanvasObject {
	return &canvas.Image{
		ScaleMode: canvas.ImageScalePixels,
		FillMode:  canvas.ImageFillOriginal,
		Resource: &fyne.StaticResource{
			StaticName:    "pcb.jpg",
			StaticContent: pcbBytes},
	}
}

func EEPROM() fyne.CanvasObject {
	return &canvas.Image{
		ScaleMode: canvas.ImageScalePixels,
		FillMode:  canvas.ImageFillOriginal,
		Resource: &fyne.StaticResource{
			StaticName:    "eeprom.jpg",
			StaticContent: eepromBytes},
	}
}

func OVERVIEW() fyne.CanvasObject {
	return &canvas.Image{
		ScaleMode: canvas.ImageScalePixels,
		FillMode:  canvas.ImageFillOriginal,
		Resource: &fyne.StaticResource{
			StaticName:    "overview.jpg",
			StaticContent: overviewBytes},
	}
}
