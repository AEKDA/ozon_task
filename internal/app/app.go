package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/AEKDA/ozon_task/internal/api/graph"
	"github.com/AEKDA/ozon_task/internal/database/psql"
	"github.com/AEKDA/ozon_task/internal/logger"
	"github.com/AEKDA/ozon_task/internal/repository/inmemory"
	"github.com/AEKDA/ozon_task/internal/repository/pgrepo"
	"github.com/AEKDA/ozon_task/internal/service"
	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

func Run() {

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		panic(err)
	}

	var repo interface {
		service.PostRepository
		service.CommentRepository
	}

	switch cfg.StorageType {
	case TypeInmemory:
		repo = inmemory.NewInMemoryDB()
	case TypePostgres:
		pgconn, err := psql.NewConnection(context.Background(), cfg.Database)
		if err != nil {
			panic(err)
		}
		repo = pgrepo.New(pgconn, log)
	default:
		panic("invalid storage type")
	}

	service := service.NewPostService(repo, repo)

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(
		graph.Config{
			Resolvers:  &graph.Resolver{PostService: service},
			Directives: graph.DirectiveRoot{Length: graph.LengthDirective},
		}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Info("starting the server", zap.String("host", cfg.App.Host), zap.Uint16("port", cfg.App.Port))
	log.Error(http.ListenAndServe(fmt.Sprintf(":%d", cfg.App.Port), nil).Error())
}
