package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thinkerou/favicon"
)

//go:embed all:out
var f embed.FS

func getAssetFS() http.FileSystem {
	//wrappedFS, err := fs.Sub(f, "dist")
	wrappedFS, err := fs.Sub(f, "out")
	if err != nil {
		log.Fatal(err)
	}
	return http.FS(wrappedFS)
}

func main() {
	app := gin.New()

	//错误捕获
	app.Use(gin.Recovery())
	//配置网页左上角图标
	//app.Use(favicon.New("./dist/favicon.ico"))
	app.Use(favicon.New("./out/favicon.ico"))

	//静态文件
	app.StaticFS("/", getAssetFS())
	if err := app.Run(":8122"); err != nil {
		panic(err)
	}
}
