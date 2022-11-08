package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/gorilla/mux"
)

type Server struct {
	Output string
	Title  string
	Addr   string
}

func (s *Server) errorHandler(f func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			s.Log(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func (s *Server) Run() error {

	s.Log("start ", s.Addr)

	router := mux.NewRouter()

	router.HandleFunc("/podcast/{program}.m4a", s.errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		dir := mux.Vars(r)["program"]

		medPath, medStat, err := s.medPath(dir)

		if _, err := os.Stat(medPath); err != nil {
			http.NotFound(w, r)
			return nil
		}

		xmlPath, _, err := s.xmlPath(dir)

		if _, err := os.Stat(xmlPath); err != nil {
			http.NotFound(w, r)
			return nil
		}

		f, err := os.Open(medPath)

		if err != nil {
			return err
		}

		defer f.Close()

		http.ServeContent(w, r, medStat.Name(), medStat.ModTime(), f)
		return nil
	}))

	router.HandleFunc("/rss", s.errorHandler(func(w http.ResponseWriter, r *http.Request) error {

		baseURL, err := url.Parse("http://" + r.Host)

		if err != nil {
			return err
		}

		rss, err := s.rss(baseURL)

		if err != nil {
			return err
		}

		var b bytes.Buffer

		b.WriteString(xml.Header)

		enc := xml.NewEncoder(&b)
		enc.Indent("", "    ")
		if err := enc.Encode(rss); err != nil {
			return err
		}

		if _, err := io.Copy(w, &b); err != nil {
			return err
		}

		return nil
	}))

	router.HandleFunc("/podcast/{program}.{ext:png|jpg|jpeg|gif|bmp}", s.errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		dir := mux.Vars(r)["program"]
		ext := mux.Vars(r)["ext"]

		imgPath, _, err := s.imgPath(dir, ext)

		if _, err := os.Stat(imgPath); err != nil {
			http.NotFound(w, r)
			return nil
		}

		if err != nil {
			return err
		}

		http.ServeFile(w, r, imgPath)

		return nil
	}))

	router.HandleFunc("/radcast.png", s.errorHandler(func(w http.ResponseWriter, r *http.Request) error {

		http.ServeFile(w, r, filepath.Join(s.Output, "radcast.png"))

		return nil
	}))

	return http.ListenAndServe(s.Addr, router)
}

func (s *Server) rss(baseURL *url.URL) (*PodcastRss, error) {

	dirs, err := ioutil.ReadDir(s.Output)

	if err != nil {
		return nil, err
	}

	items := PodcastItems{}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		item, err := s.itemByDir(dir.Name(), baseURL)

		if err != nil {
			s.Log(err)
			continue
		}

		items = append(items, *item)
	}

	sort.Sort(sort.Reverse(items))

	rss := NewPodcastRss()

	channel := PodcastChannel{}
	channel.Title = s.Title
	channel.Link = "http://radiko.jp"
	channel.Image.URL = baseURL.String() + "/radcast.png"
	channel.Image.Title = s.Title
	channel.Image.Link = "http://radiko.jp"
	channel.Description = "radiko"
	channel.Language = "ja-JP"
	channel.Copyright = "copyright 2019"

	channel.AtomLink.Href = baseURL.String() + "/rss"
	channel.AtomLink.Rel = "self"
	channel.AtomLink.Type = "application/rss+xml"
	channel.LastBuildDate = PubDate{time.Now()}

	channel.ITunesAuthor = "radiko"
	channel.ITunesSummary = "radiko"
	channel.ITunesSubtitle = "radiko"
	channel.ITunesOwner.ITunesName = "radiko"
	channel.ITunesOwner.ITunesEmail = "radiko@example.com"
	channel.ITunesExplicit = "No"
	channel.ITunesKeywords = "radiko,radio"
	channel.ITunesImage.Href = baseURL.String() + "/radcast.png"
	channel.ITunesCategory.Text = "Music"
	channel.PubDate = PubDate{time.Now()}

	channel.Items = items

	rss.Channel = channel

	return rss, nil
}

func (s *Server) itemByDir(dir string, baseURL *url.URL) (*PodcastItem, error) {

	_, medStat, err := s.medPath(dir)

	if err != nil {
		return nil, err
	}

	xmlPath, _, err := s.xmlPath(dir)

	if err != nil {
		return nil, err
	}

	xmlFile, err := os.Open(xmlPath)

	if err != nil {
		return nil, err
	}

	defer xmlFile.Close()

	dec := xml.NewDecoder(xmlFile)

	var prog RadikoProg
	if err := dec.Decode(&prog); err != nil {
		return nil, err
	}

	u, err := url.Parse("/podcast/" + dir + ".m4a")

	if err != nil {
		return nil, err
	}

	var item PodcastItem

	item.Title = prog.Title
	item.Link = prog.URL

	item.PubDate = PubDate{medStat.ModTime()}

	item.ITunesAuthor = prog.Pfm
	if utf8.RuneCountInString(prog.Info) == 0 {
		item.Description = prog.Desc
	} else {
		item.Description = prog.Info
	}
	item.Description += "<br><br>" + fmtDateTime(prog.Ft) + " - " + fmtDateTime(prog.To)
	item.Description += "<br><br><center><img src=\"" + prog.Img + "\" width=\"80%\"></center>"

	item.Enclosure.URL = baseURL.ResolveReference(u).String()
	item.Enclosure.Length = int(medStat.Size())
	item.Enclosure.Type = "audio/aac"

	item.GUID = dir
	item.ITunesDuration = fmtDuration(prog.Dur)
	item.ITunesSummary = item.Description

	ext := filepath.Ext(prog.Img)
	iu, err := url.Parse("/podcast/" + dir + ext)
	if err != nil {
		return nil, err
	}
	item.ITunesImage.Href = baseURL.ResolveReference(iu).String()

	return &item, nil
}

func (s *Server) medPath(dir string) (string, os.FileInfo, error) {
	return s.pathStat(dir, "podcast.m4a")
}

func (s *Server) xmlPath(dir string) (string, os.FileInfo, error) {
	return s.pathStat(dir, "podcast.xml")
}

func (s *Server) imgPath(dir string, ext string) (string, os.FileInfo, error) {
	return s.pathStat(dir, "podcast."+ext)
}

func (s *Server) pathStat(dir string, name string) (string, os.FileInfo, error) {
	p := filepath.Join(s.Output, dir, name)
	stat, err := os.Stat(p)

	if err != nil {
		return "", nil, err
	}

	return p, stat, nil
}

func (s *Server) Log(v ...interface{}) {
	log.Println("[server]", fmt.Sprint(v...))
}

func fmtDuration(sec string) string {
	d, _ := time.ParseDuration(sec + "s")
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func fmtDateTime(datetime string) string {
	return fmt.Sprintf("%s/%s/%s %s:%s:%s",
		datetime[0:4], datetime[4:6], datetime[6:8],
		datetime[8:10], datetime[10:12], datetime[12:14])
}
