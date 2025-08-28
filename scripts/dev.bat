@echo off
REM Windows dev script for LyricsCraw
set APP_ENV=development

if "%APP_ENV%"=="development" (
    echo Starting Application server in development mode...
) else (
    echo Starting Application server in production mode...
)

REM Execute the development server
go run src/main.go
if %errorlevel%==0 (
    echo Application server started successfully.
) else (
    echo Error starting Application server.
    exit /b 1
)
