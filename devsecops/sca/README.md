# WebShield P4 — Software Composition Analysis

## Purpose

Software Composition Analysis (SCA) examines the third-party components used by a software project and identifies dependency-related risks.

WebShield uses govulncheck to analyze Go code and dependencies against known Go vulnerability information.

SCA complements SAST: SAST looks for risky patterns in the application source; dependency scanning looks for known vulnerabilities affecting components the application uses.

## Scope

The P4 dependency job runs from `backend/`, using the independent Go module defined in `backend/go.mod`.

Do not run a module-dependent command from the repository root and assume it analyzed the backend. The command must resolve the intended module.

## Local execution

After confirming toolchain and scanner compatibility, run from the backend directory:

```powershell
Push-Location "D:\orchid-webshield\backend"
try {
    go version
    go env GOMOD
    govulncheck ./...
}
finally {
    Pop-Location
}
```

If the executable is not on `PATH`, use its full executable path:

```powershell
& "$env:USERPROFILE\go\bin\govulncheck.exe" ./...
```

Use the actual installed path when it differs.

## How dependency vulnerability analysis works

Go modules declare dependencies and versions. Dependencies may be direct or indirect, and their own dependencies may introduce additional components into the software supply chain.

Govulncheck uses Go vulnerability information to evaluate known vulnerabilities relevant to the analyzed program. Depending on available source and package information, its results can distinguish vulnerable symbols and call paths from vulnerabilities that are not shown to be reachable in the analyzed code.

Reachability analysis improves prioritization, but it does not mean that an issue outside a currently observed path is harmless under every configuration or future code change.

## Remediation workflow

When a vulnerability is reported:

1. Capture the vulnerability identifier, affected module and version, affected symbols, and scanner output.
2. Confirm that the finding applies to the dependency version used by this module.
3. Read the vulnerability advisory and identify affected and fixed versions.
4. Select an appropriate fixed or otherwise mitigated version.
5. Update dependency metadata intentionally using the Go toolchain.
6. Review the dependency diff and relevant release notes.
7. Run tests, vet, build, and govulncheck again.
8. Record the remediation and verification result.

Do not blindly upgrade every dependency without reviewing compatibility, and do not dismiss a known vulnerability merely because a direct exploit has not been reproduced.

## Read-only CI behavior

The CI quality checks use `-mod=readonly` to avoid silently rewriting module definitions and checksums during validation.

If dependency metadata must change, make the dependency change deliberately, review it, test it, and commit the resulting files as part of the appropriate change.

## Failure handling

An output that says no vulnerabilities were found is useful only when the scanner completed the intended analysis.

Treat the following as inconclusive:
- Missing `go.mod` or execution from the wrong module directory.
- Package-loading or export-data errors.
- Unsupported or incompatible toolchain behavior.
- Scanner crashes or inability to retrieve required vulnerability information.

A failed scan is not equivalent to a completed scan with zero findings. Preserve the error, confirm the executable and Go versions, check the selected module, and rerun after correcting the underlying problem.

## Risk prioritization

Prioritize based on more than a version number. Consider:
- Whether the vulnerable component is used by the backend.
- Whether the vulnerable code path is reachable in the analyzed program.
- Attack prerequisites and realistic impact.
- Availability of a patched version or practical mitigation.
- Compatibility and regression risk of an upgrade.

Document accepted risk with a responsible owner, justification, compensating controls, and review or expiration date under the security-gate policy.

## References

- Go vulnerability management: https://go.dev/doc/security/vuln/
- Go vulnerability database: https://vuln.go.dev/
- govulncheck: https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck
- Go modules: https://go.dev/ref/mod
