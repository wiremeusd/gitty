# gitty

Clone your GitHub repositories from the terminal — no browsing github.com for clone URLs, no juggling SSH keys between accounts.

Run `gitty`, fuzzy-search the list of your repositories, hit Enter, and it's cloned into the current directory. Works with multiple GitHub accounts: bind each folder to an account once, and gitty (and `git pull`/`git push` inside the cloned repos) will use the right credentials automatically.

```
$ gitty

  gitty · wiremeusd

  ❯ wiremeusd/api-gateway      ★ 12 · Go
    wiremeusd/payment-api      ★ 3  · TypeScript
    wiremeusd/dotfiles         ★ 1  · Shell

  ↑↓ select · / search · Enter clone · Esc quit

✔ Cloned into ./api-gateway
```

## Install

```sh
brew install wiremeusd/tap/gitty
```

macOS only for now (tokens are stored in the system Keychain).

## Quick start

```sh
# 1. Sign in — opens GitHub's device flow, no manual token handling
gitty auth login

# 2. Browse & clone
cd ~/projects
gitty
```

That's it for a single account. The repo list includes your own repositories, repos you collaborate on, and organization repos. The list is cached on disk, so it appears instantly and refreshes in the background.

## Multiple accounts

Bind a directory to an account once — the binding covers all subdirectories:

```sh
gitty auth login            # sign in with your work account too

cd ~/work
gitty use work-account      # ~/work → work-account

cd ~/personal
gitty use wiremeusd         # ~/personal → wiremeusd
```

From now on, running `gitty` anywhere under `~/work` lists the work account's repositories, and anywhere under `~/personal` — your personal ones. The deepest binding wins, so `~/work/clients` can point to a third account. No `.ssh/config` host aliases, no per-repo identity fiddling.

### How the right credentials reach git

gitty clones over HTTPS and registers itself as the [git credential helper](https://git-scm.com/docs/gitcredentials) in each cloned repository's local config. When you later run `git pull` or `git push` there, git asks gitty, and gitty supplies the token of the account the repo was cloned with. Your global git configuration is never touched.

## Commands

| Command | Description |
|---|---|
| `gitty` | Interactive repository list: `/` to search, Enter to clone, Esc to quit |
| `gitty auth login` | Add an account via GitHub OAuth Device Flow |
| `gitty auth list` | List added accounts |
| `gitty auth logout <account>` | Remove an account (token and folder bindings) |
| `gitty use <account>` | Bind the current directory (and subdirectories) to an account |

## Security

- Authentication uses GitHub's OAuth Device Flow — you never copy-paste tokens.
- Tokens live in the macOS Keychain, not in files. The config file (`~/.config/gitty/config.toml`) contains only account names and folder bindings.
- The token scope is `repo` (required to list and clone private repositories).

## Development

```sh
go test ./...

# device flow needs an OAuth app Client ID; for local runs pass it via env:
GITTY_CLIENT_ID=<your-oauth-app-client-id> go run ./cmd/gitty auth login
```

Releases are built by GoReleaser on tag push (see `.goreleaser.yaml`); the Homebrew formula in [`wiremeusd/homebrew-tap`](https://github.com/wiremeusd/homebrew-tap) is updated automatically.

## License

[MIT](LICENSE)
