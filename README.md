# homebrew-tap

Personal [Homebrew](https://brew.sh) tap for casks and formulae.

## Setup

```bash
brew tap jwa91/tap
```

## Casks

| Cask | Description |
|------|-------------|
| [trnscrb](https://github.com/jwa91/trnscrb) | Drop audio, PDFs, and images onto your menu bar to get markdown back |

## Install

```bash
brew install --cask jwa91/tap/trnscrb
```

## Updating this tap

After pushing changes to this repo:

```bash
brew update
brew upgrade --cask trnscrb
```

Or re-tap to get the latest formulae/casks:

```bash
brew untap jwa91/tap && brew tap jwa91/tap
```
