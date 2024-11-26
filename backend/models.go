// models.go
package main

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
