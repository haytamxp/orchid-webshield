# WebShield P4 — CI Operations

## Purpose

Continuous Integration (CI) automatically validates code changes as developers integrate them into the repository. This workflow validates the P4 Go backend and runs repository-wide secret detection.

CI is not a deployment pipeline. Release preparation and deployment are outside this milestone.

## Architecture

Developer change
    -> GitHub Actions trigger
    -> Backend quality checks
    -> Backend SAST
    -> Backend dependency vulnerability analysis
    -> Shared repository secret scan
    -> Results available on the pull request

A failed check must be investigated. A scan that could not load or analyze its inputs is not a clean security result.

## Workflow responsibilities

The workflow is `.github/workflows/backend-ci.yml`.

| Job | Responsibility |
| --- | --- |
| P4 backend quality | Verify toolchain, formatting, go vet, tests, race detection, and build |
| P4 backend SAST | Analyze backend source code with Gosec |
| P4 backend dependency scan | Check analyzed Go code and dependencies with govulncheck |
| Shared repository secret scan | Scan repository content and available Git history with Gitleaks |

The backend checks use the `backend/` Go module. The separate `webapp/` module belongs to another project responsibility and must not be silently treated as P4-owned source code.

## Continuous Integration versus deployment

CI provides rapid feedback about code quality and security before integration. Continuous Delivery prepares releasable software; Continuous Deployment automates deployment to an environment.

WebShield's current milestone implements CI checks. It does not deploy the backend.

## Running local validation

From PowerShell, run:

```powershell
& "D:\orchid-webshield\devsecops\ci\validate-local.ps1"
```

The script is designed to select the Go 1.26.9 toolchain and run formatting, vet, tests, and build checks for the backend.

The first execution may require downloading the selected toolchain. Toolchain selection does not guarantee scanner compatibility; validate the actual scanner outputs separately.

## Failure handling

1. Identify which check failed.
2. Preserve the complete error output and identify the failing command.
3. Determine whether the failure is a code defect, test regression, toolchain issue, dependency problem, or scanner loading error.
4. Fix the cause within the responsible component.
5. Rerun the relevant local checks and confirm the CI result.

Do not convert an inconclusive scan into a passing result by suppressing errors or interpreting zero findings as proof that the code is secure.

A failure caused by unavailable dependencies or a package-loading error is operationally different from a vulnerability finding, but it still means the intended validation was not completed.

## Toolchain consistency

The CI workflow selects Go 1.26.9 and uses `GOTOOLCHAIN=local` so the commands use the installed CI toolchain instead of switching automatically.

The local script selects the same named toolchain without changing the user's persistent system configuration.

A matching version string is necessary for repeatability, but scanner analysis must still be tested for compatibility with the module being analyzed.

## Merge enforcement

Successful workflow runs do not automatically prove that GitHub blocks merges when checks fail.

Configure branch protection or a repository ruleset for the target branch. Require the applicable CI status checks before merging. Confirm the actual check names in GitHub after the workflow has completed.

Do not assume merge restrictions are enabled merely because the workflow exists.

## Known limitations

- Static analysis does not prove the application is free of vulnerabilities.
- Tests cover only the behavior exercised by test cases.
- Race detection finds certain races exercised during instrumented tests; it is not exhaustive.
- Dependency scanning relies on vulnerability information available to the scanner.
- Secret scanning can generate false positives and can miss unrecognized secret formats.
- Passing CI is not a substitute for authorized security testing or review.

## References

- Go toolchains: https://go.dev/doc/toolchain
- Go command: https://go.dev/cmd/go/
- GitHub Actions: https://docs.github.com/en/actions
- GitHub protected branches: https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches
