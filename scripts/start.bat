@echo off
REM Windows start script for SpotyCraw
set APP_ENV=production

if "%APP_ENV%"=="development" (
    echo Starting Application server in development mode...
) else (
    echo Starting Application server in production mode...
)

REM Execute the built application
start bin\app.exe
if %errorlevel%==0 (
    echo Application started successfully.
) else (
    echo Error starting the application.
    exit /b 1
)
