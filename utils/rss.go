package utils

import (
	"encoding/xml"
	"html/template"
	"net/http"
)

type Item struct {
	Title     string    `xml:"title"`
	Link      string    `xml:"link"`
	Desc      string    `xml:"description"`
	Guid      string    `xml:"guid"`
	Content   template.HTML `xml:"encoded"`
	PubDate   string    `xml:"pubDate"`
	Comments	string		`xml:"comments"`
}

type Channel struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
	Desc  string `xml:"description"`
	Items []Item `xml:"item"`
}

type RSS struct {
	Channel Channel `xml:"channel"`
}

func RSSUrlToStruct(url string, payload *RSS) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	decoder := xml.NewDecoder(resp.Body)
	err = decoder.Decode(&payload)
	if err != nil {
		return err
	}
	return nil
}
