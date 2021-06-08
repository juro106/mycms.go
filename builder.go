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
    Code          bool                       `yaml:"code"`         //
    Css           []string                   `yaml:"css"`          //
    Js            []string                   `yaml:"js"`           //
    Layout        string                     `yaml:"layout"`       //
    Slug          string                     `yaml:"slug"`         //
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

func main() {
    // テンプレートの読み込み
    t := template.Must(template.ParseFiles("./_layout/default2.html"))

    // ディレクトリのファイル一覧を得る
    dirName := "./_post/"
    files, err := os.ReadDir(dirName)
    if err != nil {
        log.Fatal(err)
    }

    // タグのリスト
    var taglist []string
    // post のデータリスト
    metalist := make([]Meta, 0)

    // _post内のファイルをループ
    for _, file := range files {
        // markdownファイルからHTMLファイル作成
        fpath := file.Name()
        srcFile := filepath.Join(dirName, fpath)
        // fmt.Println(srcFile)
        slug := filepath.Base(fpath[:len(fpath)-len(filepath.Ext(fpath))])
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
        fmt.Println(reflect.TypeOf(meta))
        fmt.Printf("meta: %#v\n\n", meta)
        // コンテンツ部分を 構造体へ追加 安全性に関して考える必要がある
        fmt.Println(reflect.TypeOf(meta.Body))
        meta.Slug = slug
        metalist = append(metalist, meta)
        meta.Body = template.HTML(new_html)

        // os.WriteFile("./test-file.html", []byte(new_html), 0644)
        // ファイル生成
        new_buf := new(bytes.Buffer)
        dst := "./_site"
        dstDir := filepath.Join(dst, slug)
        err = os.MkdirAll(dstDir, 0755)
        dst = filepath.Join(dst, slug, "index.html")
        if err = t.Execute(new_buf, meta); err != nil {
            log.Println("create file", err)
        }
        // fmt.Println(new_buf)
        os.WriteFile(dst, new_buf.Bytes(), 0644)

        taglist = append(taglist, meta.Tags...)
    }

    // ------------------ page 生成
    dirName = "./_page/"
    files, err = os.ReadDir(dirName)
    if err != nil {
        log.Fatal(err)
    }
    // _page内のファイルをループ
    for _, file := range files {
        // markdownファイルからHTMLファイル作成
        fpath := file.Name()
        srcFile := filepath.Join(dirName, fpath)
        // fmt.Println(srcFile)
        slug := filepath.Base(fpath[:len(fpath)-len(filepath.Ext(fpath))])
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
        fmt.Println(reflect.TypeOf(meta))
        fmt.Printf("meta: %#v\n\n", meta)
        // コンテンツ部分を 構造体へ追加 安全性に関して考える必要がある
        fmt.Println(reflect.TypeOf(meta.Body))
        meta.Slug = slug
        metalist = append(metalist, meta)
        meta.Body = template.HTML(new_html)

        // os.WriteFile("./test-file.html", []byte(new_html), 0644)
        // ファイル生成
        new_buf := new(bytes.Buffer)
        dst := "./_site"
        if fpath == "index.md" {
            dst = filepath.Join(dst, "index.html")
        } else {
            dstDir := filepath.Join(dst, slug)
            err = os.MkdirAll(dstDir, 0755)
            dst = filepath.Join(dst, slug, "index.html")
        }

        if err = t.Execute(new_buf, meta); err != nil {
            log.Println("create file", err)
        }
        // fmt.Println(new_buf)
        os.WriteFile(dst, new_buf.Bytes(), 0644)
    }

    // taglistの重複を削除する
    taglistM := make(map[string]struct{})
    tagList := make([]string, 0)

    for _, elem := range taglist {
        // mapの第2引数には、その値が入っているかどうかの真偽値が入っている。
        if _, ok := taglistM[elem]; !ok && len(elem) != 0 {
            taglistM[elem] = struct{}{}
            tagList = append(tagList, elem)
        }
    }

    fmt.Printf("mapの中身: %#v\n", taglistM)
    fmt.Printf("重複削除後: %#v\n", tagList)

    // -----------------それぞれの tag の index page 生成
    dirName = "./_tags/"
    files, err = os.ReadDir(dirName)
    if err != nil {
        log.Fatal(err)
    }

    for _, tag := range tagList {
        s := []string{tag, ".md"}
        fName := strings.Join(s, "")
        srcFile := filepath.Join(dirName, fName)
        if !fileExists(srcFile) {
            copyFile("./_layout/tag.md", srcFile)
        }
    }

    // _tags 内のファイルをループ
    for _, file := range files {
        // markdownファイルからHTMLファイル作成
        fpath := file.Name()
        srcFile := filepath.Join(dirName, fpath)
        // fmt.Println(srcFile)
        slug := filepath.Base(fpath[:len(fpath)-len(filepath.Ext(fpath))])
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
        fmt.Println(reflect.TypeOf(meta))
        fmt.Printf("meta: %#v\n\n", meta)
        // コンテンツ部分を 構造体へ追加 安全性に関して考える必要がある
        fmt.Println(reflect.TypeOf(meta.Body))
        meta.Slug = slug
        metalist = append(metalist, meta)
        meta.Body = template.HTML(new_html)

        // os.WriteFile("./test-file.html", []byte(new_html), 0644)
        // ファイル生成
        new_buf := new(bytes.Buffer)
        dst := "./_site"
        dstDir := filepath.Join(dst, slug)
        err = os.MkdirAll(dstDir, 0755)
        dst = filepath.Join(dst, slug, "index.html")
        if err = t.Execute(new_buf, meta); err != nil {
            log.Println("create file", err)
        }
        // fmt.Println(new_buf)
        os.WriteFile(dst, new_buf.Bytes(), 0644)

        taglist = append(taglist, meta.Tags...)
    }

    // ------------------- index.html

    // fmt.Printf("type of taglist: %#v\n", reflect.TypeOf(taglist))
    // fmt.Printf("taglist: %#v\n", taglist)



    // // 日付降順に並べる トップページの一覧用
    sort.Slice(metalist, func(i, j int) bool { return time2int(metalist[i].DateModified) > time2int(metalist[j].DateModified) })
    // fmt.Println(metalist)

    // json化 javascript で扱うときのため
    sample_json, _ := json.Marshal(&metalist)
    fmt.Printf("[page-data.json]: %v\n", string(sample_json))
    os.WriteFile("./_site/data/page-data.json", sample_json, 0644)
}

