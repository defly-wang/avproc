@echo off
setlocal

set "DIST_DIR=dist"
set "FFMPEG_SRC=C:\ffmpeg\ffmpeg-2026-02-26-git-6695528af6-essentials_build\bin"
set "MINGW_DIR=C:\llvm-mingw-20260224-ucrt-x86_64"

echo === Building avproc ===

set "PATH=%MINGW_DIR%\bin;%PATH%"
set CGO_ENABLED=1

go build -ldflags="-s -w" -o "%DIST_DIR%\avproc.exe"

if errorlevel 1 (
    echo Build failed!
    exit /b 1
)

echo === Copying FFmpeg ===
if not exist "%DIST_DIR%" mkdir "%DIST_DIR%"
copy /Y "%FFMPEG_SRC%\ffmpeg.exe" "%DIST_DIR%\"
copy /Y "%FFMPEG_SRC%\ffplay.exe" "%DIST_DIR%\"
copy /Y "%FFMPEG_SRC%\ffprobe.exe" "%DIST_DIR%\"

echo === Build complete ===
echo Output: %DIST_DIR%
dir /b "%DIST_DIR%"

pause
