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
	mainTemplate     = `<html>
	<head>
		<title>{{.Title}}</title>
		<style>
			* {
				margin: 0px;
				padding: 0px;
			}
			#navigation {
    			position: fixed;
    			height: 50px;
    			top: 0;
    			width: 100%;
    			z-index: 100;
    			background-color: #EEE;
    			-webkit-box-shadow: 0px 10px 15px 0px rgba(0,0,0,0.35);
    			-moz-box-shadow: 0px 10px 15px 0px rgba(0,0,0,0.35);
    			box-shadow: 0px 10px 15px 0px rgba(0,0,0,0.35);
			}
			#content { 
    			margin-top: 80px;
				text-align: center;
				width: 100%;
			}
		    .nav {
		    	margin-top: 10px;
				clear: both;
				display: inline-block;
				position: relative;
		    }
		    .nav a {
		    	padding: 10px;
		    	font-family: sans-serif;
		    	text-decoration: none;
		    	font-size: 25px;
		    	color: #333;
		    	text-transform: uppercase;
		    	letter-spacing: -1px;
		    }
			.thumb {
				width: {{.Size}}px;
				height: {{.Size}}px;
				background-color: #F6CECE;
				clear: both;
				display: inline-block;
				position: relative;
			}
			.thumb-inner {
				display: table;
				position: absolute; 
  				left: 10px; 
  				top: 10px; 
  				width: {{.InnerSize}}px; 
  				height: {{.InnerSize}}px; 
			}
			.thumb-inner p {
				display: table-cell; 
				vertical-align: middle; 
				text-align: center; 
				font-family: sans-serif;
				color: white;
				text-shadow: 2px 2px 2px rgba(150, 150, 150, 1);
				font-weight: bold;
			}
		</style>
	</head>
	<body>
		<div id="navigation">
		{{with .Path.Dirs}}{{range .}}
			<div class="nav"><a href="{{.Link}}">&#8226;&nbsp;&nbsp;&nbsp;{{.Name}}</a></div>
		{{end}}{{end}}</div>
		
		<div id="content">
		{{with .Links}}{{range .}}
			<a href="{{.Href}}">
				<div class='thumb' style='background-image:url("{{.Thumb}}");'>
					<div class='thumb-inner'><p>{{.Text}}</p></div></div></a>
		{{end}}{{end}}
		</div>
	</body>
</html>`
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
			link := Link{
				Href:  "/" + baseLink,
				Thumb: getDirThumb(baseLink),
				Text:  f.Name(),
			}
			links = append(links, link)
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
