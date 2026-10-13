# WebShield P4 — Secret Detection and Credential Response

## Purpose

A secret is sensitive authentication material or private configuration, such as an API key, access token, password, or private key.

Gitleaks scans repository content for patterns that may indicate exposed secrets. The WebShield workflow checks out full Git history so the scan is not limited to the current working-tree snapshot.

Secret scanning reduces accidental exposure. It cannot guarantee that every kind of secret will be detected.

## Prohibited secret content

Never place the following in source code, event payloads, logs, test fixtures, or committed configuration:
- Passwords.
- API keys and bearer tokens.
- Session cookies.
- Private cryptographic keys.
- Production credentials or sensitive connection strings.

Use an approved secret manager or environment-specific secret configuration instead. Do not log authentication headers or session material.

## How Git history changes the risk

Deleting a secret from the latest version of a file does not remove it from earlier commits.

Example:
1. A developer commits an API key.
2. The developer removes the key in a later commit.
3. The old commit still contains the credential.
4. Anyone able to access that history may still retrieve the exposed value.

Repository history scanning helps identify this situation. It does not revoke a credential that has already been exposed.

## Response to a confirmed leak

1. Treat the credential as compromised.
2. Identify the owner and the system protected by the credential.
3. Revoke or rotate the credential promptly using the issuing service or system.
4. Review relevant access logs for suspicious use.
5. Replace the exposed credential in the authorized secret store and affected environments.
6. Determine whether history cleanup or additional containment is needed.
7. Document the incident and prevention measures.

Rotation or revocation is the priority. Removing the text from a commit does not make the original credential safe to continue using.

Coordinate response with the relevant system owner. Do not independently alter another component's credentials or deployment configuration.

## False positives

A finding can be a harmless example or test value, but it must be reviewed before being dismissed.

Use synthetic values designed to be clearly nonfunctional. Do not insert real credentials merely to test whether the scanner detects them.

Where suppression or allowlisting is supported:
- Keep it narrow.
- Explain why the value is not a usable secret.
- Record the owner and rationale.
- Review exceptions periodically.

Never blanket-ignore a file or directory merely because it has generated repeated findings.

## Failure handling

Distinguish between a completed scan that reports a candidate secret and a scan that failed to execute.

For a confirmed secret, initiate credential response.

For a scanner error, preserve the output and establish whether the intended repository content and history were analyzed. A failed secret scan is not a clean secret scan.

## Operational controls

The workflow uses a full-history checkout for the repository secret scan. This improves coverage of committed history but does not automatically inspect every external copy, artifact, fork, or system where a credential may have been reused.

Access controls, least privilege, short-lived credentials, rotation, and secret managers remain necessary even when automated scanning is enabled.

## References

- Gitleaks: https://github.com/gitleaks/gitleaks
- GitHub secret scanning: https://docs.github.com/en/code-security/secret-scanning
- GitHub repository history: https://docs.github.com/en/repositories
