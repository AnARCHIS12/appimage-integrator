# Contributing to Aurémi

Thank you for helping improve Aurémi.

## Before opening an issue

- Search existing issues to avoid duplicates.
- Confirm that the problem occurs with the latest release.
- Include your Linux distribution, desktop environment, CPU architecture, and AppImage type when relevant.
- Never include credentials, access tokens, or private filesystem data.

## Reporting a bug

Use the repository's bug-report form and include:

- the behavior you expected;
- what actually happened;
- exact reproduction steps;
- the Aurémi version;
- relevant command output or logs;
- a link to the affected AppImage's official download page, when possible.

For security problems, follow [SECURITY.md](SECURITY.md) instead of opening a public issue.

## Proposing a feature

Describe the user problem first, then the proposed behavior. Please explain how the feature would remain distribution-independent and safe for non-technical users.

## Development

Requirements:

- Go 1.23 or newer
- GNU Make

Run the tests:

```bash
make test
```

Build every release artifact:

```bash
make
```

Before submitting a pull request:

1. format Go files with `gofmt`;
2. run `go vet .`;
3. run `make test`;
4. keep changes focused;
5. update documentation and the changelog when behavior changes.

## Pull requests

Use a clear title and explain:

- what changed;
- why it changed;
- how it was tested;
- whether it affects compatibility or security.

By contributing, you agree that your contribution may be distributed under the project's MIT License.

