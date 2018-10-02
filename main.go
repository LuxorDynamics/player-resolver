package main

import (
	"encoding/json"
	"github.com/buaazp/fasthttprouter"
	"github.com/luxordynamics/player-resolver/internal/cassandra"
	"github.com/luxordynamics/player-resolver/internal/mojang"
	"github.com/valyala/fasthttp"
	"log"
	"strings"
)

var api = mojang.NewApi()
var session cassandra.Session

func main() {

	session, err := cassandra.New()

	if err != nil {
		log.Fatal(err)
	}

	defer session.Close()

	router := fasthttprouter.New()
	router.GET("/uuid/:name", HandleUuidRequest)
	router.GET("/name/:uuid", HandleNameRequest)
	fasthttp.ListenAndServe(":8080", router.Handler)
}

// Handles requests for resolving names to UUIDs
func HandleUuidRequest(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)

	name := ctx.UserValue("name").(string)

	if !mojang.ValidUserNameRegex.MatchString(name) {
		log.Println("Given name is not valid. (" + name + ")")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(`{"error": "NameNotValidException"}`)
		return
	}

	// TODO: check if uuid is already in database

	mapping, err := api.UuidFromName(name)

	if err != nil {
		log.Fatal(err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(`{"error": "MojangRequestException"}`)
		return
	}

	resp, err := json.Marshal(mapping)

	if err != nil {
		log.Fatal(err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(`{"error": "ProcessFailedException"}`)
		return
	}

	ctx.SetBody(resp)
}

// Handles requests for resolving UUIDs to names
func HandleNameRequest(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	uuid := ctx.UserValue("uuid").(string)

	if mojang.ValidLongRegex.MatchString(uuid) {
		uuid = strings.Replace(uuid, "-", "", -1)
	} else if !mojang.ValidShortUuidRegex.MatchString(uuid) {
		handleError(ctx, `{"error": "MalformedUuidException"}`)
		return
	}

	exists, err := session.UuidEntryExists(uuid)

	if err != nil {
		handleError(ctx, `{"error": "InternalServiceException"}`)
		return
	}

	var data *mojang.PlayerNameMapping

	if exists {
		data, err = retrieveByUuid(uuid)
	} else {
		data, err = api.NameFromUuid(uuid)
	}

	if err != nil {
		handleError(ctx, `{"error": "InternalServiceException"}`)
		return
	}

	resp, err := json.Marshal(data)

	if err != nil {
		log.Fatal(err)
		handleError(ctx, `{"error": "InternalServiceException"}`)
		return
	}

	ctx.SetBody(resp)
	// TODO: check if uuid is already in database
}

func retrieveByUuid(uuid string) (mapping *mojang.PlayerNameMapping, err error) {
	entry, err := session.EntryByUuid(uuid)

	if err != nil {
		return nil, err
	}

	// TODO: check if last update was x days ago

	return &entry.Mapping, nil
}

func handleError(ctx *fasthttp.RequestCtx, body string) {
	ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	ctx.SetBodyString(body)
}
