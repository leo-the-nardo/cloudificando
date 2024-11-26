// main.go
package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	aws "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	ctx := context.Background()
	traceProvider, loggerProvider, err := NewOtelProviders(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer traceProvider.Shutdown(ctx)
	defer loggerProvider.Shutdown(ctx)

	// Logger setup
	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})
	otelHandler := otelslog.NewHandler(os.Getenv("PROD_DOMAIN"), otelslog.WithLoggerProvider(loggerProvider))
	slog.SetDefault(slog.New(slogmulti.Fanout(consoleHandler, otelHandler)))

	// Initialize DynamoDB client
	awsConfig, err := aws.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}
	otelaws.AppendMiddlewares(&awsConfig.APIOptions)
	db := dynamodb.NewFromConfig(awsConfig)

	slog.Info("Cold start")
	router := gin.New()
	router.Use(otelgin.Middleware(os.Getenv("PROD_DOMAIN")))

	// Register endpoints
	router.GET("/blog/posts", GetPostsHandler(db))
	router.GET("/blog/tags", GetTagsHandler(db))
	router.POST("/blog/posts", UpsertPostHandler(db))

	// Run the server
	router.Run()
}
