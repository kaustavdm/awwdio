# Awwdio

A Go implementation for a lightweight audio (and possibly video) conversation application using Twilio Programmable Video.

## Build

Install Go 1.22+ (preferably the latest Go version). Then, build and run:

```bash
go build -o bin/awwdio
./bin/awwdio
```

Or, directly run the source code, after ensuring the env vars mentioned in [Setup](#setup) is available:

```
go run main.go
```

## Setup

The application uses the following env vars:

- `TWILIO_ACCOUNT_SID`: Twilio account's SID that you can find in the Twilio console. Required.
- `TWILIO_API_KEY`: The API key SID for a new API key created from the Twilio console. Required.
- `TWILIO_API_SECRET`: The corresponding API secret for the API key. Required.
- `PORT`: Default `8080`. Optional.

An easy way to set this all up is to copy `sample.env` to `.env` and edit the values. Then, run `source .env` to load the env vars.

The following env vars can be configured additionally:

- `DEBUG`: Defaults to false. Optional. Set `DEBUG=true` to enable debug logging
- `JSON_LOGGER`: Defaults to text logger. Optional. Set `JSON_LOGGER=true` to enable JSON logging output.

## License

[MIT](LICENSE).
