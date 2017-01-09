package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/robfig/cron"
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
	prevRecords map[string]Record
)

func workHorse() {
	url := "http://aichun.co"
	counter := 1
	records := make(map[string]Record)
	for counter < 5 {
		if counter > 1 {
			url += "/page/" + strconv.Itoa(counter)
		}
		doc, _ := goquery.NewDocument(url)
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
			records[r.id] = r
		})
		counter++
	}
	//compare to previous ids
	var batch []Record
	for id, rec := range records {
		fmt.Printf("every_compare_id: %v", id)
		_, ok := prevRecords[id]
		fmt.Printf("fetched result: %v", ok)
		if !ok {
			batch = append(batch, rec)
		}
	}
	//fmt.Printf("prev: %v\n", prevRecords)
	//fmt.Printf("curr: %v\n", records)
	//fmt.Printf("diff: %v\n", batch)
	if len(batch) > 0 {
		Send(batch)
	}
	prevRecords = records
}

func main() {
	fmt.Println("======program start================")
	fmt.Println("======program start================")
	c := cron.New()
	c.AddFunc("0 0 * * * *", func() {
		fmt.Println("===========    start working   ===============")
		workHorse()
		fmt.Println("===========    end agent   ===============")
	})
	c.Start()
	defer c.Stop()
	for {
		time.Sleep(1 * time.Second)
	}
}
