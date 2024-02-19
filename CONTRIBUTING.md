# Contribution Guide

## Development

### Server

```bash
go run cmd/main.go
```

#### GraphQL Schemas

Run this when you make changes to the GraphQL schema.

```bash
go get github.com/Yamashou/gqlgenc
cd internal/anilist
go run github.com/Yamashou/gqlgenc
cd ../..
go mod tidy
```

#### Tests

You should run tests individually.


Some tests require some setup before running.
- Create a dummy AniList account and grab the JWT associated with your session.
- Rename `test/jwt.json.example` to `test/jwt.json`
- If you wish to use more than one account for tests, fill both pairs (`jwt` and `username`) in the file with different data. 
Otherwise, if you want to use a single account, fill both pairs with the same data.
- Create a dummy MyAnimeList account and grab the JWT associated with your session.

```json
{
  "jwt": "your-jwt",
  "username": "your-username",
  "jwt2": "your-jwt",
  "username2": "your-username",
  "mal_jwt": "your-mal-jwt"
}
```

In some tests, you will see the use of `anilist.MockAnilistClientWrappers()`.
This is used to create an authenticated AniList client by using the data `test/jwt.json`.

```go
func Test(t *testing.T) {
	anilistClientWrapper1, anilistClientWrapper2, data := anilist.MockAnilistClientWrappers()
	// anilistClientWrapper1 -> Client from jwt, username
	// anilistClientWrapper2 -> Client from jwt2, username2
	// data -> test/jwt.json
}
```


Some tests related to media players require you to have those installed on your system.
They also require you to have a media file.


### Web

#### Development

```bash
cd seanime-web
```

```bash
npm install
```

```bash
npm run dev
```

- Go to `http://127.0.0.1:43210` to see the web interface.
- Notice that since the port is different from the server, you will not be able to authenticate with MyAnimeList.

#### Built

To run the web interface with the server like in a normal setup (i.e. accessible from `http://127.0.0.1:43211`), you should build the web interface.
This will create a `web` folder under `/seanime-web`.

You have two options:
- In `constants.go`, change the `DevelopmentWebBuild` to `true`, so the server will serve the web interface from the `seanime-web/web` folder.
- Or, move the `/seanime-web/web` folder to the root of the project and run the server.




## Contributing

Before opening a PR, you should use the Feature Request or Issue template to announce your idea and get feedback on
whether it's a good fit for the project.

Your contribution should be concise, have a clear scope, be testable, and be well documented. If you're not sure about
something, feel free to ask.

