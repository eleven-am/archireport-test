package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/functionalfoundry/graphqlws"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"

	"github.com/eleven-am/enclave/ent"
	"github.com/eleven-am/enclave/internal/auth"
	gql "github.com/eleven-am/enclave/internal/graphql"
)

// Config captures runtime configuration for the enclave server.
type Config struct {
	DatabasePath string
}

// Server bundles the Echo HTTP server, ent client and GraphQL schema.
type Server struct {
	App           *echo.Echo
	Client        *ent.Client
	Schema        graphql.Schema
	Resolver      *gql.Resolver
	Subscriptions graphqlws.SubscriptionManager
}

// New constructs the server, initializes the database schema and GraphQL handler.
func New(ctx context.Context, cfg Config) (*Server, error) {
	if cfg.DatabasePath == "" {
		cfg.DatabasePath = "enclave.db"
	}
	dsn := fmt.Sprintf("file:%s?_fk=1", cfg.DatabasePath)
	client, err := ent.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed opening database: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Schema.Create(ctx); err != nil {
		return nil, fmt.Errorf("failed creating schema resources: %w", err)
	}

	schema, resolver, err := gql.NewSchema(client)
	if err != nil {
		return nil, fmt.Errorf("failed constructing graphql schema: %w", err)
	}

	subscriptionManager := graphqlws.NewSubscriptionManager(&schema)
	resolver.RegisterNotificationListener(func(ctx context.Context, userID int, n *ent.Notification) {
		if n == nil {
			return
		}
		for conn, subs := range subscriptionManager.Subscriptions() {
			uid, ok := conn.User().(int)
			if !ok || uid != userID {
				continue
			}
			for _, sub := range subs {
				if !sub.MatchesField("notifications") {
					continue
				}
				payload := graphqlws.DataMessagePayload{
					Data: map[string]interface{}{
						"notifications": n,
					},
				}
				sub.SendData(&payload)
			}
		}
	})

	graphHandler := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	wsHandler := graphqlws.NewHandler(graphqlws.HandlerConfig{
		SubscriptionManager: subscriptionManager,
		Authenticate: func(token string) (interface{}, error) {
			if token == "" {
				return nil, fmt.Errorf("missing auth token")
			}
			id, err := strconv.Atoi(token)
			if err != nil {
				return nil, fmt.Errorf("invalid auth token")
			}
			return id, nil
		},
	})

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, err := auth.UserIDFromRequest(c.Request())
			if err == nil {
				ctx := auth.ContextWithUserID(c.Request().Context(), userID)
				c.SetRequest(c.Request().WithContext(ctx))
			}
			return next(c)
		}
	})

	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.Match([]string{http.MethodGet, http.MethodPost}, "/graphql", echo.WrapHandler(graphHandler))
	e.GET("/graphql/ws", echo.WrapHandler(wsHandler))

	return &Server{
		App:           e,
		Client:        client,
		Schema:        schema,
		Resolver:      resolver,
		Subscriptions: subscriptionManager,
	}, nil
}

// Close shuts down the server resources.
func (s *Server) Close() error {
	if s.Client != nil {
		return s.Client.Close()
	}
	return nil
}
