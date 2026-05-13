# Architecture Decision Records

Short ADRs (one paragraph each) capturing intentional design choices for this repo. Each entry has a **scope** sentence (what the rule covers) and a **why** sentence (the constraint that produced it). Reference these from PR reviews when someone proposes reverting the decision.

| # | Title |
|---|-------|
| [0001](0001-no-secret-handling-in-jwa-tobrew.md) | jwa-tobrew does not fetch or manage secrets |
| [0002](0002-ssh-only-for-tap-pushes.md) | Tap pushes use SSH, never an HTTPS-with-token flow |
| [0003](0003-no-hardcoded-secret-references.md) | No hardcoded UUIDs, item names, or secret references in source |
| [0004](0004-two-parsers-for-github-refs.md) | Two distinct parsers for GitHub references |
| [0005](0005-no-env-files-only-op-templates.md) | No `.env` files; only `.env.template` with `op://` references |
| [0006](0006-agentskills-boundary-contract.md) | Boundary contract between `homebrew-tap` and `agentskills` |
| [0007](0007-dotfiles-tap-contract.md) | Boundary contract between `homebrew-tap` and `dotfiles` |
| [0008](0008-one-binary-per-repo-and-homebrew-casks.md) | One binary per repo; ship via `homebrew_casks` |
| [0009](0009-cli-family-import-contract.md) | CLI family import contract |

The cross-project version of these rules lives at `~/dotfiles/docs/security-ground-rules.md`.
