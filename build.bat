go build -v -x -buildmode=exe -o %cd%\output\main.exe -i main.go
echo F|xcopy.exe "%cd%\rp_main.exe" "%cd%\output\rp_main.exe" /c /I /y
echo F|xcopy.exe "%cd%\7z.exe" "%cd%\output\7z.exe" /c /I /y
echo F|xcopy.exe "%cd%\7z.dll" "%cd%\output\7z.dll" /c /I /y
echo F|xcopy.exe "%cd%\profile.json" "%cd%\output\profile.json" /c /I /y
7z.exe a "%cd%\output.zip" "%cd%\output"