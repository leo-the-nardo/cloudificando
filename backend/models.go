package main

import "net/http"

type UpsertPostRequest struct {
	Title       string   `json:"title" binding:"required"`
	Tags        []string `json:"tags" binding:"required"`
	Slug        string   `json:"slug" binding:"required"`
	Description string   `json:"description" binding:"required"`
	CreatedAt   string   `json:"created_at" binding:"required"`
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
	Title       string   `json:"title" dynamodbav:"title"`
	Tags        []string `json:"tags" dynamodbav:"tags"`
	CreatedAt   string   `json:"created_at" dynamodbav:"created_at"`
	Description string   `json:"description" dynamodbav:"description"`
	Slug        string   `json:"slug" dynamodbav:"slug"`
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
