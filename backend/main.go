// main.go
package main

import (
	"context"
	aws "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/gin-gonic/gin"
	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"log"
	"log/slog"
	"os"
)

func main() {
	ctx := context.Background()
	// Initialize OpenTelemetry providers
	traceProvider, loggerProvider, err := NewOtelProviders(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer traceProvider.Shutdown(ctx)
	defer loggerProvider.Shutdown(ctx)
	// Initialize the logger
	consoleSlogHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})
	otelSlogHandler := otelslog.NewHandler(os.Getenv("PROD_DOMAIN"), otelslog.WithLoggerProvider(loggerProvider))
	slog.SetDefault(slog.New(slogmulti.Fanout(consoleSlogHandler, otelSlogHandler)))
	slog.Info("Cold start")
	// Initialize DynamoDB client
	awsConfig, err := aws.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}
	otelaws.AppendMiddlewares(&awsConfig.APIOptions)
	db := dynamodb.NewFromConfig(awsConfig)
	// Initialize the migration
	tableName := os.Getenv("AWS_DYNAMO_TABLE_NAME")
	migration := NewMigration(db, tableName)
	// Initialize the BlogRepository
	parameterStore := ssm.NewFromConfig(awsConfig)
	cdn := cloudfront.NewFromConfig(awsConfig)
	blogRepository, err := NewBlogRepository(ctx, db, tableName, cdn, parameterStore)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to initialize BlogRepository", "Error", err)
		log.Fatal(err)
	}
	// Initialize the BlogController
	blogController := NewBlogController(db, tableName, migration, blogRepository)
	router := gin.New()
	// Register Global middlewares
	router.Use(OtelGinMiddleware())
	router.Use(CdnCacheMiddleware())
	router.Use(CorsMiddleware())
	// Register endpoints
	router.GET("/blog/posts", blogController.GetPostsHandler)
	router.GET("/blog/tags", blogController.GetTagsHandler)
	router.PUT("/blog/posts", blogController.UpsertPostHandler)
	router.POST("/blog/events/posts-updated", GcpPubSubMiddleware(), blogController.UpsertPostHandler)
	router.DELETE("/blog/posts/:slug", blogController.DeletePostHandler)
	router.POST("/blog/posts/hardsync", blogController.HardSyncHandler)
	// Run the migrations
	err = migration.EnsureDbMigrations(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to run migrations", "Error", err)
		log.Fatal(err)
	}
	// Run the server
	router.Run()
}
