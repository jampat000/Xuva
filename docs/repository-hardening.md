# Repository Hardening

This document defines the repository controls for Xuva.

## Main Branch Policy

Branch: `main`

- Pull request review required: `1`
- Dismiss stale reviews: enabled
- Conversation resolution: required
- Linear history: required
- Force push: blocked
- Branch deletion: blocked
- Admin enforcement: disabled

Admin enforcement remains disabled intentionally so the repository owner/admin can still commit and push directly to `main` for emergency or operational needs.

## Security Controls

- Secret scanning: enabled
- Secret scanning push protection: enabled
- Dependabot security updates: enabled
- Security CI workflow: enabled at `.github/workflows/security.yml`

## Ownership

Code ownership is defined in:

- `.github/CODEOWNERS`

## Operational Notes

- Xuva is the active source of truth.
- Legacy repositories are not required for new implementation work.
- Local cleanup can remove stale clones/remotes once Xuva is confirmed healthy.

