# Windows Server Packaging, Service Model & In-App Upgrade

Status: **Design / proposed** (2026-05-31). Author decisions captured in this doc;
implementation to follow in sequenced PRs. This doc is the agreed architecture for
taking Xuva on Windows from "desktop launcher only" to a production-grade,
headless, self-updating media server — without losing the existing desktop app.

Related docs this extends or qualifies:
[filesystem-access.md](filesystem-access.md), [remote-access.md](remote-access.md),
[local-discovery.md](local-discovery.md), [package-verification.md](package-verification.md),
[desktop-owner-mode.md](desktop-owner-mode.md), [release-versioning.md](release-versioning.md).

---

## 1. Problem statement

Today there is exactly **one** way to run the Xuva server on Windows:

> `Xuva.exe` (Electron tray app) → `spawn("xuva-server.exe")` as a **child process** →
> opens the web UI in the default browser.

The server's lifecycle is welded to an **interactive desktop session**
(`apps/desktop/main.js`). That is correct for an HTPC/personal PC, but wrong for a
dedicated server box:

- **No login, no Xuva.** After reboot the server is down until a user logs in and
  the tray app starts. RDP-logoff tears it down. There is no start-at-boot.
- **Per-user install that force-elevates anyway.** `apps/desktop/package.json` sets
  `perMachine: false` (installs to `%LOCALAPPDATA%`, per-user shortcuts, HKCU ARP),
  yet `apps/desktop/installer.nsh` `customInit` does a hard `UAC_RunElevated` purely
  to write firewall rules + a ProgramData receipt — the worst of both worlds.
- **Accreted NSIS scar tissue.** `installer.nsh` documents failures across
  v0.0.24/25/31/32/33/34 (finish-page launch plugins, `BUILD_UNINSTALLER`
  warning-as-error guards, `StdUtils` DLLs). It works but is fragile.
- **The update mechanism only exists inside the tray app** (server stages installer →
  writes `pending-update.json` → Electron polls every 3s → `powershell RunAs /S`).
  It has no meaning on a headless box with no Electron running.

## 2. Target architecture: two SKUs from one codebase

| | **Xuva Desktop** | **Xuva Server** |
|---|---|---|
| Audience | HTPC / personal PC | dedicated / always-on server |
| Shell | Electron tray launcher (current) | **none** — headless |
| Server lifecycle | spawned child of tray app | **Windows Service**, auto-start at boot |
| Install scope | per-user (`%LOCALAPPDATA%`) | per-machine (`Program Files`) |
| Service identity | signed-in user | **LocalSystem** (+ per-library creds, §6) |
| Installer | NSIS (electron-builder) | **WiX / MSI** |
| In-app upgrade | per-user silent self-replace, no UAC | silent `msiexec`, no UAC (SYSTEM) |
| Filesystem access | user-session (sees mapped drives / UNC as the user) | local disks as SYSTEM + **per-library credentialed SMB** |

Both SKUs compile the **same** `xuva-server.exe`. The binary already takes all
configuration from `XUVA_*` env / `settings.json` (`config.FromEnv`), which is the
hard part and is already done.

> This is the same split Jellyfin/Emby/Plex make: a headless service plus an
> optional GUI helper.

### Why not one installer with a mode toggle
Rejected. A single installer that conditionally installs-as-service vs
installs-as-desktop multiplies the conditional complexity in exactly the layer
(`installer.nsh`) that is already the most fragile. Two purpose-built artifacts are
simpler to reason about, test, and verify in CI.

## 3. Native Windows Service support (the foundational enabler)

`xuva-server.exe` must run three ways from one binary:

1. `go run ./cmd/Xuva` — dev.
2. Spawned console child — the Desktop SKU (unchanged).
3. Under the **Service Control Manager (SCM)** — the Server SKU.

Implementation: integrate `golang.org/x/sys/windows/svc`. Detect SCM context
(`svc.IsWindowsService()`); when true, run via `svc.Run` with a handler that maps
SCM Stop/Shutdown to the existing graceful `http.Server.Shutdown` path. When false,
run the current console path. No behavior change for dev/desktop.

Provide thin management subcommands for manual/scripted use and for the MSI to call
if we ever need imperative control (the MSI will normally do this declaratively —
see §4):

- `xuva-server.exe service run` (entry point the SCM invokes)
- `xuva-server.exe service install|uninstall|start|stop` (convenience; optional once MSI owns it)

Service must run cleanly with **no console window**, write only to the file logger
(`XUVA_LOG_DIR`), and never block on stdin or assume a desktop.

### Install layout (Server SKU)
- Binaries: `C:\Program Files\Xuva\` (`xuva-server.exe`, `bin\ffmpeg.exe`, `bin\ffprobe.exe`, embedded web SPA).
- Runtime state: `C:\ProgramData\Xuva\` (`data\`, `logs\`, `transcode\`, `metadata\`, `cache\`, `temp\`, `downloads\`, `trailers\`). This matches the existing preferred runtime home; SYSTEM can write it; it survives upgrades.
- Service: display name "Xuva Media Server", start type Automatic (Delayed Start is a candidate to avoid boot contention), recovery = restart on failure.

## 4. Installer: WiX / MSI (Server SKU)

MSI is chosen over service-mode NSIS because the requirements (transactional
service upgrade, prompt-free in-app upgrade, enterprise deployability) are exactly
what MSI gives for free and what NSIS would force us to hand-roll fragilely.

WiX authoring outline:
- **Components**: server binary, ffmpeg/ffprobe, embedded SPA assets; `Program Files`
  target; `ProgramData\Xuva` data tree created with appropriate ACLs.
- **`ServiceInstall` / `ServiceControl`**: register "Xuva Media Server" as
  LocalSystem, Automatic start; start on install, stop+remove on uninstall. Windows
  orchestrates **stop → replace files → start** during upgrades, **transactionally
  with rollback** on failure.
- **WiX Firewall extension**: declare the inbound rules (§7) — replaces the `netsh`
  shell-outs; clean removal on uninstall.
- **ARP / metadata**: publisher, version, icon, help link, proper Add/Remove Programs
  entry, repair support.
- **Unattended properties** for headless/enterprise deployment:
  ```
  msiexec /i xuva-server.msi /qn /norestart ^
    XUVA_HTTP_PORT=8097 XUVA_DATA_DIR="D:\XuvaData"
  ```
- **Upgrade**: stable `UpgradeCode`; major-upgrade keyed on version so a newer MSI
  auto-removes the old and installs the new. This is the substrate for §5.

MSI constraints we accept: WiX toolchain added to build + CI; only one `msiexec` at a
time (fine for self-update).

## 5. In-app upgrades — prompt-free, both SKUs

In-app upgrade is a **required** product capability. The model makes it work with
**zero UAC prompts** in both SKUs, which is the strongest reason for the
MSI(server) / per-user-NSIS(desktop) split.

### Server SKU (service as LocalSystem)
The service is **already fully elevated** (SYSTEM), so it can launch the new
installer itself with no prompt:
1. Check for a newer release (GitHub Releases or our endpoint).
2. Download MSI; verify SHA-256 (Authenticode later, when signing exists).
3. Launch `msiexec /i xuva-vX.Y.Z.msi /qn /norestart /l*v <log>` **detached /
   breakaway** from the service process (same survives-parent-death trick the
   current desktop PowerShell updater uses — the updater must outlive the service
   being stopped mid-upgrade).
4. MSI major-upgrade stops the service, swaps files, restarts it; rolls back on
   failure.

Open detail to settle in implementation: whether the `msiexec` invocation is a
detached child of the service or a small **companion "Xuva Updater" scheduled task**
the MSI installs (cleaner lifetime decoupling). Lean toward the scheduled-task/helper
approach so the upgrade is never killed by its own service stop.

### Desktop SKU (per-user)
The current desktop update **forces elevation it does not need**. A per-user install
(into `%LOCALAPPDATA%`, where it already lands) can **self-replace its own files with
no UAC**, like VS Code / Slack (Squirrel / `electron-updater`). The only thing
forcing admin today is `netsh` firewall provisioning in `customInit`. Moving firewall
handling off the per-update path (do it once at first install, or rely on Windows'
native per-app firewall prompt — see §7) makes desktop upgrades **silent**.

## 6. Remote filesystem browsing & access (the real Windows trap)

This is a **server feature + service-identity** problem, not an installer problem.
Decision: **LocalSystem + per-library stored credentials.**

Why service identity matters for shares:
- **Local FS**: SYSTEM reads all local disks — trivial, no issue.
- **Remote FS (`\\NAS\Media`)**:
  - **LocalSystem on a workgroup** (typical prosumer) has **no network credentials**
    — cannot authenticate to a password-protected SMB share. Anonymous/Guest shares
    are disabled by default on modern Windows.
  - **LocalSystem on a domain** authenticates as the machine account
    (`DOMAIN\MACHINE$`), usually not granted share access.
  - **Mapped drive letters don't work in services** (per-logon-session). Must use UNC.

### Solution: per-library credentialed SMB (independent of service identity)
1. During library configuration the operator enters a UNC path (`\\NAS\Media`) plus
   optional username/password in the web UI.
2. Server **validates** by connecting with explicit credentials
   (`WNetAddConnection2` or an SMB client that takes per-connection creds).
3. Credentials are **persisted encrypted at rest** — Windows **DPAPI (machine
   scope)** is the natural fit; abstract behind a `SecretStore` interface so
   Linux/Docker get an equivalent (libsecret / file+key).
4. All access to that library uses its stored creds, regardless of how the service
   runs. Works identically on workgroup and domain.

This makes remote shares "just work" and **supersedes**, for the Server SKU, the
assumption in [filesystem-access.md](filesystem-access.md) that NAS access relies on
the signed-in user session. The Desktop SKU keeps the user-session model described
there; the Server SKU adds the credentialed path.

### Browse-during-configuration API
Server-side directory listing: given a UNC root (+ optional creds), enumerate
subdirectories via the SMB client. Notes:
- *Host discovery* (browsing `\\` to list machines) is unreliable/deprecated on
  Windows (Computer Browser / NetBIOS). Dependable UX = "operator types/pastes the
  share path + creds; we validate and browse from there." Optional sweetener: surface
  mDNS / WS-Discovery host hints.
- Reuse/extend the existing browser-only folder-browser API noted in
  filesystem-access.md, adding credentialed UNC support and validation.

Security: never log credentials; encrypt at rest; redact in diagnostics; bind the
stored secret to the machine (DPAPI machine scope) so a copied data dir can't be
trivially decrypted elsewhere.

## 7. Firewall, discovery & LAN remote management

- **Discoverable on the LAN**: mDNS/Bonjour `_xuva._tcp` (UDP 5353) is already
  advertised ([local-discovery.md](local-discovery.md)). Need inbound **UDP 5353** +
  the **HTTP port** allowed. The MSI declares these via the WiX Firewall extension
  (machine scope, owned by the installer, removed on uninstall) — retiring the
  `netsh` shell-outs in `installer.nsh`.
- **Profile classification**: keep the current rule logic — a **program-scoped**
  rule (`xuva-server.exe` only) on **any** profile (Windows often misclassifies a
  LAN as Public) plus a **port-scoped** rule limited to **Private/Domain**. Express
  declaratively in WiX.
- **Remote management from another PC**: nothing extra to build — it is the web UI
  reachable over the LAN. Bind `0.0.0.0:8097` (already default), open the firewall,
  browse from another machine. Authz/parental gating already exists (#427). Prefer
  **HTTPS** for remote admin (optional TLS + self-signed generation already exists);
  surface a "Remote management URL" with the cert fingerprint in settings.
- **Auth interlock**: the new `XUVA_AUTH_DISABLED` non-loopback startup interlock
  (#429) is especially relevant for a LAN-exposed server — auth must stay on.

## 8. Build pipeline & CI (roadmap item B)

- `packaging/windows/build-package.ps1` already builds `xuva-server.exe` + bundles
  ffmpeg/ffprobe and publishes the SPA. Add an **MSI build target** (WiX) beside the
  existing Electron/NSIS target; both consume the same compiled server.
- **CI**: add a Windows runner that builds **both** the Desktop NSIS installer and
  the Server MSI on PRs touching installer/packaging files — closing the
  "green PR, broken tag build" class that bit v0.0.34 three times (this is the
  motivation for roadmap item B). Extend `package-verification.md` /
  `release-acceptance.ps1` to cover: MSI silent install → service auto-start →
  health → upgrade-in-place → rollback → uninstall (no residue).

## 9. Sequenced implementation plan

1. **Go-native Windows Service mode** in `xuva-server.exe` (svc.Run + subcommands;
   headless-clean). Unit/integration where feasible; manual SCM verification.
2. **Secret store + credentialed SMB access** (DPAPI machine scope) and the
   **credentialed UNC browse/validate API**. Server feature, testable independent of
   packaging.
3. **WiX MSI** for the Server SKU (ServiceInstall/Control, Firewall extension, ARP,
   unattended properties, UpgradeCode).
4. **Server in-app upgrade** path (check → verify → silent `msiexec` via helper task).
5. **De-elevate Desktop updates** (per-user silent self-update; move firewall off the
   update path).
6. **CI Windows installer/MSI build + verification** (roadmap B), then real-box
   acceptance before tagging.

## 10. Open questions (to resolve during implementation)

- Updater lifetime model: detached `msiexec` child vs MSI-installed "Xuva Updater"
  scheduled task (leaning task).
- Delayed-start vs Automatic for the service (boot contention vs availability).
- `ProgramData\Xuva` ACL hardening (who beyond SYSTEM/Administrators can read
  data/secrets).
- Whether to also offer a service-account install path later for shops that prefer
  native share auth over per-library creds (kept as a future option; not built now).
- Cross-platform `SecretStore` parity (Linux/Docker) so the credential feature isn't
  Windows-only.
