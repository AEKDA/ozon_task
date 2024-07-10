package server

import (
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/AEKDA/ozon_task/internal/api/graph"
)

func New(resolver graph.ResolverRoot, host string, port uint32) *http.Server {

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(
		graph.Config{
			Resolvers:  resolver,
			Directives: graph.DirectiveRoot{Length: graph.LengthDirective},
		}))

	mux := http.NewServeMux()

	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", srv)

	return &http.Server{
		Handler: mux, Addr: fmt.Sprintf("%s:%d", host, port),
	}

}
