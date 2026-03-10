@echo off
setlocal enabledelayedexpansion

set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%"

if exist "version.env" (
    for /f "tokens=1,* delims==" %%a in ('findstr /i "VERSION" version.env') do (
        set "VERSION=%%b"
    )
    if "!VERSION!"=="" set "VERSION=0.0.1"
) else (
    set "VERSION=0.0.1"
)

set "APP_NAME=avproc"
set "ARCH=amd64"
set "OUTPUT_DIR=dist"
set "FFMPEG_SRC=C:\ffmpeg\ffmpeg-2026-02-26-git-6695528af6-essentials_build\bin"
set "MINGW_DIR=C:\llvm-mingw-20260224-ucrt-x86_64"

if "%~1"=="" goto build
if "%~1"=="build" goto build
if "%~1"=="clean" goto clean
echo Usage: %0 {build^|clean}
exit /b 1

:build
echo === Building avproc %VERSION% ===

set "PATH=%MINGW_DIR%\bin;%PATH%"
set CGO_ENABLED=1

set "OUTPUT_SUBDIR=%OUTPUT_DIR%\%APP_NAME%-%VERSION%-win-%ARCH%"
if not exist "%OUTPUT_SUBDIR%" mkdir "%OUTPUT_SUBDIR%"

go build -ldflags "-s -w -H=windowsgui" -o "%OUTPUT_SUBDIR%\%APP_NAME%.exe" .

if errorlevel 1 (
    echo Build failed!
    exit /b 1
)

echo === Copying FFmpeg ===
copy /Y "%FFMPEG_SRC%\ffmpeg.exe" "%OUTPUT_SUBDIR%\" >nul
copy /Y "%FFMPEG_SRC%\ffplay.exe" "%OUTPUT_SUBDIR%\" >nul
copy /Y "%FFMPEG_SRC%\ffprobe.exe" "%OUTPUT_SUBDIR%\" >nul

echo === Creating zip ===
cd "%OUTPUT_SUBDIR%"
powershell -Command "Compress-Archive -Path * -DestinationPath ..\\%APP_NAME%_%VERSION%_win-%ARCH%.zip -Force"
cd /d "%SCRIPT_DIR%"

echo === Build complete ===
echo Output: %OUTPUT_DIR%\%APP_NAME%_%VERSION%_win-%ARCH%.zip
dir /b "%OUTPUT_SUBDIR%"

exit /b 0

:clean
if exist "%OUTPUT_DIR%" (
    rmdir /s /q "%OUTPUT_DIR%"
    echo Cleaned %OUTPUT_DIR%
) else (
    echo Nothing to clean
)
exit /b 0
