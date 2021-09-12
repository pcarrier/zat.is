package main

import (
	"context"
	b32 "encoding/base32"
	"errors"
	f "fmt"
	"github.com/google/uuid"
	du "github.com/vincent-petithory/dataurl"
	p "github.com/jackc/pgx/v4"
	pp "github.com/jackc/pgx/v4/pgxpool"
	qr "github.com/skip2/go-qrcode"
	h "github.com/valyala/fasthttp"
	"image/color"
	"image/gif"
	"image/png"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	strLocation        = []byte("Location")
	base32At128Path, _ = regexp.Compile("/\\.([a-zA-Z2-7]{26})")
	schemeParms, _     = regexp.Compile("(?:(gif|png|txt)([0-9]*)[.+](?:([0-9a-fA-F]{3,8})[.+](?:([0-9a-fA-F]{3,8})[.+])?)?)?(.*)")
	mimeAndContent, _  = regexp.Compile("([^:]*):(.*)")
	pool               *pp.Pool
)

func serve(ctx *h.RequestCtx) {
	switch string(ctx.Host()) {
	case "zat.is":
		serveZat(ctx)
	case "l.zat.is":
		serveLink(ctx)
	case "qr.zat.is":
		serveQR(ctx)
	default:
		ctx.Error(f.Sprintf("unknown host %s", ctx.Host()), h.StatusNotFound)
	}
}

func serveZat(ctx *h.RequestCtx) {
	path := string(ctx.Path())
	match := base32At128Path.FindStringSubmatch(path)
	if match == nil {
		ctx.Error("invalid path", h.StatusNotFound)
	} else {
		id := strings.ToUpper(match[1])
		target, err := lookUp(id, ctx)
		if err != nil {
			ctx.Error(f.Sprintf("failure: %s", err), h.StatusInternalServerError)
			return
		}
		if target == "" {
			ctx.Error("no link found", h.StatusNotFound)
			return
		}

		parsed, err := url.Parse(target)
		if err != nil {
			ctx.Error(f.Sprint("unparseable link: %s", err), h.StatusExpectationFailed)
			return
		}

		if parsed.Scheme == "data" {
			decoded, err := du.DecodeString(target)
			if err != nil {
				ctx.Error(f.Sprint("invalid data URI: %s", err), h.StatusExpectationFailed)
				return
			}
			ctx.SetContentType(decoded.ContentType())
			ctx.SetBody(decoded.Data)
			return
		}

		ctx.Response.SetStatusCode(h.StatusMovedPermanently)
		ctx.Response.Header.SetCanonical(strLocation, []byte(target))
	}
}

func lookUp(path string, ctx context.Context) (target string, err error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return
	}
	defer conn.Release()
	err = conn.QueryRow(ctx, "SELECT url FROM record WHERE path = $1", path).Scan(&target)
	if err == p.ErrNoRows {
		err = nil
	}
	return
}

func insertUrlWhenAbsent(path string, url string, ctx context.Context) (err error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return
	}
	defer conn.Release()
	_, err = conn.Exec(ctx, "INSERT INTO record(path,url) VALUES ($1, $2) ON CONFLICT DO NOTHING", path, url)
	return
}

func shorten(path string) string {
	u := uuid.NewSHA1(uuid.NameSpaceURL, []byte(path))
	return b32.StdEncoding.WithPadding(b32.NoPadding).EncodeToString(u[:])
}

func parseHexColor(s string) (c color.RGBA, err error) {
	switch len(s) {
	case 8:
		_, err = f.Sscanf(s, "%02x%02x%02x%02x", &c.R, &c.G, &c.B, &c.A)
	case 6:
		c.A = 255
		_, err = f.Sscanf(s, "%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = f.Sscanf(s, "%01x%01x%01x%01x", &c.R, &c.G, &c.B, &c.A)
		c.R *= 0x11
		c.G *= 0x11
		c.B *= 0x11
		c.A *= 0x11
	case 3:
		c.A = 255
		_, err = f.Sscanf(s, "%01x%01x%01x", &c.R, &c.G, &c.B)
		c.R *= 0x11
		c.G *= 0x11
		c.B *= 0x11
	}
	return
}

func extractFromScheme(scheme string) (realScheme string, format string, size int, fg, bg color.RGBA, err error) {
	fg = color.RGBA{A: 255}
	bg = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	parts := schemeParms.FindStringSubmatch(scheme)
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
	realScheme = parts[5]
	return
}

func extractFromPath(ctx *h.RequestCtx) (target string, format string, size int, fg, bg color.RGBA, err error) {
	var s string

	s = string(ctx.Request.Header.RequestURI()[1:])

	var u *url.URL
	u, err = url.Parse(s)
	if err != nil {
		return
	}
	if u.Scheme == "" {
		err = errors.New("missing scheme")
		return
	}

	u.Scheme, format, size, fg, bg, err = extractFromScheme(u.Scheme)
	if err != nil {
		return
	}

	target = u.String()
	if l := len(target); l > 10*1024 {
		err = errors.New(f.Sprintf("too long (%d characters)", l))
		return
	}

	return
}

func serveLink(ctx *h.RequestCtx) {
	ctx.Response.Header.Set("Cache-Control", "max-age=31536000")
	target, _, _, _, _, err := extractFromPath(ctx)
	if err != nil {
		ctx.Error(f.Sprintf("invalid target: %s", err), h.StatusBadRequest)
		return
	}

	path := shorten(target)
	err = insertUrlWhenAbsent(path, target, ctx)
	if err != nil {
		ctx.Error(err.Error(), h.StatusInternalServerError)
		return
	}
	_, err = f.Fprintf(ctx, "https://zat.is/.%s", path)
	if err != nil {
		ctx.Error(err.Error(), h.StatusInternalServerError)
		return
	}
}

func serveQR(ctx *h.RequestCtx) {
	ctx.Response.Header.Set("Cache-Control", "max-age=31536000")
	target, format, size, fg, bg, err := extractFromPath(ctx)
	if err != nil {
		ctx.Error(f.Sprintf("invalid target: %s", err), h.StatusBadRequest)
		return
	}

	if format == "" {
		format = "txt"
	}

	if size == 0 {
		size = 1
	}

	if size > 100 {
		ctx.Error(f.Sprintf("too many pixels per block (%d)", size), h.StatusBadRequest)
		return
	}

	path := shorten(target)
	err = insertUrlWhenAbsent(path, target, ctx)
	if err != nil {
		ctx.Error(err.Error(), h.StatusInternalServerError)
		return
	}
	u := f.Sprintf("HTTPS://ZAT.IS/.%s", path)
	ctx.Response.Header.Set("Link", u)
	ctx.Response.Header.Set("To", target)

	var q *qr.QRCode
	q, err = qr.New(u, qr.Low)
	q.ForegroundColor = fg
	q.BackgroundColor = bg
	if err != nil {
		ctx.Error(err.Error(), h.StatusInternalServerError)
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
				ctx.Error(err.Error(), h.StatusInternalServerError)
			}
		}
	case "gif":
		ctx.SetContentType("image/gif")
		img := q.Image(-size)
		if err := gif.Encode(ctx, img, &gif.Options{NumColors: 2}); err != nil {
			ctx.Error(err.Error(), h.StatusInternalServerError)
		}
	case "png":
		ctx.SetContentType("image/png")
		img := q.Image(-size)
		if err := png.Encode(ctx, img); err != nil {
			ctx.Error(err.Error(), h.StatusInternalServerError)
			return
		}
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
