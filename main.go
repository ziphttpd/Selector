// go:generate go run github.com/rakyll/statik -src=static
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	fpath "path/filepath"
	"runtime"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rakyll/statik/fs"

	// go generate -v で作成すること
	_ "github.com/ziphttpd/Selector/statik"
)

var (
	staticFs   http.FileSystem
	dir        = flag.String("config", "", "configuration directory")
	listenPort = flag.Int("port", 8822, "listen port")
	command    string
)

func main() {
	var err error

	// フラグ
	flag.Parse()
	if *dir == "" {
		exe, _ := os.Executable()
		*dir = fpath.Dir(exe)
	}

	staticFs, err = fs.New()
	if err != nil {
		fmt.Printf("err:%s", err)
		return
	}
	if runtime.GOOS == "windows" {
		command = fpath.Join(*dir, "zhget.exe")
	} else {
		command = fpath.Join(*dir, "zhget")
	}

	echoInst := echo.New()
	echoInst.Use(middleware.CORS())

	// トップページ
	echoInst.GET("/", topPage)
	// 組み込みファイル
	echoInst.GET("/static/:name", staticFile)
	// サイト一覧
	echoInst.GET("/api/list", getList)
	// カタログ読み出し
	echoInst.GET("/api/catalog/:site", getCatalog)
	// ドキュメントのダウンロード登録
	echoInst.POST("/api/regist", registDoc)

	echoInst.Start(fmt.Sprintf(":%d", *listenPort))
}

// メインページ
func topPage(c echo.Context) error {
	file, err := staticFs.Open("/index.html")
	if err != nil {
		fmt.Printf("err:%s", err)
		return err
	}
	defer file.Close()
	b, _ := ioutil.ReadAll(file)
	return c.Blob(http.StatusOK, "text/html", b)
}

// 静的ファイルのハンドル
func staticFile(c echo.Context) error {
	name := c.Param("name")
	f, err := staticFs.Open("/" + name)
	if err != nil {
		fmt.Printf("err:%s", err)
		return err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Printf("err:%s", err)
		return err
	}
	return c.Blob(http.StatusOK, "text/html", b)
}

// サイト一覧
func getList(c echo.Context) error {
	url := "https://ziphttpd.com/api/v1/list"
	b, err := wget(url)
	if err != nil {
		fmt.Printf("err:%s", err)
		return err
	}
	return c.JSONBlob(http.StatusOK, b)
}

// サイトのカタログを読み出す
func getCatalog(c echo.Context) error {
	site := c.Param("site")
	b, err := wget(fmt.Sprintf("https://%s/sig/catalog.json", site))
	if err != nil {
		fmt.Printf("err:%s", err)
		return err
	}
	return c.JSONBlob(http.StatusOK, b)
}

// zhget の実行
func registDoc(c echo.Context) error {
	params, err := c.FormParams()
	if err != nil {
		fmt.Printf("err:%s", err)
		return err
	}
	err = zhget(params.Get("host"), params.Get("group"))
	if err != nil {
		fmt.Printf("err:%s", err)
		return err
	}
	return c.Blob(http.StatusOK, "text/html", nil)
}

// zhget
func zhget(host, group string) error {
	cmd := exec.Command(command, "-host", host, "-group", group)
	return cmd.Run()
}

// wget
func wget(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
