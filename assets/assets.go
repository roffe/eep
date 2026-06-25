package assets

import (
	_ "embed"
)

//go:embed pcb.jpg
var PcbBytes []byte

//go:embed eeprom.jpg
var EepromBytes []byte

//go:embed overview.jpg
var OverviewBytes []byte

//go:embed logo.png
var LogoBytes []byte
