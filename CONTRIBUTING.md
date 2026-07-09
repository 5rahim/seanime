# Contributing to Seanime

Contributions are welcome when they are focused, maintainable, and aligned with the project.

Before opening a pull request, take the time to understand the relevant part of the codebase, follow the existing conventions, and verify your changes locally.

Pull requests that are too large, poorly scoped, generated without understanding, or inconsistent with the project structure may be closed without review.

## Get approval

Open an issue (bug report/feature request) first and **get approval**. This is to avoid wasting your time on something that may not be accepted.
Implementation details MUST be discussed beforehand in the issue.

Before planning your contribution:

- Set up Seanime locally and make sure it runs.
- Read the surrounding code before changing it.
- Keep the change small and focused.
- Follow the existing structure and coding style.
- Test your changes locally.
- Be prepared to explain and maintain the code you submit.

**Do not submit large or speculative PRs without prior discussion.**

First-time contributors should start with small bug fixes, tests, or minor scoped changes.

## AI-Assisted Contributions

AI tools may be used, but you are responsible for everything you submit.

> [!IMPORTANT]
> If you can't understand or debug it, don't submit it. Do not expect maintainers to rewrite, restructure, or debug AI-generated changes for you.


If you used AI tools for research, code generation, refactoring, tests, debugging, or documentation, include an `AI Disclosure` section in your PR:

```markdown
## AI Disclosure

- Tool(s) used:
- What AI was used for:
- Relevant prompts or instructions:
- What you manually reviewed or changed:
- How you verified the change fits Seanime's architecture:
```

These AI guidelines are not meant to discourage AI use but to help ensure that contributors put more effort into their contributions.

## Pull Request Descriptions

Write the PR description **in your own words**.

Do not submit an AI-generated summary of the diff. Explain the reasoning in your own words:

* What problem does this solve?
* What changed?
* Why was this approach chosen?
* How was it tested?
* Are there risks, tradeoffs, or limitations?
* Was AI used?

## Coding Style

Seanime favors simple, pragmatic code.

* Match the style of the surrounding code.
* Prefer direct code over clever abstractions.
* Avoid unnecessary wrappers, interfaces, or indirection.
* Avoid unrelated formatting changes.
* Avoid broad file moves or reorganizations.
* Keep comments concise and useful. (AI-generated comments will not be accepted.)

## Generated Files

Do not manually edit generated files, including:

* `codegen/generated/`
* `seanime-web/src/api/generated/`
* `seanime-web/src/routeTree.gen.ts`

If handler signatures, routes, or returned structs change, update the routes in:

```text
internal/handlers/routes.go
```

Then run:

```bash
go generate ./codegen/main.go
```

## Development Workflow

See [DEVELOPMENT_AND_BUILD.md](DEVELOPMENT_AND_BUILD.md) for setup, backend, frontend, and build instructions.

> [!IMPORTANT]
> To avoid merge conflicts, always make changes against the most active branch! It's not always `main`.

Recommended workflow:

```bash
git remote add upstream https://github.com/5rahim/seanime.git
git checkout main
git pull upstream main
git checkout -b <feature-or-fix-name>
```

Before opening a PR:

```bash
git pull --rebase upstream main
```

Then push your branch and open a PR against `main`.

## Testing

All changes must be verified locally.

For Go changes, add relevant tests when appropriate. Use `{file}_test.go`; do not use names like `{file}_regression_test.go`.

Use shared test helpers instead of local stubs or ad hoc fakes:

* `internal/testmocks/NewFakePlatformBuilder()`
* `internal/testmocks/NewFakeMetadataProviderBuilder()`
* `internal/testmocks/NewBaseAnimeBuilder()`
* `internal/testmocks/NewBaseMangaBuilder()`

For tests involving library entries, use the helpers in:

```text
internal/library/anime/test_wrapper_test.go
```

Run the relevant tests before submitting:

```bash
go test ./path/to/package/...
```

## PR Checklist

Before opening a PR, confirm that:

* [ ] I set up Seanime locally.
* [ ] I understand the code I changed.
* [ ] My PR is small and focused.
* [ ] I followed the existing coding style.
* [ ] I avoided unrelated changes.
* [ ] I tested the change locally.
* [ ] I wrote the PR description myself.
* [ ] I disclosed any AI assistance.
* [ ] I can explain and maintain this change.
