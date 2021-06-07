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

type config struct {
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

type fmeta struct {
    Title         string                     `yaml:"title"`
    Description   string                     `yaml:"description"`
    Tags          []string                   `yaml:"tags"`
    DatePublished string                     `yaml:"datePublished"`
    DateModifield string                     `yaml:"dateModified"`
    Draft         bool                       `yaml:"draft"`
    Body          template.HTML                     `yaml:"body"`
}

func time2int (str interface{}) int {
    a := str.(string)
    var i int
    b := strings.Replace(a, "-", "", -1)
    c := strings.Replace(b, ":", "", -1)
    d := strings.Replace(c, "T", "", -1)
    e := strings.Replace(d, "+", "", -1)
    i, _ = strconv.Atoi(e)
    return i
}

func main() {
    // テンプレートの読み込み
    t := template.Must(template.ParseFiles("./_layout/default2.html"))

    // md ファイル読み込み
    b, e := os.ReadFile("./_post/001.md")
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
    if err2 := goldmark.Convert([]byte(content), &buf); err2 != nil {
        panic(err2)
    }

    // ストリング型に戻した html を確認
    new_html := buf.Bytes()
    fmt.Println("complete convert markdown")

    // yaml の操作
    var meta map[string]interface{}
    var meta2 fmeta

    err3 := yaml.Unmarshal([]byte(b), &meta)
    if err3 != nil {
        log.Fatalf("error: %v", err3)
    }
    err4 := yaml.Unmarshal([]byte(b), &meta2)
    if err4 != nil {
        log.Fatalf("error: %v", err4)
    }
    fmt.Println(reflect.TypeOf(meta))
    fmt.Printf("meta: %#v\n\n", meta)
    fmt.Println(reflect.TypeOf(meta2))
    fmt.Printf("meta2: %#v\n\n", meta2)
    // fmt.Println(&meta)
    // コンテンツ部分を 構造体へ追加 安全性に関して考える必要がある
    meta["body"] = template.HTML(new_html)
    fmt.Println(reflect.TypeOf(meta["body"]))
    meta2.Body = template.HTML(new_html)
    fmt.Println(reflect.TypeOf(meta2))
    fmt.Printf("meta2: %#v\n\n", meta2)

    // os.WriteFile("./test-file.html", []byte(new_html), 0644)

    // ファイル生成 1
    // newfile, err := os.Create("./test-file2.html")
    // if err != nil {
    //     log.Println("create file", err)
    //     return
    // }
    // err = t.Execute(newfile, meta)
    // if err != nil {
    //     log.Println("create file", err)
    //     return
    // }
    // newfile.Close()

    // ファイル生成 2
    new_buf := new(bytes.Buffer)
    dst := "./test-file3.html"
    data := meta2
    if err := t.Execute(new_buf, data); err != nil {
        log.Println("create file", err)
    }
    // fmt.Println(new_buf)
    os.WriteFile(dst, new_buf.Bytes(), 0644)

    // ディレクトリのファイル一覧を得る
    dirName := "./_post/"
    files, err := os.ReadDir(dirName)
    if err != nil {
        log.Fatal(err)
    }
    var metalist []map[string]interface{}
    var taglist []string
    // var taglist []string
    // 取得した一覧を表示
    for _, file := range files {
        // fmt.Println(file)
        newfile := filepath.Join(dirName, file.Name())
        b, e := os.ReadFile(newfile)
        if e != nil {
            log.Fatal(e)
        }
        var metadata map[string]interface{}
        err3 := yaml.Unmarshal([]byte(b), &metadata)
        if err3 != nil {
            log.Fatalf("error: %v", err3)
        }
        // data := &metadata
        fpath := file.Name()
        slug := filepath.Base(fpath[:len(fpath)-len(filepath.Ext(fpath))])
        metadata["slug"] = slug
        metalist = append(metalist, metadata)
        tagsInterface := metadata["tags"].([]interface{})
        tagsString := make([]string, len(tagsInterface))
        for i, v := range tagsInterface {
            tagsString[i] = v.(string)
            for _, tag := range tagsString {
                taglist = append(taglist, tag)
            }
        }
        // fmt.Printf("tagsString: %v", tagsString)

        // for _, v := range tags {
        //     fmt.Println(v)
        // }
    }
    fmt.Printf("type of taglist: %#v\n", reflect.TypeOf(taglist))
    fmt.Printf("taglist: %#v\n", taglist)

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


    // fmt.Println(reflect.TypeOf(metalist))
    // fmt.Println(reflect.TypeOf(metalist[5]["datePublished"]))
    // fmt.Println(metalist[5]["datePublished"])

    // tag のページを作成
    // for _, v := range tagList {
    //     tagDirName := d
    // }

    // encode json

    // 日付降順に並べる
    sort.Slice(metalist, func(i, j int) bool { return time2int(metalist[i]["dateModified"]) > time2int(metalist[j]["dateModified"]) })
    // fmt.Println(metalist)

    // json化 javascript で扱うときのため
    sample_json, _ := json.Marshal(&metalist)
    fmt.Printf("[page-data.json]: %v\n", string(sample_json))
}

