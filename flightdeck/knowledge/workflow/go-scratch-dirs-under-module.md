# ⚠ Go scratch dirs under the module are included in ./...
SUMMARY: Ordinary scratch directories such as `tmp_debug/` are still Go packages under this module, so stale imports or broken probes can make `go test ./...` fail.
READ WHEN: adding or keeping temporary Go probes/debug programs under the repo; `go test ./...` fails in a tmp/debug directory; renaming the module path or internal imports

---

Go's `./...` package pattern walks ordinary subdirectories under the module,
regardless of whether those paths are tracked by git. A directory named
`tmp_debug/` is therefore still compiled by `go test ./...` when it contains
`.go` files.

For disposable Go probes, prefer a path Go skips by convention, such as a
directory whose name starts with `_` or `.`, or put the probe under `testdata/`
when it is fixture material. If a scratch directory must stay under a normal
name, keep its imports current with the module path and expect it to participate
in full-package verification.
