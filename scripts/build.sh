#!/bin/bash

# If 'linux' is passed as an argument, build for Linux; otherwise, build for the current OS
if [ "$1" = "linux" ]; then
  echo "Building for Linux..."
  export GOOS=linux
else
  echo "Building for current OS..."
  unset GOOS
fi

cd src || { echo "Error changing to 'src' directory."; exit 1; }
CGO_ENABLED=0 go build -ldflags="-w -s" -a -installsuffix cgo -o ../bin/app || { echo "Error during build."; exit 1; }
cd ../ || { echo "Error returning to the previous directory."; exit 1; }
echo "Build completed successfully."
