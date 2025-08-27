#!/bin/bash

export APP_ENV=production

if [ "$APP_ENV" = "development" ]; then
  echo "Starting Application server in development mode..."
else
  echo "Starting Application server in production mode..."
fi
# Execute the built application
./bin/app
if [ $? -eq 0 ]; then
  echo "Application started successfully."
else
  echo "Error starting the application."
  exit 1
fi
