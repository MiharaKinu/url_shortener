package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

type Config struct {
	AllowDomain []string `yaml:"allowDomain"`
	Expire      int      `yaml:"expire"`
	Host        string   `yaml:"host"`
	ShortLength int      `yaml:"shortLength"`
	Port        int      `yaml:"port"`
}

var (
	config Config
	urlDB  *URLDatabase
)

func initConfig() {
	configFile := filepath.Join(".", "config.yaml")
	file, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}
}

func generateShortID() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	id := make([]rune, config.ShortLength)
	for i := range id {
		id[i] = letters[rand.Intn(len(letters))]
	}
	return string(id)
}

func isValidDomain(longURL string) bool {
	parsedURL, err := url.Parse(longURL)
	if err != nil {
		return false
	}

	host := parsedURL.Hostname()
	for _, domain := range config.AllowDomain {
		if domain == host {
			return true
		}
	}
	return false
}

func shortenURL(c *gin.Context) {
	var request struct {
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "-3", "msg": "Invalid request."})
		return
	}

	longURL := request.URL
	if !isValidDomain(longURL) {
		c.JSON(http.StatusOK, gin.H{"code": "-1", "msg": "The URL to be processed is not included in allowDomain."})
		return
	}

	// Check if the long URL already exists in the database
	mapping, err := urlDB.GetURLByLongURL(longURL)
	if err == nil {
		// URL exists, extend its expiration
		expiresAt := time.Now().Unix() + int64(config.Expire)
		if err := urlDB.ExtendURLExpiration(mapping.ID, expiresAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": "-4", "msg": "Failed to extend URL expiration."})
			return
		}
		shortURL := fmt.Sprintf("%s/%s", config.Host, mapping.ID)
		c.JSON(http.StatusOK, gin.H{"code": "200", "data": gin.H{"url": shortURL}})
		return
	} else if err != sql.ErrNoRows {
		// An unexpected error occurred
		c.JSON(http.StatusInternalServerError, gin.H{"code": "-5", "msg": "Failed to query database."})
		return
	}

	// URL does not exist, create a new record
	shortID := generateShortID()
	expiresAt := time.Now().Unix() + int64(config.Expire)
	if err := urlDB.CreateShortURL(shortID, longURL, expiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "-6", "msg": "Failed to shorten URL."})
		return
	}

	shortURL := fmt.Sprintf("%s/%s", config.Host, shortID)
	c.JSON(http.StatusOK, gin.H{"code": "200", "data": gin.H{"url": shortURL}})
}

func decodeURL(c *gin.Context) {
	var request struct {
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "-3", "msg": "Invalid request."})
		return
	}

	shortID := filepath.Base(request.URL)

	mapping, err := urlDB.GetURLByShortID(shortID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, gin.H{"code": "-2", "msg": "404 Not Found."})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "-7", "msg": "Failed to decode URL."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "200", "data": gin.H{"url": mapping.LongURL}})
}

func redirectURL(c *gin.Context) {
	shortID := c.Param("shortID")
	mapping, err := urlDB.GetURLByShortID(shortID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, gin.H{"code": "-2", "msg": "404 Not Found."})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "-8", "msg": "Failed to redirect URL."})
		return
	}
	parsedURL, err := url.Parse(mapping.LongURL)
	if err != nil || !parsedURL.IsAbs() {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "-9", "msg": "Invalid long URL."})
		return
	}
	c.Redirect(http.StatusFound, mapping.LongURL)
}

func main() {
	initConfig()
	urlDB = NewURLDatabase()
	defer urlDB.Close()
	go urlDB.CleanExpiredRecords()

	r := gin.Default()
	err := r.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	if err != nil {
		panic(err)
	}

	r.POST("/short", shortenURL)
	r.POST("/decode", decodeURL)
	r.GET("/:shortID", redirectURL)

	log.Fatal(r.Run(fmt.Sprintf(":%d", config.Port)))
}
