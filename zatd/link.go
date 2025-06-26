package main

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"
)

func shorten(path string) string {
	h := sha256.New()
	h.Write([]byte(path))
	return b32encoder.EncodeToString(h.Sum(nil)[:16])
}

func serveLink(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Cache-Control", "max-age=31536000")
	escaped := strings.TrimSpace(string(ctx.Request.Header.RequestURI()[1:]))
	unescaped, err := url.PathUnescape(escaped)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
	target := strings.TrimSpace(unescaped)
	shortened := shorten(target)
	err = insertUrlWhenAbsent(shortened, target, ctx)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	_, err = fmt.Fprintf(ctx, "https://zat.is/.%s", shortened)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}
}
