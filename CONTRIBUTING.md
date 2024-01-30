# Contribution Guide

## Development

### Server

```bash
go mod tidy
```

```go
// cmd/main.go

package main
//...
var development = true
//...
```

Setting `development` to `true` will not launch a Fiber web server since we will be using a Next.js dev server.
Set it to `false` to launch before deploying.

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

- Some tests require some setup before running.
- Some tests require a valid AniList JWT in `test/jwt.json`
- Do not run tests all at once.


### Web

```bash
cd seanime-web
```

```bash
npm install
```

```bash
npm run dev
```

```bash
go run cmd/main.go
```

## Contributing

Before opening a PR, you should use the Feature Request or Issue template to announce your idea and get feedback on
whether it's a good fit for the project.

Your contribution should be concise, have a clear scope, be testable, and be well documented. If you're not sure about
something, feel free to ask.

