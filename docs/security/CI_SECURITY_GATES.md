# WebShield — CI Security Gates and Merge Policy

## 1. Purpose

This policy defines the automated validation that must be completed before a change to the protected target branch is accepted.

A security gate is a condition whose result determines whether a change is eligible to progress. The workflow executes checks; repository rules must enforce that required checks pass before merging.

A green CI run is evidence that the configured checks completed successfully. It is not proof that the entire application is secure.

## 2. Gate inventory

| Gate | Scope | Required outcome |
| --- | --- | --- |
| Go toolchain verification | P4 backend | Expected toolchain is selected |
| Formatting | P4 backend | All applicable Go files are formatted |
| `go vet` | P4 backend | Command completes successfully |
| Tests | P4 backend | Tests complete successfully |
| Race detection | P4 backend, CI | Instrumented tests complete successfully |
| Build | P4 backend | Packages compile successfully |
| Gosec SAST | P4 backend | Scanner completes and returns an accepted status |
| govulncheck | P4 backend | Vulnerability analysis completes and policy is satisfied |
| Gitleaks | Shared repository and available history | Scan completes and confirmed secret exposure is handled |

The scanner's exit code and the findings it reports must be interpreted according to the scanner's behavior and team policy. Do not assume that a process exit code of zero proves comprehensive security.

## 3. Failure classifications

### A. Quality or test failure

Examples: formatting is required, vet reports a problem, a test fails, or the backend does not build.

Action: investigate the regression, correct the responsible code, and rerun the relevant checks.

### B. Security finding

Examples: Gosec reports a potentially unsafe operation, govulncheck reports a relevant vulnerability, or Gitleaks identifies a likely secret.

Action: preserve the evidence, assess the actual risk, assign the finding to the responsible owner, and remediate or use an approved exception process where permitted.

### C. Inconclusive security analysis

Examples: package loading fails, export data cannot be decoded, the scanner crashes, or the intended module cannot be found.

Action: mark the relevant scan as failed or inconclusive and fix the operational problem. Do not represent it as a clean scan.

### D. Infrastructure or dependency retrieval failure

Examples: a runner cannot retrieve a required tool or a necessary resource is unavailable.

Action: diagnose the retrieval or infrastructure problem and rerun the gate when the cause is corrected. A retry does not justify silently bypassing the check.

## 4. Security-finding triage

A finding should record, as applicable:
- Scanner and rule or vulnerability identifier.
- Affected module, component, file, and line.
- Severity and potential impact.
- Attack prerequisites and evidence.
- Responsible component owner.
- Remediation and regression-test plan.
- Status and verification evidence.

Assign fixes to the team that owns the affected component. P4 may identify issues in shared repository content, but ownership of P1, P2, and P3 source code remains with their respective owners.

## 5. Exception process

An exception should identify:
- The exact finding and affected scope.
- A technical justification supported by evidence.
- The risk owner and approving reviewer.
- Compensating controls.
- A tracking issue and remediation plan where applicable.
- An expiration or mandatory review date.

Exceptions must be narrow and time-bound. Do not use an exception to conceal a scanner failure or bypass an uncompleted analysis.

Confirmed exposed credentials require the credential-response process. Removing a finding from a report is not a substitute for revoking or rotating the credential.

## 6. GitHub merge enforcement

The workflow is configured under `.github/workflows/backend-ci.yml`, but merge enforcement is a separate repository configuration responsibility.

Configure a branch protection rule or repository ruleset for the intended target branch. Require all applicable check runs to pass before merging. Verify the actual status-check names displayed by GitHub after a successful workflow run.

The job names currently defined in the workflow are:
- `P4 backend quality`
- `P4 backend SAST`
- `P4 backend dependency scan`
- `Shared repository secret scan`

Confirm the identifiers presented by GitHub before setting the required-check list. The existence of a workflow, by itself, does not prevent a user with sufficient permissions from merging around it.

Restrict bypass permissions to explicitly authorized administrators and maintainers. Document any break-glass use and review it afterward.

## 7. Scope and limitations

- The quality, SAST, and dependency jobs analyze the independent `backend/` Go module.
- The repository secret scan is a shared control over the repository content and history available to the scanner.
- The workflow does not constitute a CD or deployment pipeline.
- Scanner findings require engineering judgment and contextual assessment.
- Tests cover the cases they exercise; they do not prove the absence of defects.
- Repository policy must be checked directly in GitHub because local workflow validation cannot establish that branch protection is enabled.

## 8. Evidence and auditability

For a change requiring review, retain:
- The commit or pull request reference.
- The CI run and check results.
- Relevant security findings and their dispositions.
- Approved exceptions and review dates.
- Evidence of remediation and successful retesting.

Do not record secret values in evidence or ticket comments.

## References

- GitHub protected branches: https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches
- GitHub Actions: https://docs.github.com/en/actions
- Go security policy: https://go.dev/security/policy
- Go vulnerability management: https://go.dev/doc/security/vuln/
- Gitleaks: https://github.com/gitleaks/gitleaks
