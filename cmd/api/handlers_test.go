package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"url/data"
)

func newApp() *Config {
	conn := connectToDB()
	redisClient := connectToRedis()
	return &Config{
		DB:          conn,
		RedisClient: redisClient,
		Models:      data.New(conn),
	}
}

func TestSaveShortURL(t *testing.T) {
	// create a mock sql.DB and redis.Client
	app := newApp()

	// test for valid input
	longUrl := "https://www.test-url.com"
	shortURL := "abcdef"
	app.SaveShortURL(longUrl, shortURL)

}

func TestGetShortHandler(t *testing.T) {
	// Set up test server and client
	app := newApp()

	ts := httptest.NewServer(http.HandlerFunc(app.GetShortHandler))
	defer ts.Close()
	client := ts.Client()

	// Test case with valid long URL
	req, _ := http.NewRequest("GET", ts.URL+"?longUrl=https://www.test-url.com", nil)
	res, _ := client.Do(req)
	if res.StatusCode != http.StatusAccepted {
		t.Errorf("Expected status code %d, got %d", http.StatusAccepted, res.StatusCode)
	}

	// Test case with missing long URL parameter
	req, _ = http.NewRequest("GET", ts.URL, nil)
	res, _ = client.Do(req)
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
}
