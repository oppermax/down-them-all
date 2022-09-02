# down-them-all

A very quick and very dirty CLI tool to download all of a user's tweets.

## Requirements

```shell
        export CONSUMER_KEY="***"
	export CONSUMER_SECRET="***"
	export ACCESS_TOKEN="***"
	export ACCESS_SECRET="***"
```

See Twitter's docs to figure out how to get the API keys.

## Run it 🚀

```shell
go run downloader/main.go download --user=foobar
```