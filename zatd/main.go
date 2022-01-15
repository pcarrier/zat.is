package main

import (
	"context"
	pp "github.com/jackc/pgx/v4/pgxpool"
	h "github.com/valyala/fasthttp"
	"log"
	"os"
	"regexp"
	"strings"
)

var (
	strLocation        = []byte("Location")
	base32At128Path, _ = regexp.Compile("/\\.([a-zA-Z2-7]{26})")
	parms, _           = regexp.Compile("(gif|png|txt)([0-9]*)(?:[.+]([0-9a-fA-F]{3,8})(?:[.+]([0-9a-fA-F]{3,8}))?)?")
	pool               *pp.Pool
)

func serve(ctx *h.RequestCtx) {
	host := string(ctx.Host())
	log.Println(host)
	if strings.HasPrefix(host, "l.") {
		serveLink(ctx)
	} else if strings.HasPrefix(host, "qr.") {
		serveQR(ctx)
	} else {
		serveRoot(ctx)
	}
}

func main() {
	var err error
	pool, err = pp.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	err = h.ListenAndServe(":8080", serve)
	if err != nil {
		log.Fatal(err)
	}
}
