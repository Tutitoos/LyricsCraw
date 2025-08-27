#!/bin/bash

export APP_ENV=development

if [ "$APP_ENV" = "development" ]; then
  echo "Starting Application server in development mode..."
else
  echo "Starting Application server in production mode..."
fi
# Execute the development server
go run src/main.go
if [ $? -eq 0 ]; then
  echo "Application server started successfully."
else
  echo "Error starting Application server."
  exit 1
fi
