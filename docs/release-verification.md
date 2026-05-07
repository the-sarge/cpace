# Release Verification

Official releases are published as signed annotated Git tags. This project does
not currently attach binary release assets to GitHub releases. The canonical
release artifact is the repository content reachable from the signed tag.

## Verify A Release Tag

Fetch tags and verify the tag signature:

```sh
git fetch --tags https://github.com/the-sarge/cpace.git
git verify-tag vX.Y.Z
```

Expected output includes a good SSH signature for
`the-sarge@the-sarge.com`.

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
  --format='%(taggername) <%(taggeremail)>'
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

Current GitHub releases have no attached binary assets. If future releases add
assets, each asset must be signed directly or included in a signed manifest that
lists cryptographic hashes for every asset. Verify those signatures or manifest
hashes before using the assets.

GitHub may display auto-generated source archives for tags. Treat the signed Git
tag as the canonical authenticity mechanism for source releases.

## Support Scope

Release support scope and duration are documented in `SECURITY.md`.
