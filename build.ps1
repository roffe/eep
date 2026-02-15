$env:GOARCH = "amd64"
fyne package --target windows --icon Icon.png --release --executable Saab_CIM_Tool_x64.exe --name Saab_CIM_Tool_x64
Start-Process -FilePath $winRarPath -ArgumentList $winRarArgs -NoNewWindow -Wait
Remove-Item .\avr\firmware.hex
Copy-Item firmware/build/firmware.ino.hex avr/firmware.hex
