# Lessons

## govulncheck CI keeps failing on stdlib CVEs — and the setup-go `check-latest` trap

**Date:** 2026-06-05

**Context:** govulncheck flagged stdlib CVEs (net/textproto, crypto/x509) fixed in
a newer Go patch. CI used `actions/setup-go` with `go-version-file: go.mod`, and
go.mod pinned `go 1.25.10`. Every ~month a new stdlib CVE turned CI red until the
pin was manually bumped.

**Mistake 1 — wrong lever:** First tried changing the `go.mod` `go` directive to a
minor-only version to make `go-version-file` float. Two problems:
- Go tooling canonicalizes `go 1.25` back to `go 1.25.0` on build, defeating the float.
- It broke `TestGoVersionPinnedConsistently`, a deliberate project invariant that
  keeps go.mod ⇄ Dockerfile ⇄ READMEs in lockstep (go.mod is the single source of
  truth for the pinned toolchain). **Run the test suite before assuming a config
  change is safe — this project encodes invariants as tests.**

**Mistake 2 — incomplete fix:** Switched setup-go to `go-version: '1.25.x'` assuming
the `.x` wildcard installs the latest patch. CI STILL reported `go1.25.10`.
**Root cause:** `setup-go` defaults to `check-latest: false`. In that mode it reuses
ANY tool-cache version satisfying the spec, and GitHub-hosted runners pre-bake a Go
that already matches `1.25.x` (1.25.10). So setup-go silently reused the stale
baked-in toolchain instead of downloading the newest patch.

**Fix that works:** `go-version: '1.25.x'` **plus** `check-latest: true`. The latter
forces setup-go to resolve+download the newest matching patch from the manifest.

**Rules for next time:**
1. govulncheck stdlib findings are fixed by the *toolchain* version, never code.
   Verified locally: a newer Go on PATH overrides a lower pinned `go.mod` under
   `GOTOOLCHAIN=auto`, and govulncheck then reports clean.
2. To make CI track the latest Go patch automatically, you need BOTH a floating
   spec (`1.25.x` or `stable`) AND `check-latest: true`. The spec alone is a no-op
   when the runner's baked-in Go already satisfies it.
3. setup-go cannot be verified locally — the real proof is a CI run after push. Say so.
4. Remaining caveat: this floats the *scan/test* toolchain. The shipped Docker image
   still builds from the pinned `ARG GO_VERSION`, so a green govulncheck does not
   guarantee the shipped binary is patched. Automating the pin bump (scheduled PR
   updating go.mod + Dockerfile + READMEs) is the fuller solution if desired.