Go SDK for Satori RTM
============================================================

[![GitHub tag](https://img.shields.io/github/tag/satori-com/satori-rtm-sdk-go.svg)](https://github.com/satori-com/satori-rtm-sdk-go/tags)
[![GoDoc Widget]][GoDoc]

[<img src="https://cdn.satori.com/assets/utilities/Satori_Landscape_Logo_LightBckgnd.png" height="150">](https://satori.com)

Use the Go SDK for the Satori platform to create server-based applications that use the RTM to publish and subscribe.  
Follow https://www.satori.com/ to meet with Satori platform.

## Go SDK Installation

Use a `go-get` tool or any go-compatible package manager to download the SDK:
```
go get github.com/satori-com/satori-rtm-sdk-go/rtm
```

Import the SDK as a part of your application:
```
import "github.com/satori-com/satori-rtm-sdk-go/rtm"
```

## Documentation and Examples

You can view the latest Go SDK documentation here:
https://godoc.org/github.com/satori-com/satori-rtm-sdk-go/rtm

Documentation generates automatically during 1 hour after pushing to master.  
Read more about the GODOC service: https://godoc.org/-/about

## Logging and Debugging

Go SDK uses STDOUT/STDERR to output information. To enable `debug mode` 
set DEBUG_SATORI_SDK environment variable to `true`:
```bash
$ DEBUG_SATORI_SDK=true go run <your_app.go>
```
or 
```bash
$ export DEBUG_SATORI_SDK=true
$ go run <your_app.go>
```

After enabling `debug mode` you will get all requests to RTM and all responses in raw. Example:
```bash
$ DEBUG_SATORI_SDK=true go run <your_app.go>
[info] 2017/04/18 15:28:38.8129 Creating new RTM object
[info] 2017/04/18 15:28:38.8131 Client: Enter Connecting
[info] 2017/04/18 15:28:38.8131 Connecting to wss://<endpoint>.satori.com
[info] 2017/04/18 15:28:39.5492 Auth: Starting authentication
[debg] 2017/04/18 15:28:39.5493 send> {"action":"auth/handshake","body":{"method":"role_secret","data":{"role":"<role>"}},"id":1}
[debg] 2017/04/18 15:28:39.7246 recv< {"action":"auth/handshake/ok","body":{"data":{"nonce":"<nonce>"}},"id":1}
[debg] 2017/04/18 15:28:39.7247 Auth: Handshake response: {"action":"auth/handshake/ok","body":{"data":{"nonce":"<nonce>"}},"id":1}
[debg] 2017/04/18 15:28:39.7247 Auth: Got nonce. Trying to authenticate
[debg] 2017/04/18 15:28:39.7247 send> {"action":"auth/authenticate","body":{"method":"role_secret","credentials":{"hash":"<generated_hash>"}},"id":2}
[debg] 2017/04/18 15:28:39.8958 recv< {"action":"auth/authenticate/ok","body":{},"id":2}
[info] 2017/04/18 15:28:39.8959 Auth: Succesfully authenticated
[info] 2017/04/18 15:28:39.8959 Client: Enter Connected
[debg] 2017/04/18 15:28:39.8960 send> {"action":"rtm/write","body":{"channel":"channel-name","message":1},"id":3}
[debg] 2017/04/18 15:28:40.0750 recv< {"action":"rtm/write/ok","body":{"position":"1492522119:0"},"id":3}
```

## Tests and Coverage Report

Tests require an active RTM to be available. The tests require `credentials.json` to be populated with the RTM properties.

The `credentials.json` file must include the following key-value pairs:
```json
{
  "endpoint": "wss://<SATORI_HOST>/",
  "appkey": "<APP_KEY>",
  "auth_role_name": "<ROLE_NAME>",
  "auth_role_secret_key": "<ROLE_SECRET_KEY>",
  "auth_restricted_channel": "<CHANNEL_NAME>"
}
```

- `endpoint` is your customer-specific DNS name for RTM access.
- `appkey` is your application key.
- `auth_role_name` is a role name that permits to publish / subscribe to `auth_restricted_channel`. Must be not `default`.
- `auth_role_secret_key` is a secret key for `auth_role_name`.
- `auth_restricted_channel` is a channel with subscribe and publish access for `auth_role_name` role only.

You must use [DevPortal](https://developer.satori.com/) to create role and set channel permissions.

After setting up `credentials.json`, run SDK tests with the following commands:
```
$ go get github.com/satori-com/satori-sdk-go/rtm_client
$ CREDENTIALS=/full/path/to/credentials.json go test ./src/github.com/satori-com/satori-sdk-go/...
```

Use the `-cover` flag to get Coverage report. The `-coverprofile` flag produces debug profile file that
allows to analyse untested parts of SDK.

[GoDoc]: https://godoc.org/github.com/satori-com/satori-rtm-sdk-go/rtm
[GoDoc Widget]: https://godoc.org/github.com/satori-com/satori-rtm-sdk-go/rtm?status.svg
[logo]: https://cdn.satori.com/assets/utilities/Satori_Landscape_Logo_LightBckgnd.png "Satori"
