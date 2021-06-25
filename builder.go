package main

import (
	"fmt"
	"log"
    "bytes"
    // "bufio"
	// "math"
	// "net/http"
	// "net/url"
    "reflect"
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
    "encoding/json"

	// "github.com/flosch/pongo2"
	// "github.com/howeyc/fsnotify"
	// "github.com/russross/blackfriday/v2"
    "github.com/yuin/goldmark"
    // "github.com/yuin/goldmark/extension"
    // "github.com/yuin/goldmark/parser"
    // "github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v2"
)

func checkFatal(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

type Meta struct {
    Title         string                     `yaml:"title"`        // meta
    Description   string                     `yaml:"description"`  //
    Tags          []string                   `yaml:"tags"`         //
    DatePublished string                     `yaml:"datePublished"`//
    DateModified  string                     `yaml:"dateModified"` //
    Draft         bool                       `yaml:"draft"`        //
    Fixed         bool                       `yaml:"fixed"`         //
    Option        []string                   `yaml:"option"`         //
    Layout        string                     `yaml:"layout"`       //
    Slug          string                     `yaml:"slug"`         //
    Permalink     string                     `yaml:"permalink"`         //
    Body          template.HTML              `yaml:"body"`         //
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

var cfg Config
// 全ページのデータリスト
var metalist = make([]Meta, 0)
var taglist []string


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

func (cfg *Config) convertFile(dirName, tpl string) {
    // テンプレートの読み込み
    t := template.Must(template.ParseFiles(tpl))

    // ディレクトリのファイル一覧を得る
    files, err := os.ReadDir(dirName)
    if err != nil {
        log.Fatal(err)
    }

    // dirName内のファイルをループ
    for _, file := range files {
        // markdownファイルからHTMLファイル作成
        var fpath = file.Name()
        fmt.Printf("fpath: %#v\n", fpath)
        srcFile := filepath.Join(dirName, fpath)
        // fmt.Println(srcFile)
        // ファイルの中身を読み取る
        b, e := os.ReadFile(srcFile)
        if e != nil {
            log.Fatal(e)
        }
        // 一旦 string型にして、frontmatter(metadata)を抜く
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
        // frontmatter を取り除いた部分を html に変換 
        var buf bytes.Buffer
        if err = goldmark.Convert([]byte(content), &buf); err != nil {
            panic(err)
        }
        new_html := buf.Bytes()

        // frontmatter を取得
        var meta Meta
        err = yaml.Unmarshal([]byte(b), &meta)
        if err != nil {
            log.Fatalf("error: %v", err)
        }
        // コンテンツ部分を 構造体へ追加 安全性に関して考える必要がある
        // fmt.Println(reflect.TypeOf(meta.Body))

        // 加工してから加える必要があるデータ
        slug := filepath.Base(fpath[:len(fpath)-len(filepath.Ext(fpath))])
        // fmt.Printf("file.Name(): %#v\n", file.Name())
        // fmt.Printf("fpath: %#v\n", fpath)
        // fmt.Printf("slug: %#v\n", slug)
        if slug == "index" {
            meta.Slug = "/"
            meta.Permalink = cfg.Baseurl
        } else {
            meta.Slug = slug
            meta.Permalink = urlJoin(cfg.Baseurl, slug)
        }

        // fmt.Printf("Bseurl: %#v\n", cfg.Baseurl)
        fmt.Printf("permalink: %#v\n", meta.Permalink)

        fmt.Println(reflect.TypeOf(meta))
        meta.Body = template.HTML(new_html)
        // fmt.Printf("meta: %#v\n", meta)

        new_buf := new(bytes.Buffer)
        dst := "./_site"
        if fpath == "index.md" && dirName == "./_pages" {
            dst = filepath.Join(dst, "index.html")
        } else {
            dstDir := filepath.Join(dst, slug)
            err = os.MkdirAll(dstDir, 0755)
            dst = filepath.Join(dst, slug, "index.html")
        }

        if err = t.Execute(new_buf, meta); err != nil {
            log.Println("create file", err)
        }
        // tag が1つ以上あったら taglist に追加
        if len(meta.Tags) > 0 {
            taglist = append(taglist, meta.Tags...)
        }
        metalist = append(metalist, meta)

        // fmt.Println(new_buf)
        os.WriteFile(dst, new_buf.Bytes(), 0644)
        fmt.Printf("WriteFile完了\n==================\n")

    }
}

func main() {
    fmt.Printf("cfg: %+v\n", cfg)
    // config を読み込み
    buf, err := os.ReadFile("config.yaml")
    if err != nil {
        log.Fatal(err)
    }
    // fmt.Printf("buf: %+v\n", string(buf))
    // var cfg Config
    err = yaml.Unmarshal(buf, &cfg)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("cfg: %+v\n", cfg)

    // テンプレートデータのpath
    layout := "./_layouts/default2.html"


    // pages ページ生成
    cfg.convertFile("./_pages", layout)
    // posts ページ生成
    cfg.convertFile("./_posts", layout)

    // tag ページ生成
    // ① taglistの重複を削除する
    taglistM := make(map[string]struct{})
    tagList := make([]string, 0)

    for _, elem := range taglist {
        // mapの第2引数には、その値が入っているかどうかの真偽値が入っている。
        if _, ok := taglistM[elem]; !ok && len(elem) != 0 {
            taglistM[elem] = struct{}{}
            tagList = append(tagList, elem)
        }
    }

    // 比較確認用
    fmt.Printf("mapの中身: %#v\n", taglistM)
    fmt.Printf("重複削除後: %#v\n", tagList)

    // ② tag ベースになる .md ファイルを生成（既にあるものはスルー）※postsを生成してからじゃないと実行できない
    dirName := "./_tags/"
    for _, tag := range tagList {
        s := []string{tag, ".md"}
        fName := strings.Join(s, "")
        srcFile := filepath.Join(dirName, fName)
        if !fileExists(srcFile) {
            copyFile("./_layouts/tag.md", srcFile)
        }
    }

    // tag ページ生成
    cfg.convertFile("./_tags", layout)

    // fmt.Printf("type of taglist: %#v\n", reflect.TypeOf(taglist))
    // fmt.Printf("taglist: %#v\n", taglist)

    jsDir := filepath.Join(cfg.Destination, "js")
    cssDir := filepath.Join(cfg.Destination, "css")
    err = os.MkdirAll(jsDir, 0755)
    err = os.MkdirAll(cssDir, 0755)
    copyFile("./_assets/js/main.js", "./_site/js/main.js")
    copyFile("./_assets/css/style.css", "./_site/css/style.css")

    f, err := os.OpenFile("./_site/js/main.js", os.O_APPEND | os.O_WRONLY, 0600)
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()
    fmt.Println(tagList);
    jsTagList := strings.Join(tagList, "\",\"")
    jsTagList = "var tags = [\"" + jsTagList + "\"];"
    fmt.Fprintln(f, jsTagList);

    // // 日付降順に並べる トップページの一覧用
    sort.Slice(metalist, func(i, j int) bool { return time2int(metalist[i].DateModified) > time2int(metalist[j].DateModified) })
    // fmt.Println(metalist)

    // json化 javascript で扱うときのため
    sample_json, _ := json.Marshal(&metalist)
    // fmt.Printf("[page-data.json]: %v\n", string(sample_json))
    os.WriteFile("./_site/data/page-data.json", sample_json, 0644)
}

