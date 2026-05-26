# Filesystem Access

Xuva must treat filesystem access as an operator-visible local capability, not as a hidden service account concern.

## Windows Desktop Package

The Windows desktop package runs in the signed-in user session by default. This is intentional:

- Local disks visible in File Explorer should be selectable.
- Mapped drives should work when the signed-in user can access them.
- UNC paths such as `\\nas\media\movies` should work when the signed-in user has permission.
- Removable USB drives should appear through the native folder picker.
- NAS paths should not require running Xuva as an elevated Windows service.

The desktop shell should provide native folder picking for:

- library folders
- cache folder
- transcode work folder
- metadata/artwork folder
- backup folder
- download/preparation folder

The browser-only folder browser remains a fallback for dev, headless, and remote-admin cases. It cannot see everything the desktop shell can see because it is limited to server-side filesystem enumeration.

## Linux Bare Metal

Linux packages should run under an explicit operator-controlled user. Media paths must be mounted into that user's namespace with readable permissions.

Supported targets:

- local disks
- mounted USB disks
- mounted SMB shares
- mounted NFS shares
- bind mounts

Do not assume `/media`, `/mnt`, or `/srv` layouts. The package should let the operator choose library and runtime paths.

## Docker

Docker cannot browse the host filesystem directly. Operators must bind mount host paths into the container.

Example:

```yaml
volumes:
  - xuva_data:/data
  - /srv/media/movies:/movies:ro
  - /srv/media/tv:/tv:ro
```

Inside Docker, Xuva should only show mounted container paths. Host paths that were not mounted are intentionally invisible.

## Runtime State

Writable runtime state must stay separate from media libraries:

- database
- settings
- logs
- backups
- cache
- transcode work
- generated metadata
- downloaded/prepared files

Media library paths may be read-only. Runtime paths must be writable by the Xuva process.

## Validation

Before a release can be considered stable:

- Windows package can add a local library folder.
- Windows package can add a mapped-drive or UNC NAS library folder.
- Windows package can move cache/transcode folders and preserve them after restart.
- Docker can scan bind-mounted read-only media folders.
- Docker persists settings/database across container replacement.
- Linux package can scan mounted SMB/NFS folders under the runtime user.
