package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/tutitoos/lyricscrawl/src/api/router"
	"github.com/tutitoos/lyricscrawl/src/cache"
	"github.com/tutitoos/lyricscrawl/src/logger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  No se pudo cargar el archivo .env: %v", err)
	}

	logger.InitLogger()

	// Init in-memory lyrics cache
	cache.InitFromEnv()

	env := os.Getenv("APP_ENV")
	switch env {
	case "development":
		gin.SetMode(gin.DebugMode)
	case "production":
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.DebugMode)
		logger.Sugar.Warnf("⚠️  Modo de entorno desconocido: %s, usando modo de desarrollo por defecto", env)
	}

	// Inicializar el router
	r := router.NewRouter()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
		logger.Sugar.Warn("⚠️  APP_PORT no está definido, usando puerto por defecto :8080")
	}

	logger.Sugar.Infof("🚀 Iniciando servidor en %s (env: %s)", port, env)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		logger.Sugar.Fatalf("❌ Error al iniciar el servidor: %v", err)
	}
}
