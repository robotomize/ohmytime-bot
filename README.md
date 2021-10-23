# Ohmytime-bot
Get the local time in the selected location. Often you need to quickly see the local time in the city. This is a telegram bot that solves this problem. Just send him the name of the city or location and he will send you the local time

Try [Ohmytime-bot](https://t.me/ohmytime_bot)

## Usage
## <img src="https://github.com/robotomize/ohmytime-bot/raw/main/docs/video_2021-10-23_08-48-55.gif">

## Install

### Docker

```bash
docker pull robotomize/ohmytime-bot:latest
# or use docker compose
docker-compose up
```

### Local

```bash
# create index internal/index/assets
go generate ./...

# set env variables

# search index generated to ./internal/index/assets/cities.idx
export PATH_TO_INDEX=PATH_TO_INDEX_FOLDER
# your telegram token
export TELEGRAM_TOKEN=Your_TELEGRAM_TOKEN

go run ./cmd/ohmytime-bot
```

## Dependencies
* [Text search bleve](https://github.com/blevesearch/bleve)
* [Telegram Bot api](https://github.com/go-telegram-bot-api/telegram-bot-api)

## Technical
The entire search index has already been created and copied into a docker container. The index was created based on public data on locations and time zones. In the internal/index folder you can find the raw data and the search index generator.


## License
Cribe is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.

## Contact
Telegram: [@robotomize](https://t.me/robotomize)
