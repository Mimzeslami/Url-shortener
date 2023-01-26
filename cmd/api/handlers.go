package main

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

func (app *Config) SaveShortURL(longUrl string, shortURL string) error {
	d := app.Models.Url
	d.LongUrl = longUrl
	d.ShortUrl = shortURL
	_, err := app.Models.Url.Insert(d)
	if err != nil {
		return err
	}
	_, err = app.RedisClient.HSet("urls", shortURL, longUrl).Result()
	if err != nil {
		return err
	}
	return nil
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
		go app.SaveShortURL(longUrl, shortUrl)
		return shortUrl
	} else if (dbURLData.ShortUrl == shortUrl) && (dbURLData.LongUrl == longUrl) {
		return shortUrl
	} else {
		return app.GenerateHashAndInsert(longUrl, startIndex+1)
	}
}

func (app *Config) GetShortHandler(res http.ResponseWriter, req *http.Request) {
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
