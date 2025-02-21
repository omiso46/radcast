package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	radikoTimeLayout = "20060102150405"
	authKey          = "bcd151073c03b352e1ef2fd66c32209da9ca0afa"
	streamMultiURL   = "https://radiko.jp/v3/station/stream/aHybrid01/%s.xml"
)

type RadikoPrograms struct {
	Stations struct {
		Station []struct {
			ID    string `xml:"id,attr"`
			Name  string `xml:"name"`
			Progs struct {
				Date string       `xml:"date"`
				Prog []RadikoProg `xml:"prog"`
			} `xml:"progs"`
		} `xml:"station"`
	} `xml:"stations"`
}

type RadikoProg struct {
	XMLName  xml.Name `xml:"prog"`
	ID       string   `xml:"id,attr"`
	MasterID string   `xml:"master_id,attr"`
	Ft       string   `xml:"ft,attr"`
	To       string   `xml:"to,attr"`
	Ftl      string   `xml:"ftl,attr"`
	Tol      string   `xml:"tol,attr"`
	Dur      string   `xml:"dur,attr"`
	Title    string   `xml:"title"`
	URL      string   `xml:"url"`
	Desc     string   `xml:"desc"`
	Info     string   `xml:"info"`
	Pfm      string   `xml:"pfm"`
	Img      string   `xml:"img"`
}

type StreamURLItem struct {
	AreaFree          bool   `xml:"areafree,attr"`
	TimeFree          bool   `xml:"timefree,attr"`
	MediaURLPath      string `xml:"media_url_path"`
	PlaylistCreateURL string `xml:"playlist_create_url"`
	PlaylistURLPath   string `xml:"playlist_url_path"`
}

type StreamUrlData struct {
	XMLName xml.Name        `XML:"urls"`
	URL     []StreamURLItem `xml:"url"`
}

func (r *RadikoProg) FtTime() (time.Time, error) {
	return time.ParseInLocation(radikoTimeLayout, r.Ft, time.Local)
}

func (r *RadikoProg) ToTime() (time.Time, error) {
	return time.ParseInLocation(radikoTimeLayout, r.To, time.Local)
}

func (r *RadikoProg) Duration() (int64, error) {
	to, err := r.ToTime()
	if err != nil {
		return 0, err
	}
	return to.Unix() - time.Now().Unix(), nil
}

type RadikoResult struct {
	MedPath string
	Prog    *RadikoProg
	Station string
}

func (r *RadikoResult) Save(dir string) error {
	programDir := filepath.Join(dir, fmt.Sprintf("%s_%s", r.Prog.Ft, r.Station))
	if err := os.MkdirAll(programDir, 0777); err != nil {
		return err
	}

	medPath := filepath.Join(programDir, "podcast.m4a")
	xmlPath := filepath.Join(programDir, "podcast.xml")

	imgName := "podcast" + filepath.Ext(r.Prog.Img)
	imgPath := filepath.Join(programDir, imgName)

	if err := RenameOrCopy(r.MedPath, medPath); err != nil {
		return err
	}

	if err := RenameOrCopy(filepath.Dir(r.MedPath)+"/"+imgName, imgPath); err != nil {
		return err
	}

	xmlFile, err := os.Create(xmlPath)
	if err != nil {
		return err
	}

	defer xmlFile.Close()

	enc := xml.NewEncoder(xmlFile)
	enc.Indent("", "    ")
	if err := enc.Encode(r.Prog); err != nil {
		return err
	}

	r.Log("saved path:", programDir)

	return nil
}

func (r *RadikoResult) Log(v ...interface{}) {
	log.Println("[radiko_result]", fmt.Sprint(v...))
}

type Radiko struct {
	Station   string
	Buffer    int64
	Converter string
	TempDir   string
	Result    *RadikoResult
}

func (r *Radiko) Run(ctx context.Context) error {

	results := r.run(ctx)

	switch len(results) {
	case 0:
		return fmt.Errorf("empty outputs")
	case 1:
		r.Result = results[0]
		return nil
	default:
		result, err := r.ConcatOutput(r.TempDir, results)
		if err != nil {
			return err
		}
		r.Result = result
		return nil
	}
}

func (r *Radiko) run(ctx context.Context) []*RadikoResult {
	errChan := make(chan error)

	retry := 0
	c := make(chan struct{}, 1)

	results := []*RadikoResult{}

	record := func() error {

		output := filepath.Join(r.TempDir, fmt.Sprintf("radiko_%d.m4a", retry))

		ret, err := r.record(ctx, output, r.Station, r.Buffer)

		if ret != nil {
			results = append(results, ret)
		}

		return err
	}

	c <- struct{}{}

	for {
		select {
		case <-c:
			r.Log("start record")
			go func() {
				errChan <- record()
			}()
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				r.Log("context err:", err)
			}

			select {
			case err := <-errChan:
				r.Log("err:", err)
			case <-time.After(time.Second * 10):
				r.Log("timeout receive err chan")
			}
			return results
		case err := <-errChan:
			r.Log("finished")
			if err == nil {
				return results
			}

			// TODO stop if recod program is changed.
			r.Log("got err:", err)
			if retry < 10 {
				sec := time.Second * 2
				time.AfterFunc(sec, func() {
					c <- struct{}{}
				})
				r.Log("retry after ", sec)
				retry++
			} else {
				return results
			}
		}
	}
}

// http://superuser.com/questions/314239/how-to-join-merge-many-mp3-files
func (r *Radiko) ConcatOutput(dir string, results []*RadikoResult) (*RadikoResult, error) {
	output := filepath.Join(dir, "radiko_concat.m4a")

	outputs := []string{}
	for _, result := range results {
		outputs = append(outputs, result.MedPath)
	}

	args := []string{
		"-i",
		fmt.Sprintf("concat:%s", strings.Join(outputs, "|")),
		"-acodec",
		"copy",
		output,
	}

	cmd := exec.Command(r.Converter, args...)
	r.Log("concat cmd:", strings.Join(cmd.Args, " "))

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return &RadikoResult{
		MedPath: output,
		Station: results[0].Station,
		Prog:    results[0].Prog,
	}, nil
}

func (r *Radiko) StationList(ctx context.Context) ([]string, error) {
	_, area, err := r.auth(ctx)
	if err != nil {
		return nil, err
	}

	progs, err := r.todayPrograms(ctx, area)
	if err != nil {
		return nil, err
	}

	stations := make([]string, len(progs.Stations.Station))

	for i, station := range progs.Stations.Station {
		stations[i] = station.ID
	}

	return stations, nil
}

func (r *Radiko) todayPrograms(ctx context.Context, area string) (*RadikoPrograms, error) {

	const layoutDate = "20060102"
	const layoutTime = "150405"

	timeNow := time.Now()
	nowTime := timeNow.Format(layoutTime)
	tmpDate := timeNow

	if nowTime >= "000000" && nowTime < "050000" {
		tmpDate = timeNow.AddDate(0, 0, -1)
	}

	u, err := url.Parse("http://radiko.jp/v3/program/date/" + tmpDate.Format(layoutDate) + "/" + area + ".xml")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	var progs RadikoPrograms
	err = r.httpDo(ctx, req, func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		if code := resp.StatusCode; code != 200 {
			return fmt.Errorf("not status code:200, got:%d", code)
		}

		return xml.NewDecoder(resp.Body).Decode(&progs)
	})

	if err != nil {
		return nil, err
	}

	return &progs, nil
}

func (r *Radiko) nowProgram(ctx context.Context, area string, station string) (*RadikoProg, error) {
	progs, err := r.todayPrograms(ctx, area)
	if err != nil {
		return nil, err
	}

	for _, s := range progs.Stations.Station {
		if s.ID == station {
			for _, prog := range s.Progs.Prog {
				ft, err := prog.FtTime()
				if err != nil {
					return nil, err
				}

				to, err := prog.ToTime()
				if err != nil {
					return nil, err
				}

				now := time.Now()

				if ft.Unix() <= now.Unix() && now.Unix() < to.Unix() {
					return &prog, nil
				}
			}
		}
	}

	return nil, errors.New("not found program")
}

func (r *Radiko) record(ctx context.Context, output string, station string, buffer int64) (*RadikoResult, error) {

	authtoken, area, err := r.auth(ctx)
	if err != nil {
		return nil, err
	}

	prog, err := r.nowProgram(ctx, area, station)
	if err != nil {
		return nil, err
	}

	os.Mkdir(r.TempDir, 0777)

	r.Log("Get Img:", prog.Img)
	img, err := http.Get(prog.Img)
	if err == nil {
		defer img.Body.Close()

		file, err := os.Create(filepath.Dir(output) + "/podcast" + filepath.Ext(prog.Img))
		if err == nil {
			defer file.Close()

			io.Copy(file, img.Body)

		}
	}

	r.Log("start recording ", prog.Title)
	duration, err := prog.Duration()
	if err != nil {
		return nil, err
	}

	duration += buffer

	err = r.hlsDownload(ctx, authtoken, station, fmt.Sprint(duration), output)

	if _, fileErr := os.Stat(output); fileErr != nil {
		return nil, err
	}

	ret := &RadikoResult{
		MedPath: output,
		Station: station,
		Prog:    prog,
	}

	return ret, err
}

// return authtoken, area, err
func (r *Radiko) auth(ctx context.Context) (string, string, error) {

	req, err := http.NewRequest("GET", "https://radiko.jp/v2/api/auth1", nil)
	if err != nil {
		return "", "", err
	}

	// req.Header.Set("pragma", "no-cache")
	// req.Header.Set("User-Agent", "radiko/4.0.1")
	// req.Header.Set("Accept", "*/*")
	req.Header.Set("X-Radiko-App", "pc_html5")
	req.Header.Set("X-Radiko-App-Version", "0.0.1")
	req.Header.Set("X-Radiko-User", "dummy_user")
	req.Header.Set("X-Radiko-Device", "pc")

	var authtoken string
	var partialkey string

	err = r.httpDo(ctx, req, func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		authtoken = resp.Header.Get("X-Radiko-Authtoken")
		keylength := resp.Header.Get("X-Radiko-Keylength")
		keyoffset := resp.Header.Get("X-Radiko-Keyoffset")

		if authtoken == "" {
			return errors.New("auth token is empty")
		}

		if keylength == "" {
			return errors.New("keylength is empty")
		}

		if keyoffset == "" {
			return errors.New("keyoffset is empty")
		}

		keylengthI, err := strconv.ParseInt(keylength, 10, 64)
		if err != nil {
			return err
		}

		keyoffsetI, err := strconv.ParseInt(keyoffset, 10, 64)
		if err != nil {
			return err
		}

		partialkeyByt := authKey[keyoffsetI : keyoffsetI+keylengthI]
		partialkey = base64.StdEncoding.EncodeToString([]byte(partialkeyByt))

		return nil
	})

	if err != nil {
		return "", "", err
	}

	req, err = http.NewRequest("GET", "https://radiko.jp/v2/api/auth2", nil)
	if err != nil {
		return "", "", err
	}

	// req.Header.Set("pragma", "no-cache")
	// req.Header.Set("User-Agent", "radiko/4.0.1")
	// req.Header.Set("Accept", "*/*")
	// req.Header.Set("X-Radiko-App", "pc_html5")
	// req.Header.Set("X-Radiko-App-Version", "0.0.1")
	req.Header.Set("X-Radiko-User", "dummy_user")
	req.Header.Set("X-Radiko-Device", "pc")
	req.Header.Set("X-Radiko-Authtoken", authtoken)
	req.Header.Set("X-Radiko-Partialkey", partialkey)

	var area string
	err = r.httpDo(ctx, req, func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		byt, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		matches := regexp.MustCompile("(.*),(.*),(.*)").FindAllStringSubmatch(string(byt), -1)
		if len(matches) == 1 && len(matches[0]) != 4 {
			return errors.New("failed to auth")
		}

		area = matches[0][1]

		return nil
	})

	if err != nil {
		return "", "", err
	}

	return authtoken, area, nil
}

func (r *Radiko) Log(v ...interface{}) {
	log.Println("[radiko]", fmt.Sprint(v...))
}

// http://blog.golang.org/context/google/google.go
func (r *Radiko) httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) error) error {
	r.Log(req.Method + " " + req.URL.String())

	errChan := make(chan error)

	go func() { errChan <- f(http.DefaultClient.Do(req)) }()

	select {
	case <-ctx.Done():
		http.DefaultTransport.(*http.Transport).CancelRequest(req)
		err := <-errChan
		if err == nil {
			err = ctx.Err()
		}
		return err
	case err := <-errChan:
		return err
	}
}

func (r *Radiko) GetStreamURL(stationID string) (string, error) {

	u := fmt.Sprintf(streamMultiURL, stationID)
	rsp, err := http.Get(u)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	urlData := StreamUrlData{}
	if err = xml.Unmarshal(b, &urlData); err != nil {
		return "", err
	}

	var streamURL string = ""
	for _, i := range urlData.URL {
		if !i.AreaFree && !i.TimeFree {
			streamURL = i.PlaylistCreateURL
			break
		}

	}

	return streamURL, err

}

func (r *Radiko) hlsDownload(ctx context.Context, authtoken string, station string, sec string, output string) error {

	streamURL, err := r.GetStreamURL(station)
	hlsRecCmd := hlsFfmpegCmd(r.Converter, streamURL, authtoken, sec, output)
	if err != nil {
		return err
	}

	r.Log("hlsFfmpegCmd:", strings.Join(hlsRecCmd.Args, " "))

	var errbuff bytes.Buffer
	hlsRecCmd.Stderr = &errbuff

	errChan := make(chan error)
	go func() {
		if err := hlsRecCmd.Run(); err != nil {
			r.Log("CmdRun err:" + errbuff.String())
			errChan <- err
			return
		}
		errChan <- nil

	}()

	select {
	case <-ctx.Done():
		err := <-errChan
		if err == nil {
			err = ctx.Err()
		}
		return err
	case err := <-errChan:
		return err
	}

}
