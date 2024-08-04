# Lyrics Database Seeder

Golang app that seeds AWS DynamoDB with lyrics (categorized by song) for a single artist.

<p style="display: flex; align-items: center;">
  <img src="./docs/@images/genius.png" width="400"/>
  <img src="./docs/@images/aws-dynamodb.png" width="200"/>
</p>

## Running the App

1. [Install Go](https://go.dev/doc/install).
1. [Configure the AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) on your local workstation.
1. This app stores an artist's songs and lyrics into separate DynamoDB tables. Create tables for "songs" and "lyrics".

    - [AWS Console](https://aws.plainenglish.io/how-to-create-a-dynamodb-table-with-the-aws-console-92d2bfdd49b)
    - [AWS CLI](https://docs.aws.amazon.com/cli/latest/reference/dynamodb/create-table.html)

1. Clone the respository

    ```sh
    git clone git@github.com:jseashell/lyrics-db-seeder.git
    cd lyrics-db-seeder
    ```

1. Create an `.env` file and add the necessary values.

    ```sh
    cp .env.example .env
    ```

    - `GENIUS_ACCESS_TOKEN`: Visit [https://docs.genius.com/](https://docs.genius.com/). Sign up for a developer account, create a new API client, and "Generate Token" for that client (do not use the client ID/secret).
    - `GENIUS_PRIMARY_ARTIST`: Name of the artist to collect.
    - `GENIUS_INCLUDE_FEATURED`: Indicates whether to scrape lyrics when GENIUS_PRIMARY_ARTIST is listed as a featured artist. This can greatly increase the amount of data to be processed.
    - `GENIUS_AFFILIATIONS`: List of affiliations to include in collections. Affiliations help the search engine, but searching will yield both explicit and implicit affiliations, or empty string. This can greatly increase the amount of data to be processed.
    - `LOG_LEVEL`: Log level. Supports "DEBUG", "INFO", "WARN", or "ERROR".
    - `AWS_DYNAMODB_SONGS_TABLE_NAME`: Name of the table in which to save songs.
    - `SKIP_DB`: Skips database operations. Typically used for debugging and verification before incurring AWS costs.

1. Run the app

    
    ```sh
    # clean, build
    make

    # clean, build, run
    make go
    ```

## Project Structure

```text
.
├── cmd
│   └── main.go             # entry point
├── docs                    # repo documentation
├── internal                # internal packages
│   ├── db                  # dynamodb operations
│   ├── genius              # genius.com integration
│   └── scraper             # web scraper
├── .env.example            # example environment file
├── .gitignore
├── go.mod                  # module dependencies
├── go.sum                  # dependency checksums
├── LICENSE
├── Makefile
└── README.md
```

## AWS

Performance will vary depending on your DynamoDB read/write capacity settings and your network connection.

## 3rd party libraries

- [aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2) - AWS SDK for the Go programming language.
- [google/uuid](https://github.com/google/uuid) -  RFC-4122 compliant UUID module by Google.
- [dotenv](https://github.com/joho/godotenv) - A Go (golang) port of the Ruby [dotenv](https://github.com/bkeepers/dotenv) project.
- [colly](https://github.com/gocolly/colly) - Lightning Fast and Elegant Scraping Framework for Gophers.

## Disclaimer

Repository contributors are not responsible for costs incurred by AWS services.

## License

This software is distributed under the terms of the [MIT License](/LICENSE).
