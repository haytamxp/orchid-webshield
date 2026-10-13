# WebShield P4 — Static Application Security Testing

## Purpose

Static Application Security Testing (SAST) analyzes source code without requiring the application to be running. WebShield uses Gosec to identify Go code patterns associated with potential security weaknesses.

SAST is a detection control. A finding is evidence that a pattern deserves investigation, not automatic proof that exploitation is possible.

## Scope

The P4 CI job executes Gosec from the `backend/` module and analyzes its Go packages.

Shared repository scanning does not transfer ownership of P1, P2, or P3 code to P4. Findings in another component should be documented and routed to its owner for remediation.

## Local execution

First confirm that the installed scanner is compatible with the selected Go toolchain.

When the scanner is available on the executable path, run it from the backend module:

```powershell
Push-Location "D:\orchid-webshield\backend"
try {
    go version
    go env GOMOD
    gosec ./...
}
finally {
    Pop-Location
}
```

If `gosec` is not on `PATH`, use the full path to the installed executable, for example:

```powershell
& "$env:USERPROFILE\go\bin\gosec.exe" ./...
```

These are examples of invocation. Do not classify the scan as successful until the process exits successfully and the output confirms that the intended packages were analyzed.

## How SAST works

A static analyzer inspects source code and applies security rules to identify suspicious operations or programming patterns.

Examples of patterns that may require review include:
- HTTP servers without explicit timeout configuration.
- File operations where an untrusted input may influence the path.
- Ignored errors that could hide an unexpected failure.
- Insecure cryptographic or TLS configuration.

The correct remediation depends on the surrounding code, threat model, and existing controls.

## Finding triage

For each finding:

1. Record its rule identifier, file, line, severity, and description.
2. Read the complete code path around the finding.
3. Identify whether attacker-controlled input can reach the operation.
4. Determine the security impact and whether other controls mitigate it.
5. Assign the finding to the owner of the affected component.
6. Fix and test the code, or document a justified exception using the project's approved process.

Do not dismiss a finding solely because the application appears to work in a manual test.

## Example: unsafe file path

Suppose a backend handler constructs a file path from a request parameter and opens the resulting file.

Potential impact:
- An attacker might influence which file the process reads.
- The actual risk depends on path construction, validation, filesystem permissions, and the files accessible to the process.

Investigation:
- Trace the request parameter to the file operation.
- Identify the permitted directory and expected filename format.
- Validate against the intended allowlist or resolve paths and verify their relationship to the permitted directory.
- Add regression tests for traversal sequences, absolute paths, and valid requests.

A rule finding alone does not prove that the attack succeeds. The engineer must verify the data flow and the effectiveness of the mitigation.

## False positives and suppressions

False positives are possible because static analysis cannot always infer application-specific intent.

Before suppressing a finding:
- Verify the code context and the rule behavior.
- Document why the finding is not exploitable in this context.
- Use the narrowest available suppression.
- Include the rule identifier and a meaningful explanation where the tool supports them.
- Record who approved the exception and when it must be reviewed.

Never add blanket suppressions just to make CI green. Suppression must not replace remediation of a real security weakness.

## Failure handling

Distinguish between:
- A completed scan with findings.
- A completed scan with no findings.
- A scanner crash, incompatible export data, package-loading error, or other incomplete scan.

Only the first two describe findings from a completed analysis. A scanner that fails to analyze the package has not established that the package is clean.

Record scanner version, Go version, module directory, command, exit status, and relevant output when diagnosing failures.

## References

- Gosec: https://github.com/securego/gosec
- Go security policy: https://go.dev/security/policy
- Go vulnerability management: https://go.dev/doc/security/vuln/
