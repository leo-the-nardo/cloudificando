package main

import "net/http"

// EventPostUpdatedRequest is the gcp pub/sub push-based subscription body send by gcp
type EventPostUpdatedRequest struct {
	Message struct {
		Data        string `json:"data" binding:"required"`
		MessageId   string `json:"message_id" binding:"required"`
		PublishTime string `json:"publish_time" binding:"required"`
		OrderingKey string `json:"ordering_key"`
		Attributes  struct {
			EventType string `json:"eventType" binding:"required"` // "POST_CREATED" ,"POST_DELETED", "CONTENT_UPDATED", "META_UPDATED"
			Slug      string `json:"slug" binding:"required"`
		} `json:"attributes" binding:"required"`
	} `json:"message" binding:"required"`
	Subscription string `json:"subscription" binding:"required"`
}
type HardSyncRequest struct {
	Posts []Post `json:"posts" binding:"required"`
}

type RestError struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Error   string `json:"error"`
}

func BadRequestError(message string) *RestError {
	return &RestError{
		Message: message,
		Status:  http.StatusBadRequest,
		Error:   "Invalid Request",
	}
}

type Post struct {
	Title       string   `json:"title" dynamodbav:"title" binding:"required"`
	Tags        []string `json:"tags" dynamodbav:"tags" binding:"required"`
	CreatedAt   string   `json:"created_at" dynamodbav:"created_at" binding:"required"`
	Description string   `json:"description" dynamodbav:"description" binding:"required"`
	Slug        string   `json:"slug" dynamodbav:"slug" binding:"required"`
}

type TagWithCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// TagMetadata represents the metadata for a tag
type TagMetadata struct {
	PK   string `dynamodbav:"PK"`
	SK   string `dynamodbav:"SK"`
	Slug string `dynamodbav:"slug"`
}
type ListPosts struct {
	Items      []Post `json:"items"`
	NextCursor string `json:"nextCursor"`
}
