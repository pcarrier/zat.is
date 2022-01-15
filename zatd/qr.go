package main

import (
	"errors"
	"fmt"
	"github.com/skip2/go-qrcode"
	"github.com/valyala/fasthttp"
	"image/color"
	"image/gif"
	"image/png"
	"net/url"
	"strconv"
	"strings"
)

func serveQR(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Cache-Control", "max-age=31536000")
	target, format, size, fg, bg, err := extractFromPath(ctx)
	if err != nil {
		ctx.Error(fmt.Sprintf("invalid target: %s", err), fasthttp.StatusBadRequest)
		return
	}

	if format == "" {
		format = "txt"
	}

	if size == 0 {
		size = 1
	}

	if size > 100 {
		ctx.Error(fmt.Sprintf("too many pixels per block (%d)", size), fasthttp.StatusBadRequest)
		return
	}

	shortened := shorten(target)

	err = insertUrlWhenAbsent(shortened, target, ctx)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	u := fmt.Sprintf("HTTPS://ZAT.IS/.%s", shortened)
	ctx.Response.Header.Set("Link", u)
	ctx.Response.Header.Set("To", target)

	var q *qrcode.QRCode
	q, err = qrcode.New(u, qrcode.Low)
	q.ForegroundColor = fg
	q.BackgroundColor = bg
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
	switch format {
	case "txt":
		for _, line := range q.Bitmap() {
			ll := len(line) + 1
			l := make([]rune, ll)
			for i, b := range line {
				if b {
					l[i] = '█'
				} else {
					l[i] = ' '
				}
			}
			l[ll-1] = '\n'
			_, err := ctx.Write([]byte(string(l)))
			if err != nil {
				ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
			}
		}
	case "gif":
		ctx.SetContentType("image/gif")
		img := q.Image(-size)
		if err := gif.Encode(ctx, img, &gif.Options{NumColors: 2}); err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		}
	case "png":
		ctx.SetContentType("image/png")
		img := q.Image(-size)
		if err := png.Encode(ctx, img); err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
			return
		}
	}
}

func extractParms(s string) (format string, size int, fg, bg color.RGBA, err error) {
	fg = color.RGBA{A: 255}
	bg = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	parts := parms.FindStringSubmatch(s)
	if len(parts) == 0 {
		err = errors.New("could not find a format")
		return
	}

	format = parts[1]
	if parts[2] != "" {
		size, err = strconv.Atoi(parts[2])
		if err != nil {
			return
		}
	}
	if parts[3] != "" {
		fg, err = parseHexColor(parts[3])
		if err != nil {
			return
		}
		if parts[4] != "" {
			bg, err = parseHexColor(parts[4])
			if err != nil {
				return
			}
		}
	}
	return
}

func parseHexColor(s string) (c color.RGBA, err error) {
	switch len(s) {
	case 8:
		_, err = fmt.Sscanf(s, "%02x%02x%02x%02x", &c.R, &c.G, &c.B, &c.A)
	case 6:
		c.A = 255
		_, err = fmt.Sscanf(s, "%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "%01x%01x%01x%01x", &c.R, &c.G, &c.B, &c.A)
		c.R *= 0x11
		c.G *= 0x11
		c.B *= 0x11
		c.A *= 0x11
	case 3:
		c.A = 255
		_, err = fmt.Sscanf(s, "%01x%01x%01x", &c.R, &c.G, &c.B)
		c.R *= 0x11
		c.G *= 0x11
		c.B *= 0x11
	}
	return
}

func extractFromPath(ctx *fasthttp.RequestCtx) (target string, format string, size int, fg, bg color.RGBA, err error) {
	path := ctx.Request.Header.RequestURI()
	total := string(path[1:])
	parts := strings.SplitN(total, "~", 2)

	if len(parts) != 2 {
		err = errors.New(fmt.Sprintf("expected /format~URL, got %s", string(path)))
		return
	}

	format, size, fg, bg, err = extractParms(parts[0])
	if err != nil {
		return
	}

	target = strings.ReplaceAll(parts[1], "%23", "#")

	var u *url.URL
	u, err = url.Parse(target)
	if err != nil {
		return
	}
	if u.Scheme == "" {
		err = errors.New("missing scheme")
		return
	}
	if l := len(target); l > 10*1024 {
		err = errors.New(fmt.Sprintf("URL too long (%d characters)", l))
		return
	}

	return
}
