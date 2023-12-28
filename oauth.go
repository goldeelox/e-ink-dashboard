package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

func NewOauth2Client() *http.Client {
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		slog.Error("Unable to read client secret file", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		slog.Error("Unable to parse client secret file to config", err)
	}
	return getClient(config)
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := TOKEN_FILE
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		getTokenFromWeb(config)
	}
	return config.Client(context.Background(), tok)
}

func callbackServer(config *oauth2.Config) {
	mux := http.NewServeMux()
	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authCode := r.FormValue("code")
		if authCode == "" {
      slog.Warn("got request on callback listener that i could not understand",
        "uri", r.RequestURI)
			w.WriteHeader(http.StatusBadRequest)
      w.Write([]byte("No code provided"))
			return
		}

		slog.Info("got new auth code")
		tok, err := config.Exchange(context.TODO(), authCode)
		if err != nil {
			slog.Error("Unable to retrieve token from web", "error", err)
		}
		saveToken("token.json", tok)
		w.Write([]byte("OK"))
		go srv.Shutdown(context.Background())
		slog.Info("shutting down webserver")
	})
	slog.Info("listening on :8080")
	srv.ListenAndServe()
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	authMsg := fmt.Sprintf("Go to the following link in your browser %v", authURL)
	slog.Info(authMsg)
	callbackServer(config)
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	slog.Info("Saving credential file", "path", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		slog.Error("Unable to cache oauth token", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
