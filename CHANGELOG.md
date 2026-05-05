# Changelog

## Unreleased

- Define the package-owned wire framing as format v1 with prefix byte `0xc1`.
  No released versions used the earlier draft-revision byte.
- Add `ErrRandomness` for random-source read failures and unusable scalar
  samples.
- Document and test that `Finish` consumes protocol state even when parsing or
  confirmation fails.
