# AWS Genius Lyrics

Golang app that seeds AWS DynamoDB with songs and lyrics for a single artist.

<p style="display: flex; align-items: center;">
  <img src="./docs/genius.png" width="400"/>
  <img src="./docs/aws-dynamodb.png" width="200"/>
</p>

## Running the App

1. [Install Go](https://go.dev/doc/install).
1. [Configure the AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) on your local workstation.
1. This app stores an artist's songs and lyrics into separate DynamoDB tables. Create tables for "songs" and "lyrics".

    - [AWS Console](https://aws.plainenglish.io/how-to-create-a-dynamodb-table-with-the-aws-console-92d2bfdd49b)
    - [AWS CLI](https://docs.aws.amazon.com/cli/latest/reference/dynamodb/create-table.html)

1. Clone the respository

    ```sh
    git clone git@github.com:jseashell/genius-lyrics-seed-service
    cd genius-lyrics-seed-service
    ```

1. Create an `.env` file and add the necessary values.

    ```sh
    cp .env.example .env
    ```

    - `GENIUS_ACCESS_TOKEN`: Visit [https://docs.genius.com/](https://docs.genius.com/).
    - `GENIUS_ARTIST_ID`: Search for your artist with the [API client](https://docs.genius.com/#search-h2).
    - `AWS_DYNAMODB_SONGS_TABLE_NAME`: Name of the table in which to save songs.
    - `AWS_DYNAMODB_LYRICS_TABLE_NAME`: Name of the table in which to save lyrics.
    - `SKIP_DB`: Skips database operations. Outputs song names and lyrics to stdout. Typically used for debugging and verification before incurring AWS costs.

1. Run the app

    ```go
    make run
    ```

## Project Structure

```text
.
├── cmd
│   └── main.go             # entry point
├── internal                # internal packages
│   ├── db                  # dynamodb operations
│   ├── genius              # genius.com integration
│   └── scraper             # web scraper
├── .env.example            # example environment file
├── .gitignore
├── go.mod                  # module dependencies
├── go.sum                  # dependency checksums
├── LICENSE
└── README.md
```

## 3rd party libraries

- [aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2) - AWS SDK for the Go programming language.
- [google/uuid](https://github.com/google/uuid) -  RFC-4122 compliant UUID module by Google.
- [dotenv](https://github.com/joho/godotenv) - A Go (golang) port of the Ruby [dotenv](https://github.com/bkeepers/dotenv) project.
- [colly](https://github.com/gocolly/colly) - Lightning Fast and Elegant Scraping Framework for Gophers.

## Disclaimer

Repository contributors are not responsible for costs incurred by AWS services.

## License

This software is distributed under the terms of the [MIT License](/LICENSE).