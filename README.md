# gitty

**[gitty.pro](https://gitty.pro)** · Clone your GitHub repositories from the terminal — no browsing github.com for clone URLs, no juggling SSH keys between accounts.

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
brew install --cask wiremeusd/tap/gitty
```

Works on macOS and Linux (casks need Homebrew ≥ 4.5 on Linux). No Homebrew? Grab a binary:

```sh
# linux_amd64; also: linux_arm64, darwin_arm64, darwin_amd64
curl -L https://github.com/wiremeusd/gitty/releases/latest/download/gitty_linux_amd64.tar.gz | tar xz
sudo install gitty /usr/local/bin/
```

Tokens are stored in the system keyring: the Keychain on macOS, the Secret Service (GNOME Keyring or KWallet) on Linux.

> **Linux note:** on a desktop (GNOME, KDE) tokens go into the Secret Service keyring out of the box.

### Running on a VPS / headless server

Headless machines have no Secret Service keyring. There are two ways to run gitty there:

**1. `GITTY_TOKEN` — quickest, for CI, containers, and ephemeral boxes.** Instead of running `gitty auth login`, export a GitHub [personal access token](https://github.com/settings/tokens) (scope `repo`) as `GITTY_TOKEN`. It takes priority over the keyring for every account, so no daemon is needed:

```sh
export GITTY_TOKEN=ghp_xxx   # shell profile, or a systemd Environment= line
gitty                        # browse & clone; git push/pull work via the credential helper
```

**2. `pass` — encrypted at rest, for a server you log into interactively.** gitty stores the token in a
[`pass`](https://www.passwordstore.org/) store instead — encrypted at rest with
GPG. One-time setup:

```sh
sudo apt install pass          # or: dnf install pass, apk add pass, ...
gpg --quick-gen-key "you@example.com"
pass init you@example.com      # use the key id/email from the line above
```

Then `gitty auth login` as usual. gitty auto-detects `pass` when no keyring is
running; `git pull`/`git push` in cloned repos read the token the same way.
`gpg-agent` asks for your passphrase once and caches it (for the gpg-agent TTL),
so subsequent operations run without prompts — the same feel as the macOS
Keychain. This assumes you use gitty interactively over SSH.

To force a backend explicitly, set `GITTY_KEYRING_BACKEND` to `secret-service`
or `pass` (default: `auto`).

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
- Tokens live in the system keyring (macOS Keychain / Linux Secret Service), not in files. The config file (`~/.config/gitty/config.toml`) contains only account names and folder bindings.
- The token scope is `repo` (required to list and clone private repositories).

## Development

```sh
go test ./...

# device flow needs an OAuth app Client ID; for local runs pass it via env:
GITTY_CLIENT_ID=<your-oauth-app-client-id> go run ./cmd/gitty auth login
```

Releases are built by GoReleaser on tag push (see `.goreleaser.yaml`); the Homebrew cask in [`wiremeusd/homebrew-tap`](https://github.com/wiremeusd/homebrew-tap) is updated automatically.

## License

[MIT](LICENSE)
