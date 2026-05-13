# 0009 - CLI family import contract

**Scope.** The personal CLI family is four separate binaries with process-level
integration only. `jwa-harden` owns secret/runtime hardening and local
signing/notarization preflight. `agentskills` owns skill discovery,
bootstrap, copy/symlink behavior, and harness skill links. `prehandover`
owns agent-harness stop hooks and repo-local handover validation.
`jwa-tobrew` owns tap publishing, project release scaffolding, and tap
metadata refresh. None of these CLIs may import sibling CLI Go packages.
Shared behavior must either live in the owning CLI and be called as a child
process, or stay duplicated until there is a strong case for a small,
versioned library that has no dependency on any CLI.

**Allowed runtime edges.** `jwa-tobrew` may execute `jwa-harden` for
secret-aware release wrappers and signing/hardening checks, execute
`agentskills` to materialize project skills, and suggest or execute
`prehandover init/install` when bootstrapping harness validation. The reverse
edges are forbidden: `jwa-harden`, `agentskills`, and `prehandover` must not
depend on `jwa-tobrew` to build, release, install, or operate. The tap repo is
not a library and must not be imported by any source repo; it is only the
generated distribution sink.

**Bootstrap order.** `homebrew-tap` exists first as a distribution repository.
`jwa-harden` is the foundation and may release itself with raw
`op run --env-file=.env.template -- goreleaser release --clean` because it
cannot depend on itself. After that, `agentskills`, `prehandover`, and
`jwa-tobrew` can use `jwa-harden run -- goreleaser release --clean`.
`agentskills` and `prehandover` must be releasable without `jwa-tobrew`; their
GoReleaser configs write their own `Casks/*.rb` into this tap. `jwa-tobrew`
comes last because it orchestrates project setup and may call the other CLIs.

**Why.** These tools are related, but they are not one product. Keeping source
imports acyclic makes each CLI independently buildable and releasable, avoids
bootstrap deadlocks, and keeps the tap as a mechanical distribution channel
instead of a shared framework. Process boundaries are explicit enough for the
workflow while preserving small repos and clean ownership.
