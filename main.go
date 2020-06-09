package main

import (
	"fmt"
	"log"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

// Index
func Index(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, "no se que estoy haciendo pero funciono \n")
}

// Hello
func Hello(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "hello, %s!\n", ctx.UserValue("hola"))
}

// MultiParams
func MultiParams(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "hi, %s, %s!\n", ctx.UserValue("name"), ctx.UserValue("word"))
}

// QueryArgs is used for uri query args test #11:
// if the req uri is /ping?name=foo, output: Pong! foo
// if the req uri is /piNg?name=foo, redirect to /ping, output: estoy en pign!
func QueryArgs(ctx *fasthttp.RequestCtx) {
	name := ctx.QueryArgs().Peek("name")
	fmt.Fprintf(ctx, "estoy en ping! %s\n", string(name))
}

func main() {
	router := fasthttprouter.New()
	router.GET("/", Index)
	//router.GET("/hello/:hola", Hello)
	//router.GET("/multi/:name/:word", MultiParams)
	//router.GET("/ping", QueryArgs)

	log.Fatal(fasthttp.ListenAndServe(":8080", router.Handler))
}
