# Security Policy

## Supported Scope

Xuva is pre-alpha. Security fixes are accepted for:

- `server/**`
- `apps/web/svelte/**`
- `.github/workflows/**`

## Reporting a Vulnerability

Please use GitHub Security Advisories for private disclosure:

- [Xuva Security Advisories](https://github.com/jampat000/Xuva/security/advisories)

Do not open public issues for active vulnerabilities.

Include:

- affected route/feature
- reproduction steps
- expected vs actual behavior
- impact assessment

## Repository Hardening Baseline

Current baseline:

- main branch protection enabled
- stale review dismissal enabled
- one approval required for PR merge
- conversation resolution required before merge
- linear history required
- force-push blocked on main
- branch deletion blocked on main
- admins/owner can still push directly when needed
- secret scanning enabled
- push protection enabled
- Dependabot security updates enabled

Detailed policy:

- [Repository Hardening](docs/repository-hardening.md)

