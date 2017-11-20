package main

import (
	"net/http"
	"github.com/dpineda64/graphql-gateway/graphql"
	"github.com/graphql-go/handler"
)
func main(){

	services := new(graphql.Helper)
	schema := services.BuildSchema()

	h := handler.New(&handler.Config{
		Schema: &schema,
	})

	http.Handle("/graphql", h)
	http.ListenAndServe(":8087", nil)
}
