// handlers.go
package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
	"log"
	"log/slog"
	"sort"
	"strconv"
	"sync"
	"time"
)

func GetPostsHandler(db *dynamodb.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var errs []string
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "6")
		tag := c.DefaultQuery("tag", "")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			errs = append(errs, "Invalid page number")
		}
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 6 {
			errs = append(errs, "Invalid limit")
		}
		if len(errs) > 0 {
			c.AbortWithStatusJSON(400, gin.H{
				"error": errs,
			})
			return
		}

		var pk string
		if tag != "" {
			pk = fmt.Sprintf("TAG#%s", tag)
		} else {
			pk = "POST"
		}

		input := &dynamodb.QueryInput{
			TableName:              aws.String("BlogTable"),
			KeyConditionExpression: aws.String("PK = :pk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: pk},
			},
			Limit:            aws.Int32(int32(limit)),
			ScanIndexForward: aws.Bool(false), // Descending order
		}

		result, err := db.Query(ctx, input)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to query posts", err)
			c.AbortWithStatusJSON(500, gin.H{
				"error": "Failed to retrieve posts",
			})
			return
		}

		var posts []Post
		for _, item := range result.Items {
			var post Post
			err = attributevalue.UnmarshalMap(item, &post)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to unmarshal post", err)
				continue
			}
			posts = append(posts, post)
		}

		c.JSON(200, gin.H{
			"posts": posts,
			"page":  page,
		})
		slog.InfoContext(ctx, "Posts retrieved successfully")
	}
}

func GetTagsHandler(db *dynamodb.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Define the initial query to fetch all tag slugs
		queryInput := &dynamodb.QueryInput{
			Select:                 types.SelectSpecificAttributes,
			ProjectionExpression:   aws.String("slug"),
			TableName:              aws.String("BlogTable"),
			KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk_prefix)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":        &types.AttributeValueMemberS{Value: "TAG"},
				":sk_prefix": &types.AttributeValueMemberS{Value: "TAG#"},
			},
		}

		// Execute the query to fetch tag slugs
		result, err := db.Query(ctx, queryInput)
		if err != nil {
			log.Printf("Failed to query tags: %v", err)
			c.JSON(500, gin.H{"error": "Failed to retrieve tags"})
			return
		}

		// Unmarshal the tag slugs
		tagNames := make([]string, 0, len(result.Items))
		for _, item := range result.Items {
			var tagMetadata TagMetadata
			if err := attributevalue.UnmarshalMap(item, &tagMetadata); err != nil {
				log.Printf("Failed to unmarshal tag metadata: %v", err)
				continue
			}
			tagNames = append(tagNames, tagMetadata.Slug)
		}

		// Prepare for concurrent counting
		var wg sync.WaitGroup
		tagsWithCountsChan := make(chan TagWithCount, len(tagNames))
		errorsChan := make(chan error, len(tagNames))

		// Iterate over tags and count posts concurrently
		for _, tag := range tagNames {
			wg.Add(1)
			go func(tag string) {
				defer wg.Done()

				countInput := &dynamodb.QueryInput{
					TableName:              aws.String("BlogTable"),
					KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk_prefix)"),
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":pk":        &types.AttributeValueMemberS{Value: fmt.Sprintf("TAG#%s", tag)},
						":sk_prefix": &types.AttributeValueMemberS{Value: "CREATED_AT#"},
					},
					Select: types.SelectCount,
				}

				countResult, err := db.Query(ctx, countInput)
				if err != nil {
					log.Printf("Failed to count posts for tag '%s': %v", tag, err)
					errorsChan <- err
					return
				}

				tagsWithCountsChan <- TagWithCount{
					Tag:   tag,
					Count: int(countResult.Count),
				}
			}(tag)
		}

		// Wait for all goroutines to finish
		wg.Wait()
		close(tagsWithCountsChan)
		close(errorsChan)

		// Collect the results
		tagsWithCounts := make([]TagWithCount, 0, len(tagNames))
		for tagWithCount := range tagsWithCountsChan {
			tagsWithCounts = append(tagsWithCounts, tagWithCount)
		}
		//sort by count
		sort.Slice(tagsWithCounts, func(i, j int) bool {
			return tagsWithCounts[i].Count > tagsWithCounts[j].Count
		})

		// Optionally handle errors (e.g., log them or return partial results)
		if len(errorsChan) > 0 {
			// Here, we're just logging the number of errors
			log.Printf("Encountered %d errors while counting tags", len(errorsChan))
		}

		// Respond with the tags and their counts
		c.JSON(200, gin.H{
			"tags": tagsWithCounts,
		})
	}
}
func UpsertPostHandler(db *dynamodb.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
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

		// Insert Tag Metadata Conditionally
		for _, tag := range post.Tags {
			// Prepare TagMetadata item
			tagMetadataItem := map[string]types.AttributeValue{
				"PK":         &types.AttributeValueMemberS{Value: "TAG"},
				"SK":         &types.AttributeValueMemberS{Value: "TAG#" + tag},
				"slug":       &types.AttributeValueMemberS{Value: tag},
				"created_at": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
				"Type":       &types.AttributeValueMemberS{Value: "TAG"},
				// Add other metadata fields as needed
			}

			// Prepare PutItemInput with ConditionExpression
			putItemInput := &dynamodb.PutItemInput{
				TableName:           aws.String("BlogTable"),
				Item:                tagMetadataItem,
				ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
			}

			// Attempt to put the TagMetadata item
			_, err := db.PutItem(ctx, putItemInput)
			if err != nil {
				var cce *types.ConditionalCheckFailedException
				if errors.As(err, &cce) {
					// Tag metadata already exists; ignore the error
					slog.Info("Tag metadata already exists, ignoring:", "Tag", tag)
					continue
				}
				// For other errors, log and optionally handle them
				slog.Error("Failed to insert tag metadata", "Tag", tag, "Error", err)
				// Depending on requirements, you might choose to abort here
				// For this example, we'll continue
				continue
			}
			slog.Info("Inserted tag metadata", "Tag", tag)
		}

		// Begin Transaction for Upserting Post and Tag-Post Mappings
		transactItems := []types.TransactWriteItem{}

		// Upsert Post by Creation Date
		postByDateItem := map[string]types.AttributeValue{
			"PK":          &types.AttributeValueMemberS{Value: "POST"},
			"SK":          &types.AttributeValueMemberS{Value: fmt.Sprintf("CREATED_AT#%s#POST#%s", post.CreatedAt, post.Slug)},
			"title":       &types.AttributeValueMemberS{Value: post.Title},
			"tags":        &types.AttributeValueMemberL{Value: convertTagsToDynamoDB(post.Tags)},
			"created_at":  &types.AttributeValueMemberS{Value: post.CreatedAt},
			"description": &types.AttributeValueMemberS{Value: post.Description},
			"slug":        &types.AttributeValueMemberS{Value: post.Slug},
			"Type":        &types.AttributeValueMemberS{Value: "POST"},
		}

		transactItems = append(transactItems, types.TransactWriteItem{
			Put: &types.Put{
				TableName: aws.String("BlogTable"),
				Item:      postByDateItem,
			},
		})

		// Upsert Tag-Post Mapping
		for _, tag := range post.Tags {
			// Tag-Post Item
			tagPostItem := map[string]types.AttributeValue{
				"PK":          &types.AttributeValueMemberS{Value: fmt.Sprintf("TAG#%s", tag)},
				"SK":          &types.AttributeValueMemberS{Value: fmt.Sprintf("CREATED_AT#%s#POST#%s", post.CreatedAt, post.Slug)},
				"title":       &types.AttributeValueMemberS{Value: post.Title},
				"tags":        &types.AttributeValueMemberL{Value: convertTagsToDynamoDB(post.Tags)},
				"created_at":  &types.AttributeValueMemberS{Value: post.CreatedAt},
				"description": &types.AttributeValueMemberS{Value: post.Description},
				"slug":        &types.AttributeValueMemberS{Value: post.Slug},
				"Type":        &types.AttributeValueMemberS{Value: "TAG_POST"},
			}

			transactItems = append(transactItems, types.TransactWriteItem{
				Put: &types.Put{
					TableName: aws.String("BlogTable"),
					Item:      tagPostItem,
				},
			})
		}

		// Execute Transaction
		_, err := db.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
			TransactItems: transactItems,
		})

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
}

// Helper function to convert tags to DynamoDB AttributeValue List
func convertTagsToDynamoDB(tags []string) []types.AttributeValue {
	var avList []types.AttributeValue
	for _, tag := range tags {
		avList = append(avList, &types.AttributeValueMemberS{Value: tag})
	}
	return avList
}
