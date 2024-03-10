# Contribution Guide

1. Fork it
2. Create your feature branch (`git checkout -b new-feature`)
3. Commit your changes (`git commit -am 'Added a feature'`)
4. Push to the branch (`git push origin new-feature`)
5. Create new Pull Request

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

`WIP`

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

