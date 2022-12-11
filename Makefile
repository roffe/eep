.PHONY: gui
release:
	goreleaser release --rm-dist
test:

512:
	go run . write --chip 66 --size 512 --org 8 1001-01.bin --erase
	go run . read --chip 66 --size 512 --org 8 1001-01_read.bin

256:
	go run . write --chip 66 --size 256 --org 16 256.bin --erase
	go run . read --chip 66 --size 256 --org 16 256_read.bin

gui:
	rm -f avr/firmware.hex
	cp firmware/build/firmware.ino.hex avr/firmware.hex
	fyne package --target windows --icon Icon.png --release
# --appVersion $$(git describe --tags `git rev-list --tags --max-count=1`|cut -d"v" -f2)


