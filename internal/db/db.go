package db

import (
	"context"
	"errors"
	"fmt"
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
	ce := "attribute_not_exists(ID)"

	if !skipDb {
		_, err := dbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
			Item:                av,
			TableName:           aws.String(songsTableName),
			ConditionExpression: &ce,
		})

		if err != nil {
			if t := new(types.ConditionalCheckFailedException); !errors.As(err, &t) {
				slog.Error("Failed song.", song.Title, err.Error())
			}
		}
	} else {
		slog.Warn(fmt.Sprintf("Skipping song insert \"%s\".", song.Title))
	}
}

func PutLyric(lyric genius.Lyric) {
	dbClient := newClient()
	skipDb, _ := strconv.ParseBool(os.Getenv("SKIP_DB"))
	lyricsTableName := os.Getenv("AWS_DYNAMODB_LYRICS_TABLE_NAME")

	if !skipDb {
		av, _ := attributevalue.MarshalMap(lyric)
		ce := "attribute_not_exists(ID)"

		_, err := dbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
			Item:                av,
			TableName:           aws.String(lyricsTableName),
			ConditionExpression: &ce,
		})

		if err != nil {
			if t := new(types.ConditionalCheckFailedException); !errors.As(err, &t) {
				slog.Error("Failed lyric.", "error", err.Error())
			}
		}
	} else {
		slog.Warn(fmt.Sprintf("Skipping lyric insert \"%s\".", lyric.Value))
	}
}
