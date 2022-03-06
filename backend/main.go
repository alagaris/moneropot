package main

import (
	"embed"
	"log"
	"net/http"

	"moneropot/api"
	"moneropot/db"
	"moneropot/util"
)

//go:embed dist
var staticFS embed.FS

func main() {
	util.ParseArgs()
	db.Init()

	api.StaticFS = staticFS
	go db.RunBackground()
	srv := &http.Server{
		Handler: api.NewServer(),
		Addr:    util.Config.Bind,
	}
	log.Printf("moneropot: http://%s", util.Config.Bind)
	log.Fatal(srv.ListenAndServe())
}
