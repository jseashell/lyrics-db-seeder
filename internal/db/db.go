// Copyright 2024 John Schellinger.
// Use of this file is governed by the MIT license that can
// be found in the LICENSE.txt file in the project root.

// Package `db` integrates with AWS DynamoDB.
package db

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/jseashell/genius-lyrics-seed-service/internal/scraper"
)

func PutSong(song scraper.ScrapedSong) {
	dbClient := newClient()
	skipDb, _ := strconv.ParseBool(os.Getenv("SKIP_DB"))
	songsTableName := os.Getenv("AWS_DYNAMODB_SONGS_TABLE_NAME")

	av, _ := attributevalue.MarshalMap(song)

	if !skipDb {
		_, err := dbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(songsTableName),
		})

		if err != nil {
			if t := new(types.ConditionalCheckFailedException); !errors.As(err, &t) {
				slog.Warn("Insert failed", "song", song, "error", err)
			}
		} else {
			slog.Info("Insert success", "song", song)
		}
	}
}

func newClient() *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		slog.Error("Unable to load AWS SDK config.", "error", err)
		panic(err)
	}
	return dynamodb.NewFromConfig(cfg)
}
