package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tutitoos/lyricscrawl/src/cache"
	"github.com/tutitoos/lyricscrawl/src/logger"
	"github.com/tutitoos/lyricscrawl/src/scraper"
)

type LyricsController struct {
}

func (tc *LyricsController) GetLyrics(c *gin.Context) {

	query := c.Query("query")
	if query == "" {
		err := fmt.Errorf("query parameter is required")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Try cache first
	if v, ok := cache.Default.Get(query); ok {
		c.JSON(http.StatusOK, gin.H{"data": v, "cached": true})
		return
	}

	lyrics, err := scraper.ScrapeVagalume(query)
	if err != nil {
		fmt.Println("Failed to get token:", err)
		logger.Sugar.Info("Failed to get token", "error", err)
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Store in cache
	cache.Default.Set(query, lyrics)

	c.JSON(http.StatusOK, gin.H{"data": lyrics, "cached": false})
}
