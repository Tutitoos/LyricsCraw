@echo off
REM Windows build script for LyricsCraw

if "%1"=="linux" (
    echo Building for Linux...
    set GOOS=linux
    set CGO_ENABLED=0
    if not exist bin mkdir bin
    REM This only works if you have cross-compiling configured in Go on Windows
    call go build -ldflags="-w -s" -a -installsuffix cgo -o bin\app src/main.go
) else (
    echo Building for current OS...
    set CGO_ENABLED=0
    if not exist bin mkdir bin
    call go build -ldflags="-w -s" -a -installsuffix cgo -o bin\app.exe src/main.go
)
if %errorlevel% neq 0 (
    echo Error during build.
    exit /b 1
)
echo Build completed successfully.
