.PHONY: gui
release:
	goreleaser release --rm-dist
test:
	go test -v ./...

gui:
	rm -f avr/firmware.hex
	cp firmware/build/firmware.ino.hex avr/firmware.hex
	GOARCH=amd64 fyne package --target windows --icon Icon.png --release --executable Saab_CIM_Tool_x64.exe --name Saab_CIM_Tool_x64
	"/c/Program Files/WinRAR/WinRAR.exe" a -afzip Saab_CIM_Tool_x64.zip Saab_CIM_Tool_x64.exe
	GOARCH=386 fyne package --target windows --icon Icon.png --release --executable Saab_CIM_Tool_x86.exe --name Saab_CIM_Tool_x86
	"/c/Program Files/WinRAR/WinRAR.exe" a -afzip Saab_CIM_Tool_x86.zip Saab_CIM_Tool_x86.exe

# --appVersion $$(git describe --tags `git rev-list --tags --max-count=1`|cut -d"v" -f2)


