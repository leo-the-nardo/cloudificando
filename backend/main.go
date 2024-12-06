// main.go
package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	aws "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/gin-gonic/gin"
	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
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
	slog.InfoContext(ctx, "Allowed origins", "Origins", os.Getenv("ALLOWED_ORIGINS"))
	//router.Use(CorsMiddleware(strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","))) //todo: understand why duplicate cors headers are being sent
	router.Use(CdnCacheMiddleware())
	// Register endpoints
	router.GET("/blog/posts", blogController.GetPostsHandler)
	router.GET("/blog/tags", blogController.GetTagsHandler)
	router.PUT("/blog/posts", blogController.UpsertPostHandler)
	router.POST("/blog/events/posts-updated", GcpPubSubAuthMiddleware(), blogController.PostsUpdatedGcpSubscriptionHandler)
	router.DELETE("/blog/posts/:slug", blogController.DeletePostHandler)
	router.POST("/blog/hardsync", blogController.HardSyncHandler)

	router.Run()
}
