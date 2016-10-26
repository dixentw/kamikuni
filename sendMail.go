package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// Send mail to clients
func Send(records []Record) {
	ctx := context.Background()
	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("cannot read client secret %v", err)
	}
	conf, err := google.ConfigFromJSON(b, gmail.MailGoogleComScope)
	if err != nil {
		log.Fatalf("cannot parse secret file, %v", err)
	}
	client := getClient(ctx, conf)
	srv, _ := gmail.New(client)

	var msg gmail.Message
	temp := []byte("From: 'me'\r\n" +
		"reply-to: dixentw@gmail.com\r\n" +
		"To:  dixentw@icloud.com\r\n" +
		"Subject: Kamikuni Daily Update \r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"\r\n" + formatRecord(records))

	msg.Raw = base64.StdEncoding.EncodeToString(temp)
	msg.Raw = strings.Replace(msg.Raw, "/", "_", -1)
	msg.Raw = strings.Replace(msg.Raw, "+", "-", -1)
	msg.Raw = strings.Replace(msg.Raw, "=", "", -1)

	_, e := srv.Users.Messages.Send("me", &msg).Do()
	if e != nil {
		log.Fatalf("unable to send %v", e)
	}

}

func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

func formatRecord(records []Record) string {
	result := "<ul>"
	for _, rec := range records {
		result += "<li><b>Title : </b><h3>" + rec.title + "</h3>"
		result += `<li><b>Cover</b><img src="` + rec.cover + `" height="250" width="250"/>`
		result += "<li><b>Info : </b>" + rec.info
		result += "<li><b>Introduction: </b>" + rec.intro
		links := strings.Split(rec.downloadLinks, ",")
		for _, l := range links {
			result += "<li><b>Links</b>" + l
		}
		result += "<hr>"
	}
	result += "</ul>"
	return result
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("gmail-go-quickstart.json")), err
}
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
