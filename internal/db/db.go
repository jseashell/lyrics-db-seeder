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
	"github.com/jseashell/genius-lyrics-seed-service/internal/genius"
)

func newClient() *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		slog.Error("Unable to load AWS SDK config.")
		panic(err)
	}
	return dynamodb.NewFromConfig(cfg)
}

func PutSong(song genius.Song) {
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
				slog.Error("Insert failed", "song", song)
			}
		}
	}
}
