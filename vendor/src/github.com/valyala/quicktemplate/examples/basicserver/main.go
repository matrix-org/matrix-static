// The following line is needed for generating go code from templates
// with `go generate`.
// See https://blog.golang.org/generate for more info.
// Quicktemplate compiler (qtc) must be installed before running
// `go generate`:
//
//     go get -u github.com/valyala/quicktemplate/qtc
//
//go:generate qtc -dir=templates
package main

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/valyala/fasthttp"
	"github.com/valyala/quicktemplate/examples/basicserver/templates"
)

func main() {
	log.Printf("starting the server at http://localhost:8080 ...")
	err := fasthttp.ListenAndServe("localhost:8080", requestHandler)
	if err != nil {
		log.Fatalf("unexpected error in server: %s", err)
	}
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/":
		mainPageHandler(ctx)
	case "/table":
		tablePageHandler(ctx)
	default:
		errorPageHandler(ctx)
	}
	ctx.SetContentType("text/html; charset=utf-8")
}

func mainPageHandler(ctx *fasthttp.RequestCtx) {
	p := &templates.MainPage{
		CTX: ctx,
	}
	templates.WritePageTemplate(ctx, p)
}

func tablePageHandler(ctx *fasthttp.RequestCtx) {
	rowsCount := ctx.QueryArgs().GetUintOrZero("rowsCount")
	if rowsCount == 0 {
		rowsCount = 10
	}
	p := &templates.TablePage{
		Rows: generateRows(rowsCount),
	}
	templates.WritePageTemplate(ctx, p)
}

func errorPageHandler(ctx *fasthttp.RequestCtx) {
	p := &templates.ErrorPage{
		Path: ctx.Path(),
	}
	templates.WritePageTemplate(ctx, p)
	ctx.SetStatusCode(fasthttp.StatusBadRequest)
}

func generateRows(rowsCount int) []string {
	var rows []string
	for i := 0; i < rowsCount; i++ {
		r := fmt.Sprintf("row %d", i)
		if rand.Intn(20) == 0 {
			r = "bingo"
		}
		rows = append(rows, r)
	}
	return rows
}
