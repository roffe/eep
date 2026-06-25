.PHONY: gui firmware

release:
	goreleaser release --rm-dist

test:
	go test -v ./...

firmware/build/firmware.ino.hex: firmware/firmware.ino
	arduino-cli compile -e -b arduino:avr:nano --output-dir ./firmware/build/ ./firmware/firmware.ino
	rm -f avr/firmware.hex
	cp firmware/build/firmware.ino.hex avr/firmware.hex

firmware: firmware/build/firmware.ino.hex

gui: firmware/build/firmware.ino.hex
	rm -f avr/firmware.hex
	cp firmware/build/firmware.ino.hex avr/firmware.hex
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOARCH=amd64 fyne package --target windows --release --executable Saab_CIM_Tool_x64.exe --name Saab_CIM_Tool_x64
	zip Saab_CIM_Tool_x64.zip Saab_CIM_Tool_x64.exe
	CGO_ENABLED=1 CC=i686-w64-mingw32-gcc GOARCH=386 fyne package --target windows --release --executable Saab_CIM_Tool_x86.exe --name Saab_CIM_Tool_x86
	zip Saab_CIM_Tool_x86.zip Saab_CIM_Tool_x86.exe

# --appVersion $$(git describe --tags `git rev-list --tags --max-count=1`|cut -d"v" -f2)


