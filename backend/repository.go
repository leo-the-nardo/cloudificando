package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log/slog"
	"sort"
	"sync"
)

type BlogRepository struct {
	Db        *dynamodb.Client
	TableName string
}

func NewBlogRepository(db *dynamodb.Client, tableName string) *BlogRepository {
	return &BlogRepository{
		Db:        db,
		TableName: tableName,
	}
}
func (r *BlogRepository) UpsertPost(ctx context.Context, post Post) error {
	db := r.Db
	tableName := r.TableName
	// Insert Tag Metadata Conditionally
	for _, tag := range post.Tags {
		// Prepare TagMetadata item
		tagMetadataItem := map[string]types.AttributeValue{
			"PK":   &types.AttributeValueMemberS{Value: "TAG"},
			"SK":   &types.AttributeValueMemberS{Value: "TAG#" + tag},
			"slug": &types.AttributeValueMemberS{Value: tag},
			"Type": &types.AttributeValueMemberS{Value: "TAG"},
			// Add other metadata fields as needed
		}

		// Prepare PutItemInput with ConditionExpression
		putItemInput := &dynamodb.PutItemInput{
			TableName:           &tableName,
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
			continue
		}
		slog.Info("Inserted tag metadata", "Tag", tag)
	}

	// Begin Transaction for Upserting Post and Tag-Post Mappings
	var transactItems []types.TransactWriteItem

	// Upsert Post by Creation Date
	postByDateItem := map[string]types.AttributeValue{
		"PK":          &types.AttributeValueMemberS{Value: "POST"},
		"SK":          &types.AttributeValueMemberS{Value: fmt.Sprintf("POST#%s", post.Slug)},
		"SK_LSI1":     &types.AttributeValueMemberS{Value: fmt.Sprintf("CREATED_AT#%s#POST#%s", post.CreatedAt, post.Slug)},
		"title":       &types.AttributeValueMemberS{Value: post.Title},
		"tags":        &types.AttributeValueMemberL{Value: stringSliceToDynamoDB(post.Tags)},
		"created_at":  &types.AttributeValueMemberS{Value: post.CreatedAt},
		"description": &types.AttributeValueMemberS{Value: post.Description},
		"slug":        &types.AttributeValueMemberS{Value: post.Slug},
		"Type":        &types.AttributeValueMemberS{Value: "POST"},
	}
	transactItems = append(transactItems, types.TransactWriteItem{
		Put: &types.Put{
			TableName: &tableName,
			Item:      postByDateItem,
		},
	})

	// Upsert Tag-Post Mapping
	for _, tag := range post.Tags {
		// Tag-Post Item
		tagPostItem := map[string]types.AttributeValue{
			"PK":          &types.AttributeValueMemberS{Value: fmt.Sprintf("TAG#%s", tag)},
			"SK":          &types.AttributeValueMemberS{Value: fmt.Sprintf("POST#%s", post.Slug)},
			"SK_LSI1":     &types.AttributeValueMemberS{Value: fmt.Sprintf("CREATED_AT#%s#POST#%s", post.CreatedAt, post.Slug)},
			"title":       &types.AttributeValueMemberS{Value: post.Title},
			"tags":        &types.AttributeValueMemberL{Value: stringSliceToDynamoDB(post.Tags)},
			"created_at":  &types.AttributeValueMemberS{Value: post.CreatedAt},
			"description": &types.AttributeValueMemberS{Value: post.Description},
			"slug":        &types.AttributeValueMemberS{Value: post.Slug},
			"Type":        &types.AttributeValueMemberS{Value: "TAG_POST"},
		}

		transactItems = append(transactItems, types.TransactWriteItem{
			Put: &types.Put{
				TableName: &tableName,
				Item:      tagPostItem,
			},
		})
	}

	// Execute Transaction
	_, err := db.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: transactItems,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *BlogRepository) UpsertPostsBatch(ctx context.Context, posts []Post) error {
	db := r.Db
	tableName := r.TableName

	var tagsSet = make(map[string]string) //  deduplicate tags
	for _, post := range posts {
		for _, tag := range post.Tags {
			tagsSet[tag] = tag
		}
	}
	// Insert Tag Metadata
	var tagBatchItems []types.WriteRequest
	for tag := range tagsSet {
		tagBatchItems = append(tagBatchItems, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"PK":   &types.AttributeValueMemberS{Value: "TAG"},
					"SK":   &types.AttributeValueMemberS{Value: "TAG#" + tag},
					"slug": &types.AttributeValueMemberS{Value: tag},
					"Type": &types.AttributeValueMemberS{Value: "TAG"},
				},
			},
		})
	}
	_, err := db.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			tableName: tagBatchItems,
		},
	})
	if err != nil {
		return err
	}
	// Upsert Posts and Tag-Post Mappings
	var transactItems []types.TransactWriteItem
	for _, post := range posts {
		// Upsert Post by Creation Date
		postByDateItem := map[string]types.AttributeValue{
			"PK":          &types.AttributeValueMemberS{Value: "POST"},
			"SK":          &types.AttributeValueMemberS{Value: fmt.Sprintf("POST#%s", post.Slug)},
			"SK_LSI1":     &types.AttributeValueMemberS{Value: fmt.Sprintf("CREATED_AT#%s#POST#%s", post.CreatedAt, post.Slug)},
			"title":       &types.AttributeValueMemberS{Value: post.Title},
			"tags":        &types.AttributeValueMemberL{Value: stringSliceToDynamoDB(post.Tags)},
			"created_at":  &types.AttributeValueMemberS{Value: post.CreatedAt},
			"description": &types.AttributeValueMemberS{Value: post.Description},
			"slug":        &types.AttributeValueMemberS{Value: post.Slug},
			"Type":        &types.AttributeValueMemberS{Value: "POST"},
		}
		transactItems = append(transactItems, types.TransactWriteItem{
			Put: &types.Put{
				TableName: &tableName,
				Item:      postByDateItem,
			},
		})

		// Upsert Tag-Post Mapping
		for _, tag := range post.Tags {
			// Tag-Post Item
			tagPostItem := map[string]types.AttributeValue{
				"PK":          &types.AttributeValueMemberS{Value: fmt.Sprintf("TAG#%s", tag)},
				"SK":          &types.AttributeValueMemberS{Value: fmt.Sprintf("POST#%s", post.Slug)},
				"SK_LSI1":     &types.AttributeValueMemberS{Value: fmt.Sprintf("CREATED_AT#%s#POST#%s", post.CreatedAt, post.Slug)},
				"title":       &types.AttributeValueMemberS{Value: post.Title},
				"tags":        &types.AttributeValueMemberL{Value: stringSliceToDynamoDB(post.Tags)},
				"created_at":  &types.AttributeValueMemberS{Value: post.CreatedAt},
				"description": &types.AttributeValueMemberS{Value: post.Description},
				"slug":        &types.AttributeValueMemberS{Value: post.Slug},
				"Type":        &types.AttributeValueMemberS{Value: "TAG_POST"},
			}

			transactItems = append(transactItems, types.TransactWriteItem{
				Put: &types.Put{
					TableName: &tableName,
					Item:      tagPostItem,
				},
			})
		}

		// Execute Transaction
		_, err := db.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
			TransactItems: transactItems,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
func (r *BlogRepository) GetPosts(ctx context.Context, limit int, tag, cursor string) (*ListPosts, error) {
	tableName := r.TableName
	db := r.Db

	var pk string
	if tag != "" {
		pk = fmt.Sprintf("TAG#%s", tag)
	} else {
		pk = "POST"
	}
	input := &dynamodb.QueryInput{
		TableName:              &tableName,
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: pk},
		},
		Limit:            aws.Int32(int32(limit)),
		IndexName:        aws.String("LSI1"),
		ScanIndexForward: aws.Bool(false),
	}
	if cursor != "" {
		startKey, err := decodeCursor(cursor)
		if err != nil {
			return nil, err
		}
		input.ExclusiveStartKey = startKey
	}
	result, err := db.Query(ctx, input)

	if err != nil {
		return nil, err
	}

	var posts []Post
	for _, item := range result.Items {
		var post Post
		err = attributevalue.UnmarshalMap(item, &post)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	listPostsResult := &ListPosts{
		Items: posts,
	}
	if result.LastEvaluatedKey != nil && len(result.LastEvaluatedKey) > 0 {
		nextCursor, err := encodeCursor(result.LastEvaluatedKey)
		if err != nil {
			return nil, err
		}
		listPostsResult.NextCursor = nextCursor
	}
	return listPostsResult, nil
}
func (r *BlogRepository) GetTags(ctx context.Context) (*[]TagWithCount, error) {
	db := r.Db
	tableName := r.TableName
	// Define the initial query to fetch all tag slugs
	queryInput := &dynamodb.QueryInput{
		Select:                 types.SelectSpecificAttributes,
		ProjectionExpression:   aws.String("slug"),
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk_prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":        &types.AttributeValueMemberS{Value: "TAG"},
			":sk_prefix": &types.AttributeValueMemberS{Value: "TAG#"},
		},
	}

	// Execute the query to fetch tag slugs
	result, err := db.Query(ctx, queryInput)
	if err != nil {
		return nil, err
	}

	// Unmarshal the tag slugs
	tagNames := make([]string, 0, len(result.Items))
	for _, item := range result.Items {
		var tagMetadata TagMetadata
		if err := attributevalue.UnmarshalMap(item, &tagMetadata); err != nil {
			return nil, err
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
				TableName:              &tableName,
				KeyConditionExpression: aws.String("PK = :pk"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("TAG#%s", tag)},
				},
				Select: types.SelectCount,
			}

			countResult, err := db.Query(ctx, countInput)
			if err != nil {
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
		for err = range errorsChan {
			slog.ErrorContext(ctx, "Failed to count posts for tag", "Error", err)
		}
		return nil, errors.New("failed to count posts for some tags")
	}
	return &tagsWithCounts, nil

}

func (r *BlogRepository) DeletePost(ctx context.Context, slug string) error {
	db := r.Db
	tableName := r.TableName
	// Fetch the post from the database to get the tags
	getItemInput := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "POST"},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("POST#%s", slug)},
		},
	}

	result, err := db.GetItem(ctx, getItemInput)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get post", "Slug", slug, "Error", err)
		return err
	}

	if result.Item == nil {
		slog.ErrorContext(ctx, "Post not found", "Slug", slug)
		return errors.New("post not found")
	}

	// Extract the tags from the item
	var post Post
	err = attributevalue.UnmarshalMap(result.Item, &post)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal post", "Slug", slug, "Error", err)
		return err
	}

	// Begin Transaction for deleting Post and Tag-Post Mappings
	var transactItems []types.TransactWriteItem

	// Delete Post item
	deletePostItem := types.TransactWriteItem{
		Delete: &types.Delete{
			TableName: aws.String(tableName),
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "POST"},
				"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("POST#%s", slug)},
			},
		},
	}

	transactItems = append(transactItems, deletePostItem)

	// Delete Tag-Post mappings
	for _, tag := range post.Tags {
		deleteMapping := types.TransactWriteItem{
			Delete: &types.Delete{
				TableName: aws.String(tableName),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("TAG#%s", tag)},
					"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("POST#%s", slug)},
				},
			},
		}
		transactItems = append(transactItems, deleteMapping)
	}

	// Execute Transaction
	_, err = db.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: transactItems,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to transact delete items", "Error", err)
		return err
	}
	return nil
}
func stringSliceToDynamoDB(slice []string) []types.AttributeValue {
	var avList []types.AttributeValue
	for _, tag := range slice {
		avList = append(avList, &types.AttributeValueMemberS{Value: tag})
	}
	return avList
}

// Cursor represents the pagination cursor with PK and SK as strings
type Cursor struct {
	PK     string `json:"PK"`
	SK     string `json:"SK"`
	SKLSI1 string `json:"SK_LSI1"`
}

// encodeCursor encodes the ExclusiveStartKey into a base64 string
func encodeCursor(key map[string]types.AttributeValue) (string, error) {
	slog.Info("key", "key", key)
	pkAttr, ok := key["PK"].(*types.AttributeValueMemberS)
	if !ok {
		return "", errors.New("invalid cursor: missing or invalid PK")
	}
	skAttr, ok := key["SK"].(*types.AttributeValueMemberS)
	if !ok {
		return "", errors.New("invalid cursor: missing or invalid SK")
	}
	sklsi1Attr, ok := key["SK_LSI1"].(*types.AttributeValueMemberS)
	if !ok {
		return "", errors.New("invalid cursor: missing or invalid SK_LSI1")
	}
	cursor := Cursor{
		PK:     pkAttr.Value,
		SK:     skAttr.Value,
		SKLSI1: sklsi1Attr.Value,
	}

	marshaled, err := json.Marshal(cursor)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(marshaled)
	return encoded, nil
}

// decodeCursor decodes the base64 encoded cursor into a map[string]types.AttributeValue
func decodeCursor(cursor string) (map[string]types.AttributeValue, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, errors.New("invalid cursor encoding")
	}

	var c Cursor
	if err := json.Unmarshal(decoded, &c); err != nil {
		slog.Error("Failed to unmarshal cursor", "Error", err)
		return nil, errors.New("invalid cursor format")
	}

	key := map[string]types.AttributeValue{
		"PK":      &types.AttributeValueMemberS{Value: c.PK},
		"SK":      &types.AttributeValueMemberS{Value: c.SK},
		"SK_LSI1": &types.AttributeValueMemberS{Value: c.SKLSI1},
	}

	return key, nil
}
