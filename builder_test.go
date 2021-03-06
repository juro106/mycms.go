package main

import (
    "testing"
	"fmt"
	"log"
    "bytes"
    // "bufio"
	// "math"
	// "net/http"
	// "net/url"
    // "reflect"
	"io"
	"os"
	// "os/exec"
	"path/filepath"
	"path"
	// "runtime"
	"sort"
	"strconv"
	"strings"
	// "time"
    "html/template"
    // "encoding/json"
    "sync"

	// "github.com/flosch/pongo2"
	// "github.com/howeyc/fsnotify"
	// "github.com/russross/blackfriday/v2"
    "github.com/yuin/goldmark"
    // "github.com/yuin/goldmark/extension"
    // "github.com/yuin/goldmark/parser"
    // "github.com/yuin/goldmark-highlighting"
    // "github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v2"
)

func checkFatal(err error) {
    if err != nil {
        log.Fatal(err)
    }
}
type Item struct {
    Title string
    Slug  string
    Body  template.HTML
    Description string
}
type PageB2Tag map[string][]Item

type Meta struct {
    Title         string                     `yaml:"title"`        // meta
    Description   string                     `yaml:"description"`  //
    Tags          []string                   `yaml:"tags"`         //
    DateP         string
    DatePublished string                     `yaml:"datePublished"`//
    DateM         string
    DateModified  string                     `yaml:"dateModified"` //
    Draft         bool                       `yaml:"draft"`        //
    Home          bool                       `yaml:"home"`        //
    Fixed         bool                       `yaml:"fixed"`         //
    Code          bool                       `yaml:"code"`         //
    Option        []string                   `yaml:"option"`         //
    Layout        string                     `yaml:"layout"`       //
    Slug          string                     `yaml:"slug"`         //
    Permalink     string                     `yaml:"permalink"`         //
    Body          template.HTML              `yaml:"body"`         //
    Dist          string
    PageTag       PageB2Tag
    B2Page        []Item
    Plist         []Meta
    Site          *Config
}

type Config struct {
    Baseurl     string                       `yaml:"baseurl"`
    Title       string                       `yaml:"title"`
    Source      string                       `yaml:"source"`
    Name        string                       `yaml:"name"`
    Destination string                       `yaml:"destination"`
    Posts       string                       `yaml:"posts"`
    Data        string                       `yaml:"data"`
    Includes    string                       `yaml:"includes"`
    Layouts     string                       `yaml:"layouts"`
    Permalink   string                       `yaml:"permalink"`
    Exclude     []string                     `yaml:"exclude"`
    Host        string                       `yaml:"host"`
    Port        int                          `yaml:"port"`
    LimitPosts  int                          `yaml:"limit_posts"`
    MarkdownExt string                       `yaml:"markdown_ext"`
}

// config
var cfg Config
// ?????????????????????????????????
var metalist = make([]Meta, 0)
// ???????????????????????????
var postlist = make([]Meta, 0)
// ???????????????????????????
var pagelist = make([]Meta, 0)
// ???????????????????????????
var tagplist = make([]Meta, 0)
// tag ?????????
var taglist []string
// page-belong-to-tag ?????????
var pageB2taglist PageB2Tag = PageB2Tag{}
// map[string][]string = map[string][]string{}

func time2int (args interface{}) int {
    dateTime := args.(string)
    var i int
    dateTime = strings.Replace(dateTime, "-", "", -1)
    dateTime = strings.Replace(dateTime, ":", "", -1)
    dateTime = strings.Replace(dateTime, "T", "", -1)
    dateTime = strings.Replace(dateTime, "+", "", -1)
    i, _ = strconv.Atoi(dateTime)
    return i
}

func str(s interface{}) string {
	if ss, ok := s.(string); ok {
		return ss
	}
	return ""
}

func copyFile(src, dst string) (int64, error) {
    sf, err := os.Open(src)
    if err != nil {
        return 0, err
    }
    defer sf.Close()
    df, err := os.Create(dst)
    if err != nil {
        return 0, err
    }
    defer df.Close()
    return io.Copy(df, sf)
}

func fileExists(filename string) bool {
    _, err := os.Stat(filename)
    return err == nil
}

func urlJoin(l, r string) string {
    r = path.Clean(r)
    ls := strings.HasSuffix(l, "/")
    rp := strings.HasPrefix(r, "/")

    if ls && rp {
        return l + r[1:] + "/"
    }
    if !ls && !rp {
        return l + "/" + r + "/"
    }
    return l + r + "/"
}

// ?????????????????????????????????????????????????????????????????????????????????
func (cfg *Config) collectData(dirName string) {
    // ????????????????????????????????????????????????
    files, err := os.ReadDir(dirName)
    if err != nil {
        log.Fatal(err)
    }
    // dirName??????????????????????????????
    for _, file := range files {
        // ???????????????????????????????????????
        var meta Meta
        // ??????????????????????????? metadata ????????????
        meta.Site = cfg
        // ????????????????????????????????????
        var fpath = file.Name()
        srcFile := filepath.Join(dirName, fpath)
        b, e := os.ReadFile(srcFile)
        if e != nil {
            log.Fatal(e)
        }
        // ?????? string???????????????frontmatter(metadata)?????????
        content := string(b)
        lines := strings.Split(content, "\n")
        if len(lines) > 2 && lines[0] == "---" {
            var n int
            var line string
            for n, line = range lines[1:] {
                if line == "---" {
                    break
                }
            }
            content = strings.Join(lines[n+2:], "\n")
        }
        // frontmatter ??????????????????????????? html ????????? 
        var buf bytes.Buffer
        if err = goldmark.Convert([]byte(content), &buf); err != nil {
            panic(err)
        }
        // markdown ?????? html ???????????????????????? Body ?????????
        meta.Body = template.HTML(buf.Bytes())
        // frontmatter ?????????
        err = yaml.Unmarshal([]byte(b), &meta)
        if err != nil {
            log.Fatalf("error: %v", err)
        }
        // slug ?????????
        slug := filepath.Base(fpath[:len(fpath)-len(filepath.Ext(fpath))])
        if slug == "index" {
            meta.Slug = "/"
            meta.Permalink = cfg.Baseurl
        } else {
            meta.Slug = slug
            meta.Permalink = urlJoin(cfg.Baseurl, slug)
        }
        // ????????????
        if len(meta.DatePublished) > 0 {
            meta.DateP = meta.DatePublished[:10]
        }
        if len(meta.DateModified) > 0 {
            meta.DateM = meta.DateModified[:10]
        }
        // ????????????????????????
        dst := "./_site"
        if fpath == "index.md" && dirName == "./_pages" {
            meta.Dist = filepath.Join(dst, "index.html")
        } else {
            dstDir := filepath.Join(dst, slug)
            err = os.MkdirAll(dstDir, 0755)
            meta.Dist = filepath.Join(dst, slug, "index.html")
        }
        // tag ???1????????????????????? taglist ?????????
        if len(meta.Tags) > 0 {
            // page-tag-list
            taglist = append(taglist, meta.Tags...)
            // tag-page-list
            for _, tag := range meta.Tags {
                var item Item
                item.Title = meta.Title
                item.Slug = meta.Slug
                item.Description = meta.Description
                item.Body = meta.Body
                pageB2taglist[tag] = append(pageB2taglist[tag], item)
            }
        }
        // ???????????????????????????
        if dirName == "./_pages" {
            pagelist = append(pagelist, meta)
        } else if dirName == "./_posts" {
            postlist = append(postlist, meta)
        } else if dirName == "./_tags" {
            tagplist = append(tagplist, meta)
        }
        // ??????????????????????????????
        metalist = append(metalist, meta)
    }
}

// ???????????????????????????????????????
func (cfg *Config) convertFile(tpl, ptype string) {
    var list []Meta
    if ptype == "page" {
        list = pagelist
    } else if ptype == "post" {
        list = postlist
    } else if ptype == "tag" {
        list = tagplist
    }
    // ?????????????????????????????????
    t := template.Must(template.ParseFiles(tpl, "./_layouts/head.html", "./_layouts/footer.html"))

    // semaphore
    semaphore := make(chan struct{}, 10)
    var wg sync.WaitGroup
    // var mu sync.Mutex

    for _, meta := range list {
        wg.Add(1)
        semaphore <- struct{}{}
        go func() {
            defer func() {
                wg.Done()
                <-semaphore
            }()

            if ptype == "page" && meta.Slug == "/" {
                meta.Plist = postlist
            }
            if ptype == "post" {
                var postB2tag PageB2Tag = PageB2Tag{}
                for _, tag := range meta.Tags {
                    postB2tag[tag] = pageB2taglist[tag]
                }
                meta.PageTag = postB2tag
            }
            if ptype == "tag" {
                meta.B2Page = pageB2taglist[meta.Slug]
            }
            // ????????????????????????????????????????????????
            new_buf := new(bytes.Buffer)
            if err := t.Execute(new_buf, meta); err != nil {
                log.Println("create file", err)
            }
            os.WriteFile(meta.Dist, new_buf.Bytes(), 0644)
            fmt.Printf("%s WriteFile?????? ======>>>\n", meta.Dist)
        }()
    }
    wg.Wait()
}

func main() {
    // fmt.Printf("cfg: %+v\n", cfg)
    // config ???????????????
    buf, err := os.ReadFile("config.yaml")
    if err != nil {
        log.Fatal(err)
    }
    err = yaml.Unmarshal(buf, &cfg)
    if err != nil {
        log.Fatal(err)
    }
    // ??????????????? 
    cfg.collectData("./_pages")
    cfg.collectData("./_posts")

    // tag md??????????????????
    // ??? taglist????????????????????????
    taglistM := make(map[string]struct{})
    tagList := make([]string, 0)

    for _, elem := range taglist {
        // map??????2???????????????????????????????????????????????????????????????????????????????????????
        if _, ok := taglistM[elem]; !ok && len(elem) != 0 {
            taglistM[elem] = struct{}{}
            tagList = append(tagList, elem)
        }
    }
    // fmt.Printf("map?????????: %#v\n", taglistM)
    // fmt.Printf("???????????????: %#v\n", tagList)
    // ??? tag ?????????????????? .md ????????????????????????????????????????????????????????????posts??????????????????????????????????????????????????????
    dirName := "./_tags/"
    for _, tag := range tagList {
        s := []string{tag, ".md"}
        fName := strings.Join(s, "")
        srcFile := filepath.Join(dirName, fName)
        if !fileExists(srcFile) {
            copyFile("./_layouts/tag.md", srcFile)
        }
    }
    cfg.collectData("./_tags")

    // pages ???????????????
    cfg.convertFile("./_layouts/default.html", "page")
    // posts ???????????????
    cfg.convertFile("./_layouts/post.html", "post")
    // tag ???????????????
    cfg.convertFile("./_layouts/tag.html", "tag")

    func BenchmarkConvertFile (b *testing.B) {
        for n := 0; n < b.N; n++ {
            cfg.convertFile("./_layouts/post.html", "post")
        }
    }
    // fmt.Printf("type of taglist: %#v\n", reflect.TypeOf(taglist))

    jsDir := filepath.Join(cfg.Destination, "js")
    cssDir := filepath.Join(cfg.Destination, "css")
    err = os.MkdirAll(jsDir, 0755)
    err = os.MkdirAll(cssDir, 0755)
    // copyFile("./_assets/js/main.js", "./_site/js/main.js")
    // copyFile("./_assets/js/top-page.js", "./_site/js/top-page.js")
    copyFile("./_assets/js/prism.js", "./_site/js/prism.js")
    copyFile("./_assets/css/style.css", "./_site/css/style.css")

    // f, err := os.OpenFile("./_site/js/main.js", os.O_APPEND | os.O_WRONLY, 0600)
    // if err != nil {
    //     log.Fatal(err)
    // }
    // defer f.Close()
    // fmt.Printf("tagList: %s\n", tagList)
    // jsTagList := strings.Join(tagList, "\",\"")
    // jsTagList = "var tags = [\"" + jsTagList + "\"];"
    // fmt.Fprintln(f, jsTagList);

    // // ???????????????????????? ??????????????????????????????
    sort.Slice(postlist, func(i, j int) bool { return time2int(postlist[i].DateModified) > time2int(postlist[j].DateModified) })
    // for _, elem := range metalist {
    //     fmt.Printf("%v\n", elem.Dist)
    // }

    // // json??? javascript ????????????????????????
    // data_json, _ := json.Marshal(&metalist)
    // post_json, _ := json.Marshal(&postlist)
    // // fmt.Printf("[page-data.json]: %v\n", string(sample_json))
    // os.WriteFile("./_site/data/page-data.json", data_json, 0644)
    // os.WriteFile("./_site/data/post-data.json", post_json, 0644)

    sitemap := make([]string, 0)
    plist := make([]string, 0)
    for _, post := range metalist {
        p := "<url><loc>" + post.Permalink + "</loc></url>"
        plist = append(plist, p)
    }
    // fmt.Printf("plist: %s\n", plist)
    sitemap = append(sitemap, "<urlset xmlns='http://www.sitemaps.org/schemas/sitemap/0.9'>")
    // strings.Join(plist, "")
    sitemap = append(sitemap, plist...)
    sitemap = append(sitemap, "</urlset>")
    sitemapStr := strings.Join(sitemap, "\n")
    // fmt.Printf("sitemap: %s\n", sitemapStr)
    os.WriteFile("./_site/sitemap.xml", []byte(sitemapStr), 0644)
}

