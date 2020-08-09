# NextEngine Go SDK Example

NextEngine SDK for Go

https://github.com/takaaki-s/ne-sdk-go

## GettingStarted

Create your nextengine application to get client_id and client_secret

https://base.next-engine.org/apps/make/new/

in API tab `テスト環境設定` Redirect URI set to https://localhost:8080/callback

Then set the environment variable in your shell

```shell
$ export CLIENT_ID="YOUR_CLIENT_ID"
$ export CLIENT_SECRET="YOUR_CLIENT_SECRET"
```

and execute

```shell
$ docker-compose build
$ docker-compose up
```

Open http://localhost:8080 in your browser
