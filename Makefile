.PHONY: gui
release:
	goreleaser release --rm-dist
test:
	go test -v ./...

gui:
	rm -f avr/firmware.hex
	cp firmware/build/firmware.ino.hex avr/firmware.hex
	fyne package --target windows --icon Icon.png --release
# --appVersion $$(git describe --tags `git rev-list --tags --max-count=1`|cut -d"v" -f2)


