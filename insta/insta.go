package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/net/proxy"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	flagUserName   = flag.String("username", "hre926", "instagram username")
	flagImgUrl     = flag.String("imgurl", "", "download the image url")
	flagDir        = flag.String("dir", "./uploadir", "instagram save path")
	flagHttpProxy  = flag.String("http_proxy", "", "http proxy url")
	flagTcpProxy   = flag.String("tcp_proxy", "", "tcp proxy")
	flagPageSize   = flag.Int64("pagesize", 100, "number each page to download")
	flagImgNum     = flag.Int64("imgnum", 0, "download the image number")
	flagThumbnails = flag.Bool("thumbnails", false, "whether to download a thumbnail")
	help           = flag.Bool("h", false, "help")
	version        = flag.Bool("version", false, "insta version")
)
var (
	instagramHomeUrl  = "https://www.instagram.com"
	instagramQueryUrl = instagramHomeUrl + "/query/"
	userAgent         = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_5)" +
		" AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36"
	csrftoken        string
	referer          string
	imgCount         int64
	imgNodes         []*Nodes
	defaultCookieJar http.CookieJar
	transport        *http.Transport
	loger            *log.Logger
)

func init() {
	defaultCookieJar, _ = cookiejar.New(nil)
	transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
}
func main() {
	loger = log.New(os.Stdout, "", log.LstdFlags)
	loger.Println("insta start...")
	flag.Parse()
	if *help {
		Help()
		return
	}
	if *version {
		Version()
		return
	}
	if *flagHttpProxy != "" {
		urli := url.URL{}
		httpProxy, err := urli.Parse(*flagHttpProxy)
		if err != nil {
			loger.Println(err)
			return
		}
		transport.Proxy = http.ProxyURL(httpProxy)
	}
	if *flagTcpProxy != "" {
		dialer, err := proxy.SOCKS5("tcp", *flagTcpProxy, nil, proxy.Direct)
		if err != nil {
			loger.Println(err)
			return
		}
		transport.Dial = dialer.Dial
	}
	dir := filepath.Join(*flagDir, *flagUserName)
	if !IsExist(dir) {
		Mkdir(dir)
	}
	defer func() {
		for _, v := range imgNodes {
			err := downLoadImg(v.DisplaySrc, dir, "")
			if err != nil {
				loger.Println(err)
			}
			if *flagThumbnails {
				err = downLoadImg(v.ThumbnailSrc, dir, "min_")
				if err != nil {
					loger.Println(err)
				}
			}
			if *flagImgNum > 0 && imgCount >= *flagImgNum {
				break
			}

		}
		loger.Println("download photo number is", imgCount)
	}()
	if *flagImgUrl != "" {
		imgNodes = append(imgNodes, &Nodes{DisplaySrc: *flagImgUrl})
		return
	}
	userHomeUrl := instagramHomeUrl + "/" + *flagUserName
	referer = userHomeUrl
	homePage, err := HttpGetToString(userHomeUrl)
	if err != nil {
		loger.Println(err.Error())
		return
	}
	extryData := findStringSubmatch(homePage, `window._sharedData = (.*);`)

	if extryData == "" {
		loger.Println("homePage extry_data is null")
		return
	}
	var ih InstagramHome
	err = json.Unmarshal([]byte(extryData), &ih)
	if err != nil {
		loger.Println("homePage extry_data decoding err")
		return
	}
	if len(ih.ExtryData.ProfilePage) == 0 {
		loger.Println("homePage ProfilePage is null")
		return
	}
	var user = ih.ExtryData.ProfilePage[0].User
	for _, v := range user.Media.Nodes {
		imgNodes = append(imgNodes, v)
	}
	if !user.Media.PageInfo.HasNextPage {
		return
	}
	var (
		userId     = user.Id
		mediaStart = user.Media.PageInfo.EndCursor
	)
	for {
		var q = fmt.Sprintf("ig_user(%s) { media.after(%s, %d) {nodes { code, date, display_src, thumbnail_src},page_info}}",
			userId, mediaStart, *flagPageSize)
		param := url.Values{}
		param.Add("q", q)
		var i InstagramQuery
		err := HttpGetToJson(instagramQueryUrl, param, &i)
		if err != nil {
			loger.Println(err.Error())
			break
		}

		if i.Status != "ok" {
			loger.Println("instagram status is not ok")
			break
		}
		for _, v := range i.Media.Nodes {
			imgNodes = append(imgNodes, v)
		}
		if !i.Media.PageInfo.HasNextPage {
			break
		}
		if i.Media.PageInfo.EndCursor == "" {
			break
		}
		mediaStart = i.Media.PageInfo.EndCursor
	}
}

func downLoadImg(imgUrl, dir, prefix string) error {
	if imgUrl == "" {
		return errors.New("ImgUrl Is Null")
	}
	filename := findStringSubmatch(imgUrl, `\/([0-9A-Za-z_.]*)\.*$`)
	if filename == "" {
		return errors.New("From " + imgUrl + " To File Failed")
	}
	if prefix != "" {
		filename = prefix + filename
	}
	var fullFilename = filepath.Join(dir, filename)
	if IsExist(fullFilename) {
		errInfo := fmt.Sprintf("filename is exist,filename:%s", fullFilename)
		return errors.New(errInfo)
	}
	HttpGetToFile(imgUrl, fullFilename)
	loger.Println("download success:" + fullFilename)
	imgCount++
	return nil
}
func findAllStringSubmatch(s, m string) []string {
	re := regexp.MustCompile(m)
	all := re.FindAllStringSubmatch(s, -1)
	var ss = []string{}
	for _, v := range all {
		if len(v) == 2 {
			ss = append(ss, v[1])
		}
	}
	return ss
}

func findStringSubmatch(s, m string) string {
	re := regexp.MustCompile(m)
	mc := re.FindStringSubmatch(s)
	if len(mc) == 2 {
		return mc[1]
	}
	return ""
}
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
func Mkdir(src string) error {
	if IsExist(src) {
		return nil
	}
	if err := os.MkdirAll(src, 0777); err != nil {
		if os.IsPermission(err) {
		}
		return err
	}

	return nil
}

func HttpGetToString(url string) (string, error) {
	resp, err := httpGet(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if csrftoken == "" {
		for _, v := range resp.Cookies() {
			if v.Name == "csrftoken" {
				csrftoken = v.Value
			}

		}
	}
	return string(b), nil
}
func HttpGetToFile(url, filename string) error {
	resp, err := httpGet(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fs, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fs.Close()
	_, err = io.Copy(fs, resp.Body)
	return err
}
func HttpGetToJson(url string, data url.Values, v interface{}) error {
	resp, err := httpPostForm(url, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
func httpPostForm(url string, data url.Values) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader((data.Encode())))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("referer", referer)
	req.Header.Set("x-csrftoken", csrftoken)
	client := &http.Client{
		Transport: transport,
		Jar:       defaultCookieJar,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
func httpGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	client := &http.Client{
		Transport: transport,
		Jar:       defaultCookieJar,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
func Help() {
	loger.Printf(
		"\nUseage: %s [ Options ]\n\n"+
			"Options:\n"+
			"  -username      instagram username [Default: hre926]\n"+
			"  -imgurl        download the image url\n"+
			"  -dir           instagram save path  [Default: uploadir]\n"+
			"  -http_proxy    http proxy url  [Example: http://127.0.0.1:1087]\n"+
			"  -tcp_proxy     tcp proxy  [Example: 127.0.0.1:1086]\n"+
			"  -pagesize      number each page to download [Default: 100]\n"+
			"  -imgnum        download the image max number\n"+
			"  -thumbnails    whether to download a thumbnail\n"+
			"  -h             help\n"+
			"  -version       insta version\n"+
			"------------------------------------------------------\n\n",
		os.Args[0])

	os.Exit(0)
}
func Version() {
	loger.Println("insta version:1.0")
}

//homePage struct
type InstagramHome struct {
	Hostname  string     `json:"hostname"`
	ExtryData *ExtryData `json:"entry_data"`
}
type ExtryData struct {
	ProfilePage []*ProfilePage `json:"ProfilePage"`
}
type ProfilePage struct {
	User *User `json:"user"`
}
type User struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Fullname string `json:"full_name"`
	Media    *Media `json:"media"`
}

//query response struct
type InstagramQuery struct {
	Status string `json:"status"`
	Media  *Media `json:"media"`
}
type Media struct {
	Nodes    []*Nodes  `json:"nodes"`
	PageInfo *PageInfo `json:"page_info"`
}
type Nodes struct {
	Code         string `json:"code"`
	Date         int64  `json:"date"`
	ThumbnailSrc string `json:"thumbnail_src"`
	DisplaySrc   string `json:"display_src"`
}
type PageInfo struct {
	HasPreviousPage bool   `json:"has_previous_page"`
	StartCursor     string `json:"start_cursor"`
	EndCursor       string `json:"end_cursor"`
	HasNextPage     bool   `json:"has_next_page"`
}
