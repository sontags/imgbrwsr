package main

import (
	"bytes"
	"flag"
	"html/template"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codegangsta/martini"
	"github.com/disintegration/imaging"
	"github.com/sontags/imgbrwsr/thumbcache"
)

const (
	defaultCacheSize = 300
	defaultThumbSize = 200
)

type Page struct {
	Path      Path
	Title     string
	Size      int
	InnerSize int
	Links     []Link
}

type Link struct {
	Href  string
	Thumb string
	Text  string
}

type Path struct {
	Dirs []Dir
}

type Dir struct {
	Name string
	Link string
}

var cacheSize int
var thumbSize int

func init() {
	flag.IntVar(&cacheSize, "c", defaultCacheSize, "The thumbnail buffer size")
	flag.IntVar(&thumbSize, "s", defaultThumbSize, "The thumbnail size")
}

func getImage(name string) image.Image {
	file, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}

	img, err := jpeg.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	return img
}

func makeThumb(name string) image.Image {
	file, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}

	img, err := jpeg.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	file.Close()

	m := imaging.Thumbnail(img, thumbSize, thumbSize, imaging.Linear)
	return m
}

func getThumb(name string, c *thumbcache.ThumbCache) image.Image {
	tmb := thumbcache.Thumb{"", nil}
	if c.HasThumb(name) {
		tmb = c.GetThumb(name)
	} else {
		tmb = thumbcache.Thumb{name, makeThumb(name)}
		c.AddThumb(tmb)
	}
	return tmb.Image
}

func listDir(path string) string {
	if path == "" {
		path = "."
	}

	files, _ := ioutil.ReadDir(path)
	links := make([]Link, 0)

	templateMain, _ := template.New("Main").Parse(mainTemplate)

	for _, f := range files {
		if f.IsDir() {
			baseLink := path + "/" + f.Name()
			thumb := getDirThumb(baseLink)
			if thumb != "" {
				link := Link{
					Href:  "/" + baseLink,
					Thumb: getDirThumb(baseLink),
					Text:  f.Name(),
				}
				links = append(links, link)
			}
		} else {
			if strings.HasSuffix(f.Name(), ".jpg") {
				imgPath := path + "/" + f.Name()

				link := Link{
					Href:  "/image/" + imgPath,
					Thumb: "/thumb/" + imgPath,
					Text:  " ",
				}
				links = append(links, link)
			}
		}
	}

	page := Page{
		Path:      getPath(path),
		Title:     path,
		Size:      thumbSize,
		InnerSize: thumbSize - 20,
		Links:     links,
	}

	buff := bytes.NewBufferString("")
	templateMain.Execute(buff, page)
	return buff.String()
}

func getPath(path string) Path {
	dirs := make([]Dir, 0)

	name := filepath.Base(path)

	if name == "." {
		name = "Home"
	}

	dir := Dir{
		Name: name,
		Link: "/" + path,
	}

	parent := filepath.Dir(path)

	if path != "." {
		parentPath := getPath(parent)
		dirs = parentPath.Dirs
	}

	dirs = append(dirs, dir)

	path_array := Path{
		Dirs: dirs,
	}

	return path_array
}

func getDirThumb(path string) string {
	filesInDir, _ := ioutil.ReadDir(path)
	for _, file := range filesInDir {
		if !file.IsDir() {
			if strings.HasSuffix(file.Name(), ".jpg") {
				return "/thumb/" + path + "/" + file.Name()
			}
		} else {
			str := getDirThumb(path + "/" + file.Name())
			if str != "" {
				return str
			}
		}

	}

	return ""
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	m := martini.Classic()

	thumbs := make([]thumbcache.Thumb, (cacheSize))
	c := thumbcache.ThumbCache{0, thumbs}

	m.Get("/thumb/**", func(params martini.Params, res http.ResponseWriter) {
		res.Header().Set("Content-Type", "image/jpeg")
		thumb := getThumb(params["_1"], &c)
		_ = jpeg.Encode(res, thumb, &jpeg.Options{90})
	})

	m.Get("/image/**", func(params martini.Params, res http.ResponseWriter) {
		u, _ := url.Parse(params["_1"])
		request := u.Path
		res.Header().Set("Content-Type", "image/jpeg")
		thumb := getImage(request)
		_ = jpeg.Encode(res, thumb, &jpeg.Options{90})
	})

	m.Get("/**", func(params martini.Params) string {
		u, _ := url.Parse(strings.Trim(params["_1"], "/"))
		request := u.Path
		return listDir(request)
	})
	m.Run()
}
