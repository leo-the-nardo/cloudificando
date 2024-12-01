package main

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log/slog"
)

type Migration struct {
	Db        *dynamodb.Client
	TableName string
}

func NewMigration(db *dynamodb.Client, tableName string) *Migration {
	return &Migration{
		Db:        db,
		TableName: tableName,
	}
}

func (m *Migration) EnsureDbMigrations(ctx context.Context) error {
	//	ignore err if table already exists
	err := m.Up(ctx)
	if err != nil {
		var resourceInUseException *types.ResourceInUseException
		if !errors.As(err, &resourceInUseException) {
			slog.ErrorContext(ctx, "Failed to create table", "Error", err)
			return err
		}
	}
	slog.InfoContext(ctx, "Table created", "tableName", m.TableName)
	return nil
}

func (m *Migration) Up(ctx context.Context) error {
	tableName := m.TableName
	db := m.Db
	createTableInput := &dynamodb.CreateTableInput{
		//on demand default
		BillingMode: types.BillingModePayPerRequest,

		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("SK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("SK_LSI1"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("SK_LSI2"),
				AttributeType: types.ScalarAttributeTypeN,
			},
			{
				AttributeName: aws.String("SK_LSI3"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("SK_LSI4"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("SK_LSI5"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("PK"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("SK"),
				KeyType:       types.KeyTypeRange,
			},
		},
		LocalSecondaryIndexes: []types.LocalSecondaryIndex{
			{
				IndexName: aws.String("LSI1"),
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("PK"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("SK_LSI1"),
						KeyType:       types.KeyTypeRange,
					},
				},
			},
			{
				IndexName: aws.String("LSI2"),
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("PK"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("SK_LSI2"),
						KeyType:       types.KeyTypeRange,
					},
				},
			},
			{
				IndexName: aws.String("LSI3"),
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("PK"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("SK_LSI3"),
						KeyType:       types.KeyTypeRange,
					},
				},
			},
			{
				IndexName: aws.String("LSI4"),
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("PK"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("SK_LSI4"),
						KeyType:       types.KeyTypeRange,
					},
				},
			},
			{
				IndexName: aws.String("LSI5"),
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeKeysOnly,
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("PK"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("SK_LSI5"),
						KeyType:       types.KeyTypeRange,
					},
				},
			},
		},
	}

	_, err := db.CreateTable(ctx, createTableInput)
	if err != nil {
		return err
	}
	return nil
}

func (m *Migration) Down(ctx context.Context) error {
	tableName := m.TableName
	db := m.Db
	deleteTableInput := &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	}
	_, err := db.DeleteTable(ctx, deleteTableInput)
	if err != nil {
		// ignore if is not exists, otherwise return error
		var resourceNotFoundException *types.ResourceNotFoundException
		if !errors.As(err, &resourceNotFoundException) {
			slog.ErrorContext(ctx, "Failed to delete table", "Error", err)
			return err
		}
	}
	return nil
}
