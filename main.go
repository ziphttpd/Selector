// go:generate go run github.com/rakyll/statik -f -src=static
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
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rakyll/statik/fs"
	"github.com/xorvercom/util/pkg/json"

	// go generate -v で作成すること
	_ "github.com/ziphttpd/Selector/statik"
)

var (
	staticFs    http.FileSystem
	dir         = flag.String("config", "", "configuration directory")
	listenPort  = flag.Int("port", 8822, "listen port")
	contentType = map[string]string{
		".html": "text/html",
		".htm":  "text/html",
		".js":   "text/javascript",
		".json": "text/json",
		".css":  "text/css",
		".txt":  "text/plain",
		".bmp":  "image/bmp",
		".gif":  "image/gif",
		".ico":  "image/ico",
		".jpg":  "image/jpeg",
		".png":  "image/png",
		".svg":  "image/svg+xml",
	}
	command string
)

func main() {
	var err error

	// フラグ
	flag.Parse()
	if *dir == "" {
		exe, _ := os.Executable()
		*dir = fpath.Dir(exe)
	}

	// pidファイル作成
	pidfile := fpath.Join(*dir, "selector.pid")
	os.Remove(pidfile)
	if pidf, err := os.Create(pidfile); err == nil {
		fmt.Fprintf(pidf, "%d\n", os.Getpid())
		pidf.Close()
		defer os.Remove(pidfile)
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
	echoInst.HideBanner = true
	echoInst.Use(middleware.CORS())
	echoInst.Use(middleware.Logger())

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
	ext := fpath.Ext(name)
	ct := contentType[strings.ToLower(ext)]
	return c.Blob(http.StatusOK, ct, b)
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
	const tokenHeader = "X-Requested-With" // パスワードのヘッダ
	// 独自ヘッダの確認
	password := c.Request().Header.Get(tokenHeader)
	if password == "" || false == passwordCheck(password) {
		// 不正なリクエスト
		mes := "need header X-Requested-With"
		return c.String(http.StatusBadRequest, mes)
	}

	params, err := c.FormParams()
	if err != nil {
		fmt.Printf("err:%s", err)
		return err
	}

	// ドキュメントのダウンロードの登録
	err = zhget(params.Get("host"), params.Get("group"))
	if err != nil {
		fmt.Printf("err:%s", err)
		return err
	}
	return c.Blob(http.StatusOK, "text/html", nil)
}

// パスワードチェック (password.json に selector のパスワードがない場合にもＯＫ)
func passwordCheck(password string) bool {
	j, err := json.LoadFromJSONFile(fpath.Join(*dir, "password.json"))
	if err != nil {
		fmt.Printf("err:%s", err)
		return true
	}
	if jo, ok := j.AsObject(); ok {
		if js, ok := jo.Child("selector").AsString(); ok {
			if js.Text() != password {
				// パスワードが設定されている場合のみＮＧ
				return false
			}
		}
	}
	return true
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
