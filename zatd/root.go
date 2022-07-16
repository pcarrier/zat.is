package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/vincent-petithory/dataurl"
	"net/url"
	"strings"
)

func serveRoot(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	match := base32At128Path.FindStringSubmatch(path)
	if match == nil {
		ctx.Error("invalid path", fasthttp.StatusNotFound)
	} else {
		id := strings.ToLower(match[1])
		target, err := lookUp(id, ctx)
		if err != nil {
			ctx.Error(fmt.Sprintf("failure: %s", err), fasthttp.StatusInternalServerError)
			return
		}
		if target == "" {
			ctx.Error("no link found", fasthttp.StatusNotFound)
			return
		}

		parsed, err := url.Parse(target)
		if err != nil {
			ctx.Error(fmt.Sprintf("unparseable link: %s", err), fasthttp.StatusExpectationFailed)
			return
		}

		if parsed.Scheme == "data" {
			decoded, err := dataurl.DecodeString(target)
			if err != nil {
				ctx.Error(fmt.Sprintf("invalid data URI: %s", err), fasthttp.StatusExpectationFailed)
				return
			}
			ctx.SetContentType(decoded.ContentType())
			ctx.SetBody(decoded.Data)
			return
		}

		ctx.Response.SetStatusCode(fasthttp.StatusMovedPermanently)
		ctx.Response.Header.SetCanonical(strLocation, []byte(target))
	}
}
