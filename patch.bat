@echo off
setlocal

set "OLD_EXE=fuge.exe"

start "" /wait "%OLD_EXE% patch ggez"

move /y "downloads/fuge.exe" "fuge.exe"
move /y "downloads/config.ini" "config.ini"
move /y "downloads/fuge.jpg" "fuge.jpg"
move /y "downloads/patch.bat" "patch.bat"

:: Confirm success
if exist "%OLD_EXE%" (
    echo Patch was successful!
) else (
    echo Patch failed.. We'll get him next time
    exit /b 1
)

exit /b 0
