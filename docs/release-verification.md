# Release Verification

Official releases are published as signed annotated Git tags. This project does
not currently attach binary release assets to GitHub releases. The canonical
release artifact is the repository content reachable from the signed tag.

## Verify A Release Tag

Git verifies SSH signatures against an allowed signers file. If this is the
first time verifying this project's SSH-signed tags, create a project-specific
allowed signers file from the maintainer's public GitHub SSH keys:

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

Future compiled assets must also ship with an SBOM as described in
`docs/release-checklist.md`. Verify that the release notes identify the SBOM
location before using compiled assets.

GitHub may display auto-generated source archives for tags. Treat the signed Git
tag as the canonical authenticity mechanism for source releases.

## Support Scope

Release support scope and duration are documented in `SECURITY.md`.
