# Makefile

ifeq ($(OS),Windows_NT)
	SHELL := cmd.exe
	.SHELLFLAGS := /C
	DEV_CMD := scripts\dev.bat
	BUILD_CMD := scripts\build.bat
	START_CMD := scripts\start.bat
	MKDIR_CMD := if not exist bin mkdir bin
else
	DEV_CMD := sh scripts/dev.sh
	BUILD_CMD := sh scripts/build.sh
	START_CMD := sh scripts/start.sh
	MKDIR_CMD := mkdir -p bin
endif

build:
	@echo "[BUILD] Creating bin directory..."
	$(MKDIR_CMD)
	@echo "[BUILD] Building..."
	$(BUILD_CMD)
	@echo "[BUILD] Build complete."

dev:
	$(DEV_CMD)

start:
	$(START_CMD)
