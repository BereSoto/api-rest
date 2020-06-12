package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

type DomainRequest struct {
	Domain string `json:"domain"`
}

type ServerResponse struct {
	Address  string `json:"address"`
	SslGrade string `json:"ssl_grade"`
	Country  string `json:"country"`
	Owner    string `json:"owner"`
}

type DomainResponse struct {
	Servers          []ServerResponse `json:"servers"`
	ServersChanged   bool             `json:"servers_changed"`
	SslGrade         string           `json:"ssl_grade"`
	PreviousSslGrade string           `json:"previous_ssl_grade"`
	Logo             string           `json:"logo"`
	Title            string           `json:"title"`
	IsDown           bool             `json:"is_down"`
}

// Index
func Index(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, "no se que estoy haciendo pero funciono \n")
}

// Hello
func domainCreate(ctx *fasthttp.RequestCtx) {
	var postBody DomainRequest

	err := json.Unmarshal(ctx.PostBody(), &postBody)
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}
	response := &DomainResponse{
		Servers:          []ServerResponse{},
		ServersChanged:   true,
		SslGrade:         "a+",
		PreviousSslGrade: "b+",
		Logo:             "",
		Title:            postBody.Domain,
		IsDown:           true,
	}
	responseJSON, _ := json.Marshal(response)

	fmt.Fprintf(ctx, "the requested domain was %s.\n", postBody.Domain)
	fmt.Fprintf(ctx, string(responseJSON))
}

func main() {
	router := fasthttprouter.New()
	router.GET("/", Index)
	router.POST("/dominios", domainCreate)
	//router.GET("/multi/:name/:word", MultiParams)
	//router.GET("/ping", QueryArgs)

	log.Fatal(fasthttp.ListenAndServe(":8081", router.Handler))
}
