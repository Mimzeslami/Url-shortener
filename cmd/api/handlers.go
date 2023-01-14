package main

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

func (app *Config) StoreTinyURL(longUrl string, shortURL string) {
	d := app.Models.Url
	d.LongUrl = longUrl
	d.ShortUrl = shortURL
	app.Models.Url.Insert(d)
	app.RedisClient.HSet("urls", shortURL, longUrl)
}

func (app *Config) GenerateHashAndInsert(longUrl string, startIndex int) string {

	byteURLData := []byte(longUrl)
	hashedURLData := fmt.Sprintf("%x", md5.Sum(byteURLData))
	tinyURLRegex, err := regexp.Compile("[/+]")
	if err != nil {
		return "Unable to generate short URL"
	}

	tinyURLData := tinyURLRegex.ReplaceAllString(base64.URLEncoding.EncodeToString([]byte(hashedURLData)), "_")
	if len(tinyURLData) < (startIndex + 6) {
		return "Unable to generate short URL"
	}

	shortUrl := tinyURLData[startIndex : startIndex+6]

	dbURLData, err := app.Models.Url.GetByShortUrl(shortUrl)

	if err != nil {
		fmt.Println(dbURLData, "in not found")
		go app.StoreTinyURL(longUrl, shortUrl)
		return shortUrl
	} else if (dbURLData.ShortUrl == shortUrl) && (dbURLData.LongUrl == longUrl) {
		fmt.Println(dbURLData, "in found and equal")
		return shortUrl
	} else {
		return app.GenerateHashAndInsert(longUrl, startIndex+1)
	}
}

func (app *Config) GetTinyHandler(res http.ResponseWriter, req *http.Request) {
	requestParams, err := req.URL.Query()["longUrl"]
	if !err || len(requestParams[0]) < 1 {
		app.errorJSON(res, errors.New("URL parameter longUrl is missing"), http.StatusBadRequest)
		return

	} else {
		longUrl := requestParams[0]
		shortUrl := app.GenerateHashAndInsert(longUrl, 0)
		app.writeJSON(res, http.StatusAccepted, shortUrl)
	}
}

func (app *Config) GetLongHandler(res http.ResponseWriter, req *http.Request) {

	requestParams, err := req.URL.Query()["shortUrl"]

	if !err || len(requestParams[0]) < 1 {
		app.errorJSON(res, errors.New("URL parameter shortUrl is missing"), http.StatusBadRequest)
		return

	}

	shortUrl := requestParams[0]

	redisSearchResult := app.RedisClient.HGet("urls", shortUrl)

	if redisSearchResult.Val() != "" {
		app.writeJSON(res, http.StatusAccepted, redisSearchResult.Val())

	} else {
		url, err := app.Models.Url.GetByShortUrl(shortUrl)

		if err != nil {
			app.errorJSON(res, err)
			return
		}

		if url.LongUrl != "" {
			app.RedisClient.HSet("urls", shortUrl, url.LongUrl)
			app.writeJSON(res, http.StatusAccepted, url.LongUrl)
		} else {
			app.errorJSON(res, err)
			return
		}
	}
}
