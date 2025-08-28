package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tutitoos/lyricscrawl/src/api/controller"
)

func NewRouter() *gin.Engine {

	router := gin.New()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	lyricsController := controller.LyricsController{}
	v1 := router.Group("/v1")
	{
		v1.GET("/lyrics", lyricsController.GetLyrics)
		v1.GET("/lyrics/all", lyricsController.GetAllLyrics)
	}

	return router
}
