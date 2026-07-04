# OSS-Fuzz-Compatible Staging

This directory keeps the OSS-Fuzz-compatible build files for local helper validation, ClusterFuzzLite experiments, or a future OSS-Fuzz resubmission if project eligibility changes. The upstream submission `google/oss-fuzz#15480` was closed on 2026-05-11 after OSS-Fuzz maintainers declined the project for current project-size/user-base reasons and suggested ClusterFuzzLite instead.

The active fuzz targets remain in this repository in `fuzz_test.go` and are registered locally in `.github/fuzz-targets.json`, where each entry names the target function, package, and OSS-Fuzz-compatible binary name; `go test ./...` checks those entries against this build script.

If an upstream OSS-Fuzz resubmission becomes appropriate later, prefer a small delegate `build.sh` in `google/oss-fuzz/projects/cpace` that executes this repository's `ossfuzz/build.sh` instead of duplicating the target list there.

For local OSS-Fuzz helper validation:

1. Copy these files into a temporary `google/oss-fuzz` checkout under `projects/cpace`.
2. Confirm `primary_contact` in `project.yaml` is the maintainer Google-account-associated email that should receive ClusterFuzz access and private bug notifications if this is used for an upstream resubmission.
3. Build and check locally from the OSS-Fuzz checkout:

```sh
python3 infra/helper.py build_image cpace
python3 infra/helper.py build_fuzzers --sanitizer address cpace /path/to/cpace
python3 infra/helper.py check_build cpace
```

On Apple Silicon hosts, use the production `x86_64` path explicitly:

```sh
DOCKER_DEFAULT_PLATFORM=linux/amd64 python3 infra/helper.py build_fuzzers --architecture x86_64 --sanitizer address cpace /path/to/cpace
DOCKER_DEFAULT_PLATFORM=linux/amd64 python3 infra/helper.py check_build cpace
```

OSS-Fuzz native Go fuzzing builds `testing.F` fuzzers as libFuzzer binaries.
The Go integration currently supports `libfuzzer` with the `address` sanitizer.
For native Go fuzzers, `F.Add` seeds are not imported automatically by
OSS-Fuzz; add explicit seed corpora later if the first coverage reports show a
need for them.

Local validation on 2026-05-07 used a temporary `google/oss-fuzz` checkout,
mounted this repository as the source path, and passed:

```sh
DOCKER_DEFAULT_PLATFORM=linux/amd64 python3 infra/helper.py build_fuzzers --architecture x86_64 --sanitizer address cpace /Users/josh/code/github.com/the-sarge/cpace
DOCKER_DEFAULT_PLATFORM=linux/amd64 python3 infra/helper.py check_build cpace
```
