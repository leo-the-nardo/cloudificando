// handlers.go
package main

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
)

type BlogController struct {
	db         *dynamodb.Client
	migration  *Migration
	tableName  string
	repository *BlogRepository
}

func NewBlogController(db *dynamodb.Client, tableName string, migration *Migration, repository *BlogRepository) *BlogController {
	return &BlogController{
		db:         db,
		tableName:  tableName,
		migration:  migration,
		repository: repository,
	}
}

// GetPostsHandler handles fetching paginated posts from DynamoDB
func (bc *BlogController) GetPostsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	repository := bc.repository
	cursor := c.DefaultQuery("cursor", "")
	limitStr := c.DefaultQuery("limit", "6")
	tag := c.DefaultQuery("tag", "")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 6 {
		c.AbortWithStatusJSON(400, gin.H{
			"error": "Invalid limit",
		})
	}
	result, err := repository.GetPosts(ctx, limit, tag, cursor)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"error": "Failed to retrieve posts",
		})
		return
	}
	c.JSON(200, result)

	slog.InfoContext(ctx, "Posts retrieved successfully")
}

func (bc *BlogController) GetTagsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	repository := bc.repository
	result, err := repository.GetTags(ctx)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"error": "Unexpected error",
		})
		return
	}
	// Respond with the tags and their counts array outside object
	c.JSON(200, result)
}
func (bc *BlogController) PostsUpdatedGcpSubscriptionHandler(c *gin.Context) {
	ctx := c.Request.Context()
	repository := bc.repository
	var event EventPostUpdatedRequest

	// Bind JSON payload
	if err := c.ShouldBindJSON(&event); err != nil {
		slog.ErrorContext(ctx, "Invalid request payload", "Error", err)
		c.AbortWithStatusJSON(400, gin.H{
			"error": "Invalid request payload",
		})
		return
	}

	// Handle event types with a switch statement
	switch event.Message.Attributes.EventType {
	case "POST_CREATED", "META_UPDATED":
		var post Post
		dataDecoded, _ := base64.StdEncoding.DecodeString(event.Message.Data)
		if err := json.Unmarshal(dataDecoded, &post); err != nil {
			c.AbortWithStatusJSON(400, gin.H{
				"error": "Invalid post data",
			})
			return
		}
		if err := repository.UpsertPost(ctx, post); err != nil {
			slog.Error("Failed to upsert post", "Error", err)
			c.AbortWithStatusJSON(500, gin.H{
				"error": "Failed to upsert post",
			})
			return
		}
		c.JSON(200, gin.H{
			"message": "Post upserted successfully",
		})
		slog.Info("Post upserted successfully", "Slug", post.Slug)

	case "POST_DELETED":
		slug := event.Message.Attributes.Slug
		if slug == "" {
			c.AbortWithStatusJSON(400, gin.H{
				"error": "Missing slug attribute",
			})
			return
		}
		if err := repository.DeletePost(ctx, slug); err != nil {
			slog.Error("Failed to delete post", "Error", err)
			c.AbortWithStatusJSON(500, gin.H{
				"error": "Failed to delete post",
			})
			return
		}
		c.JSON(200, gin.H{
			"message": "Post deleted successfully",
		})
		slog.Info("Post deleted successfully", "Slug", slug)
	case "CONTENT_UPDATED":
		c.JSON(200, gin.H{
			"message": "ok",
		})
	default:
		c.AbortWithStatusJSON(400, gin.H{
			"error": "Invalid event type",
		})
		return
	}
	_ = repository.InvalidateCdnCache(ctx, "/blog/*")
}

func (bc *BlogController) UpsertPostHandler(c *gin.Context) {
	//db := bc.db
	ctx := c.Request.Context()
	repository := bc.repository
	var post Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"error": "Invalid request payload",
		})
		return
	}

	// Validate required fields
	if post.Slug == "" || post.Title == "" || post.CreatedAt == "" {
		c.AbortWithStatusJSON(400, gin.H{
			"error": "Missing required fields",
		})
		return
	}
	err := repository.UpsertPost(ctx, post)

	if err != nil {
		slog.Error("Failed to transact write items", "Error", err)
		c.AbortWithStatusJSON(500, gin.H{
			"error": "Failed to upsert post",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "Post upserted successfully",
	})
	slog.Info("Post upserted successfully", "Slug", post.Slug)
}

func (bc *BlogController) DeletePostHandler(c *gin.Context) {
	ctx := c.Request.Context()
	repository := bc.repository
	slug := c.Param("slug")
	if slug == "" {
		c.AbortWithStatusJSON(400, gin.H{
			"error": "Missing slug parameter",
		})
		return
	}
	err := repository.DeletePost(ctx, slug)
	if err != nil {
		slog.Error("Failed to delete post", "Error", err)
		c.AbortWithStatusJSON(500, gin.H{
			"error": "Failed to delete post",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "Post deleted successfully",
	})
	slog.Info("Post deleted successfully", "Slug", slug)
}

func (bc *BlogController) HardSyncHandler(c *gin.Context) {
	//migration := bc.migration
	repository := bc.repository
	ctx := c.Request.Context()
	var body HardSyncRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"error": err,
		})
		return
	}
	// TODO: this makes big downtime, refactor to delete all items instead of dropping table
	//err := migration.Down(ctx)
	//if err != nil {
	//	slog.ErrorContext(ctx, "Failed to run migrations", "Error", err)
	//	c.AbortWithStatusJSON(500, gin.H{
	//		"error": "Unexpected error",
	//	})
	//	return
	//}
	//err = migration.Up(ctx)
	//if err != nil {
	//	slog.ErrorContext(ctx, "Failed to run migrations", "Error", err)
	//	c.AbortWithStatusJSON(500, gin.H{
	//		"error": "Unexpected error",
	//	})
	//	return
	//}
	err := repository.UpsertPostsBatch(ctx, body.Posts)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to upsert posts", "Error", err)
		c.AbortWithStatusJSON(500, gin.H{
			"error": "Unexpected error",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "ok",
	})
	_ = repository.InvalidateCdnCache(ctx, "/blog/*")
}
