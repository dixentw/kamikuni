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
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"

	"github.com/PuerkitoBio/goquery"
	"github.com/robfig/cron"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Record : struct to store a record
type Record struct {
	id            string
	title         string
	cover         string
	downloadLinks string
	intro         string
	info          string
}

var (
	latestID = ""
	tmpID    = ""
	batch    []Record
)

func workHorse() {
	url := "http://aichun.co"
	counter := 1
	isHit := false
	for {
		if counter > 1 {
			url += "/page/" + strconv.Itoa(counter)
		}
		doc, _ := goquery.NewDocument(url)
		var records []Record
		doc.Find("#table td").Each(func(_ int, s *goquery.Selection) {
			mp3id, _ := s.Find(".box").Attr("id")
			c, _ := s.Find(".box").Attr("data-cover")
			l, _ := s.Find(".box").Attr("data-links")
			d, _ := s.Find(".box").Attr("data-info")
			r := Record{
				id:            mp3id,
				title:         s.Find(".full_title").Text(),
				cover:         c,
				downloadLinks: l,
				intro:         s.Find(".text").Text(),
				info:          d,
			}
			records = append(records, r)
		})
		//compare to previous id
		for _, rec := range records {
			if rec.id == latestID {
				isHit = true
				break
			}
			batch = append(batch, rec)
		}
		//save latest mp3id
		if counter == 1 {
			tmpID = records[0].id
		}
		//send to slack
		counter++
		if counter > 3 || isHit {
			send(batch)
			/**
			for _, rec := range records {
				log.Printf("%v\n", rec)
				log.Println("------------------------")
			}*/
			batch = batch[:0]
			latestID = tmpID
			break
		}
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
		result += "<li><b>Title : </b>" + rec.title
		result += `<li><b>Cover</b><img src="` + rec.cover + `" height="200" width="200"/>`
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

func send(records []Record) {
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

func main() {
	fmt.Println("======program start================")
	c := cron.New()
	c.AddFunc("0 0 22 * * *", func() {
		fmt.Println("===========    start working   ===============")
		workHorse()
		fmt.Println("===========    end agent   ===============")
		fmt.Println(latestID)
	})
	c.Start()
	defer c.Stop()
	for {
		time.Sleep(1 * time.Second)
	}
}
