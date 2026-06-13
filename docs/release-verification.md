# Release Verification

Official releases are published as signed annotated Git tags. The canonical source release artifact is the repository content reachable from the signed tag. GitHub Releases for signed `v*` tags also attach a CycloneDX SBOM and the SBOM's GitHub/Sigstore attestation bundle as distribution metadata; these assets do not replace signed-tag verification. Tags in the `v0.x` range and SemVer prerelease tags are GitHub prereleases and are not marked latest.

## Verify A Release Tag

Git verifies SSH signatures against an allowed signers file. CI uses the checked-in signer snapshot at `.github/allowed_signers`. Because that file is read from the tagged source tree, consumers should cross-check it against the maintainer's current public GitHub SSH keys before relying on it.

If this is the first time verifying this project's SSH-signed tags, create a project-specific allowed signers file from the maintainer's public GitHub SSH keys:

```sh
mkdir -p ~/.config/git
curl -fsSL https://github.com/the-sarge.keys |
  awk 'NF { print "the-sarge@the-sarge.com " $0 }' \
  > ~/.config/git/cpace-allowed-signers
```

Review the key fingerprints:

```sh
ssh-keygen -lf ~/.config/git/cpace-allowed-signers
```

If you have a checkout of the release source, verify that every checked-in signer key blob still appears in the freshly fetched maintainer keys. Extra fetched GitHub account keys are not trusted automatically; investigate them before adding them to an allowed signers file:

```sh
awk '{ print $2 " " $3 }' .github/allowed_signers | sort -u > /tmp/cpace-checked-in-keys
awk '{ print $2 " " $3 }' ~/.config/git/cpace-allowed-signers | sort -u > /tmp/cpace-github-keys
missing="$(comm -23 /tmp/cpace-checked-in-keys /tmp/cpace-github-keys)"
test -z "$missing" || { printf 'checked-in signer keys missing from GitHub keys:\n%s\n' "$missing" >&2; exit 1; }
```

Then either verify with that file for this command:

```sh
git -c gpg.ssh.allowedSignersFile="$HOME/.config/git/cpace-allowed-signers" \
  verify-tag vX.Y.Z
```

or configure it as the default SSH allowed signers file:

```sh
git config --global gpg.ssh.allowedSignersFile \
  ~/.config/git/cpace-allowed-signers
```

If you already maintain a global allowed signers file, append the generated
`the-sarge@the-sarge.com` entries to that file instead of replacing your
existing Git configuration.

Fetch tags and verify the tag signature:

```sh
git fetch --tags https://github.com/the-sarge/cpace.git
git verify-tag vX.Y.Z
```

Expected output includes a good SSH signature for
`the-sarge@the-sarge.com`, such as:

```text
Good "git" signature for the-sarge@the-sarge.com with ED25519 key SHA256:...
```

The expected release signer identity is:

- Joshua Sargent
- `the-sarge@the-sarge.com`

If `git verify-tag` reports a different signer identity or cannot verify the
signature, do not treat the release as authentic until the discrepancy is
resolved.

To inspect the signed tag and the commit it names:

```sh
git show --no-patch --format=fuller vX.Y.Z
git rev-list -n 1 vX.Y.Z
```

The tag must be an annotated tag, not a lightweight tag:

```sh
git cat-file -t vX.Y.Z
```

Expected output:

```text
tag
```

## Verify Release Author Identity

Official release tags should be authored by the project maintainer identity
listed above. Inspect the tagger identity:

```sh
git for-each-ref refs/tags/vX.Y.Z \
  --format='%(taggername) %(taggeremail)'
```

Expected output:

```text
Joshua Sargent <the-sarge@the-sarge.com>
```

Then verify the tag signature:

```sh
git verify-tag vX.Y.Z
```

The signature identity and tagger identity should both correspond to the
expected maintainer identity. If a future release is produced by automation,
the release notes must identify that process and the process signing identity
before the release is considered official.

## Verify Source Content

To inspect the exact source tree for a release:

```sh
git checkout --detach vX.Y.Z
git status --short
```

The worktree should be clean after checkout. Compare the commit SHA with the
release notes or evidence docs for that release.

## Release Assets

GitHub Releases for signed `v*` tags attach these release-managed assets:

- `cpace-<tag>.cdx.json`: CycloneDX JSON 1.5 SBOM for the source release.
- `cpace-<tag>.cdx.json.sigstore.json`: GitHub/Sigstore bundle emitted by the SBOM attestation workflow.

The release body includes the SBOM's SHA-256 checksum for corruption detection:

```sh
shasum -a 256 cpace-<tag>.cdx.json
```

Compare the computed digest with the release-body value before using the SBOM. This checksum is not an authenticity mechanism because the GitHub Release body is mutable release metadata in the same trust domain as the uploaded assets. Use signed-tag verification for source authenticity and the SBOM attestation for SBOM authenticity.

Optional layered SBOM attestation verification with the GitHub CLI:

```sh
gh attestation verify cpace-<tag>.cdx.json --repo the-sarge/cpace --predicate-type https://cyclonedx.org/bom
```

GitHub may display auto-generated source archives for tags. Treat the signed Git tag as the canonical authenticity mechanism for source releases.

## Support Scope

Release support scope and duration are documented in `SECURITY.md`.
