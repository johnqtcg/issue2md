# Security Policy

This document explains how to report vulnerabilities, expected response timelines, and disclosure principles for `issue2md`.

## 1. Supported Versions

By default, only the latest code on `main` is guaranteed to receive security fixes.

| Version | Supported |
|---|---|
| `main` | Yes |
| Older versions / tags | No (unless explicitly announced) |

## 2. Reporting a Vulnerability

Please do not disclose vulnerability details in public Issue/PR threads.

Preferred private channel:
- GitHub Security Advisory (Private Report):
  `https://github.com/johnqtcg/issue2md/security/advisories/new`

If you cannot use that channel:
- Open a public Issue without sensitive details and request a private contact method.
- Do not publish PoC, exploit path, sensitive config, or reproducible payloads before private coordination.

## 3. What to Include

To speed up triage and remediation, include:
- Affected version/commit
- Vulnerability type and impact
- Reproduction steps
- Minimal PoC (sanitized if needed)
- Optional fix suggestion

## 4. Response SLA (Target)

Maintainers target:
- Acknowledgement within 48 hours
- Initial assessment within 7 calendar days
- Ongoing updates in the private report thread until remediation

This SLA is a best-effort target, not a legal guarantee.

## 5. Disclosure Policy

- Before a fix is released, both reporter and maintainers should avoid public disclosure of exploitable details.
- After remediation, responsible disclosure may be coordinated.
- If active exploitation is observed, maintainers may publish mitigations before full details.

## 6. Out of Scope

These are typically out of scope:
- Issues limited to local-only development setup with no realistic attack path
- Findings requiring extreme non-default configuration with no practical impact
- Third-party platform/network incidents not caused by this repository's code

Maintainers make the final triage decision for disputed cases.
