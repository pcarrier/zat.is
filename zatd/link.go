package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"net/url"
)

func shorten(path string) string {
	u := uuid.NewSHA1(uuid.NameSpaceURL, []byte(path))
	return b32encoder.EncodeToString(u[:])
}

func serveLink(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Cache-Control", "max-age=31536000")
	target, err := url.PathUnescape(string(ctx.Request.Header.RequestURI()[1:]))
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
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
