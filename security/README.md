# Repository security policy

Run `task ci:security` to check Go packages and tests with govulncheck, install
locked npm development dependencies without executing lifecycle scripts, and
then scan the repository. The reusable Trivy
module continues to fail on findings at every severity, including unfixed
vulnerabilities. Repository-specific decisions live in `trivy.yaml`.

The checker is pinned to `golang.org/x/vuln` v1.8.0 and run through `go run`
so it is built with the active Go toolchain rather than using an older installed
binary. Go must be available on PATH, as for the repository's Go tests.

The dependency-free `taskfiles/npm` test fixture does not produce a
`node_modules` directory after `npm ci`. Trivy may still print its informational
installation suggestion for that fixture; there are no dependency licenses to
collect there.

## Approved licenses

Apache-2.0, BlueOak-1.0.0, BSD-2-Clause, BSD-3-Clause, ISC and MIT are accepted
with their copyright, license and applicable notice requirements. Preserve
upstream license files and notices when redistributing dependency code.

MPL-2.0 is accepted for the unmodified HashiCorp dependencies
`aws-sdk-go-base/v2`, `go-cleanhttp`, `go-getter` and `go-version`. Preserve their
license notices and make the covered source available when distributing covered
software. Modifications to covered files must retain the applicable MPL terms.

OFL-1.1 is accepted for Chroma's embedded Liberation Mono font. Chroma v2.27.0's
`COPYING` contains both the MIT license and the font's copyright and full OFL
text. Preserve these notices, keep redistributed fonts under OFL, respect
reserved font names for modifications, and do not sell the fonts on their own.

The allowlist records acceptance; it does not remove these obligations. New
license identifiers still produce findings and require review.

## OpenPGP applicability

`store.openvex.json` records GO-2026-5932 as not affecting this repository's use
of `golang.org/x/crypto` v0.57.0. On 2026-10-10, `govulncheck ./...` reported zero
affected symbols or imported packages, and `go list -deps -test ./...` contained
no `golang.org/x/crypto/openpgp` packages. Other parts of x/crypto remain in use.

The statement is scoped to this repository and the exact dependency version.
Reassess it whenever Go imports or dependencies change. Do not reuse it to claim
that x/crypto's OpenPGP implementation itself is safe. If OpenPGP becomes used,
remove the statement and migrate to a maintained implementation.

To inspect the scan without table rendering, run:

```sh
task ci:security TRIVY_FS_EXTRA_ARGS='--format json --output /tmp/store-trivy.json'
```

Use `task trivy:version` to check the active scanner. The default Unix install
comes from Nixpkgs; its packaged version can lag upstream releases. Updating the
scanner does not itself resolve dependency vulnerabilities or license findings.
