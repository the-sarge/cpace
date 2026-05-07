# VEX Records

This document records vulnerability-exploitability decisions for vulnerabilities
in software components that do not affect this project.

Current VEX entries: none.

## Policy

If SCA reports a vulnerability in a dependency, tool, generated artifact, or
other software component used by the project, first determine whether the
project is affected. If the project is affected or impact is uncertain, fix,
upgrade, replace, or remove the component before release.

Only use a VEX entry when the project is not affected and the rationale is
specific enough for reviewers and downstream users to inspect. Do not use VEX
to hide uncertainty.

Each VEX entry should include:

- vulnerability ID;
- affected component and version;
- project status, such as `not_affected`, `fixed`, or `under_investigation`;
- affected project releases or commits;
- technical justification for non-exploitability or non-applicability;
- review date and reviewer;
- links to advisories, tool output, pull requests, or release notes.

When the first real VEX entry is needed, prefer adding a machine-readable
OpenVEX or CycloneDX VEX artifact alongside the human-readable summary here so
downstream SCA tools can consume the decision.

## Entry Template

```text
Vulnerability:
Component:
Component version:
Project status:
Affected project releases:
Justification:
Reviewed on:
Reviewer:
References:
```
