package main

import (
	"encoding/xml"
	"time"
)

const (
	pubDateFormat = "Mon, 2 Jan 2006 15:04:05 -0700"
)

type PodcastRss struct {
	XMLName     xml.Name `xml:"rss"`
	XmlnsItunes string   `xml:"xmlns:itunes,attr,omitempty"`
	XmlnsAtom   string   `xml:"xmlns:atom,attr,omitempty"`
	Version     string   `xml:"version,attr,omitempty"`
	Channel     PodcastChannel
}

type PodcastChannel struct {
	XMLName xml.Name `xml:"channel"`
	Title   string   `xml:"title,omitempty"`
	Link    string   `xml:"link,omitempty"`
	Image   struct {
		URL   string `xml:"url,omitempty"`
		Title string `xml:"title,omitempty"`
		Link  string `xml:"link,omitempty"`
	} `xml:"image,omitempty"`
	Description string `xml:"description,omitempty"`
	Language    string `xml:"language,omitempty"`
	Copyright   string `xml:"copyright,omitempty"`
	AtomLink    struct {
		Href string `xml:"href,attr,omitempty"`
		Rel  string `xml:"rel,attr,omitempty"`
		Type string `xml:"type,attr,omitempty"`
	} `xml:"atom:link,omitempty"`
	LastBuildDate  PubDate `xml:"lastBuildDate,omitempty"`
	ITunesAuthor   string  `xml:"itunes:author,omitempty"`
	ITunesSummary  string  `xml:"itunes:summary,omitempty"`
	ITunesSubtitle string  `xml:"itunes:subtitle,omitempty"`
	ITunesOwner    struct {
		ITunesName  string `xml:"itunes:name,omitempty"`
		ITunesEmail string `xml:"itunes:email,omitempty"`
	} `xml:"itunes:owner,omitempty"`
	ITunesExplicit string `xml:"itunes:explicit,omitempty"`
	ITunesKeywords string `xml:"itunes:keywords,omitempty"`
	ITunesImage    struct {
		Href string `xml:"href,attr,omitempty"`
	} `xml:"itunes:image,omitempty"`
	ITunesCategory struct {
		Text string `xml:"text,attr,omitempty"`
	} `xml:"itunes:category,omitempty"`
	PubDate PubDate `xml:"pubDate,omitempty"`
	Items   PodcastItems
}

type PodcastItem struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title,omitempty"`
	Link        string   `xml:"link,omitempty"`
	PubDate     PubDate  `xml:"pubDate,omitempty"`
	Description string   `xml:"description,omitempty"`
	Enclosure   struct {
		URL    string `xml:"url,attr,omitempty"`
		Length int    `xml:"length,attr,omitempty"`
		Type   string `xml:"type,attr,omitempty"`
	} `xml:"enclosure,omitempty"`
	GUID           string `xml:"guid,omitempty"`
	ITunesAuthor   string `xml:"itunes:author,omitempty"`
	ITunesSubtitle string `xml:"itunes:subtitle,omitempty"`
	ITunesDuration string `xml:"itunes:duration,omitempty"`
	ITunesSummary  string `xml:"itunes:summary,omitempty"`
	ITunesImage    struct {
		Href string `xml:"href,attr,omitempty"`
	} `xml:"itunes:image,omitempty"`
}

type PodcastItems []PodcastItem

func (p PodcastItems) Len() int {
	return len(p)
}

func (p PodcastItems) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p PodcastItems) Less(i, j int) bool {
	return p[i].PubDate.Unix() <= p[j].PubDate.Unix()
}

type PubDate struct {
	time.Time
}

func (p PubDate) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	e.EncodeToken(start)
	e.EncodeToken(xml.CharData(p.Format(pubDateFormat)))
	e.EncodeToken(xml.EndElement{Name: start.Name})
	return nil
}

func NewPodcastRss() *PodcastRss {
	return &PodcastRss{
		XmlnsItunes: "http://www.itunes.com/dtds/podcast-1.0.dtd",
		XmlnsAtom:   "http://www.w3.org/2005/Atom",
		Version:     "2.0",
	}
}
