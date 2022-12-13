package gui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewCimInfoWidget() fyne.Widget {
	return &cimInfoWidget{}
}

type cimInfoWidget struct {
	widget.BaseWidget
	Text string
}

func (c *cimInfoWidget) CreateRenderer() fyne.WidgetRenderer {
	cr := &cimInfoWidgetRenderer{
		background: canvas.NewRectangle(theme.MenuBackgroundColor()),
		objects: []fyne.CanvasObject{
			widget.NewLabel("CIM Info"),
		},
	}
	c.ExtendBaseWidget(c)
	return cr
}

// MouseIn is called when a desktop pointer enters the widget
func (c *cimInfoWidget) MouseIn(*desktop.MouseEvent) {
	log.Println("mouse in")
}

// MouseMoved is called when a desktop pointer hovers over the widget
func (c *cimInfoWidget) MouseMoved(*desktop.MouseEvent) {
}

// MouseOut is called when a desktop pointer exits the widget
func (c *cimInfoWidget) MouseOut() {
	log.Println("mouse out")
}

// =================================================================================================

type cimInfoWidgetRenderer struct {
	background *canvas.Rectangle
	objects    []fyne.CanvasObject
}

func (c *cimInfoWidgetRenderer) Layout(size fyne.Size) {
	c.background.Resize(size)
}

// Implements: fyne.WidgetRenderer
func (c *cimInfoWidgetRenderer) MinSize() (size fyne.Size) {
	for _, obj := range c.objects {
		size = size.Max(obj.MinSize())
	}
	return
}

func (c *cimInfoWidgetRenderer) Refresh() {
	log.Println("refresh")
}

// Implements: fyne.WidgetRenderer
func (c *cimInfoWidgetRenderer) Destroy() {
}

// Objects returns the objects that should be rendered.
//
// Implements: fyne.WidgetRenderer
func (c *cimInfoWidgetRenderer) Objects() []fyne.CanvasObject {
	return c.objects
}

// SetObjects updates the objects of the renderer.
func (c *cimInfoWidgetRenderer) SetObjects(objects []fyne.CanvasObject) {
	c.objects = objects
}
