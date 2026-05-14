# Product Principles

## Local First

Xuva should treat the user's server as the source of truth.

Local playback, local users, local libraries, local configuration, and local device pairing must work without a vendor account or internet connection.

## No Required Vendor Infrastructure

Xuva should not require company-hosted servers for core use.

In v1, Xuva will not provide:

- Mandatory central authentication.
- Vendor-hosted media relay.
- Vendor-hosted dashboard.
- Vendor metadata lock-in.
- Required cloud account for LAN playback.

## User-Owned Remote Access

Xuva provides mechanisms and diagnostics to help users configure remote access, but users own their remote infrastructure.

Supported paths include:

- LAN.
- Direct port forwarding.
- User-managed reverse proxy.
- User-managed VPN or mesh network.
- User-managed domain and TLS.
- Offline downloads.

## Explainable Playback

Users should never be left guessing why a file is transcoding.

Every playback session should expose:

- Container.
- Video codec and profile.
- Audio codec and channel layout.
- Subtitle format.
- Client capability match.
- Chosen playback path.
- Exact fallback reason.
- Estimated server cost.

## Personal Media First

The first screen is the user's library. Commercial streaming content, ads, and unrelated recommendations should not take priority over the user's own media.

## Durable Ownership

The server should keep functioning if the company disappears. Paid licenses may be locally cached and validated offline. Core local playback should not depend on a live vendor service.

