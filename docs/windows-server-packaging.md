# Windows Packaging, Service Model & In-App Upgrade

Status: **Design / proposed** (revised 2026-05-31). Author decisions captured in this
doc; implementation follows in sequenced PRs. This is the agreed architecture for
taking Xuva on Windows from "desktop launcher only" to a production-grade,
headless-capable, self-updating media server — without losing the existing desktop
experience.

> **Revision note (2026-05-31):** an earlier draft of this doc split Windows into
> **two separate SKUs/installers** (Desktop NSIS + Server MSI). That split is
> **superseded** by the single-installer model below. The driver: a user who wants an
> always-on box that *also* lets them watch locally would otherwise have had to install
> two products — and worse, two competing server engines. The model is now **one
> installer, one binary, install-time component choices**, matching how Emby/Jellyfin
> actually present it. See §2.

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

We also want, on a **single machine**, to support the full range without forcing the
user to pick a "product": just a tray app; a headless always-on service; or **both at
once** (always-on service *plus* a local window to watch on that box).

## 2. Target architecture: one installer, one binary, selectable components

A **single Windows installer** (WiX/MSI — see §4) installs the **one** `xuva-server.exe`
and lets the user choose, in a single pass, which capabilities to enable. There is **no
second product** and **never two server engines on one machine**.

### The one binary, three run modes
`xuva-server.exe` already takes all configuration from `XUVA_*` env / `settings.json`
(`config.FromEnv`) and self-detects its execution context. It runs three ways from one
build:

1. `go run ./cmd/Xuva` — dev.
2. Spawned **console child** of the Electron tray shell — the "watch here" path (unchanged).
3. Under the **Service Control Manager (SCM)** — the headless always-on path
   (`svc.IsWindowsService()` is true).

### Installer components (toggleable, pick any combination in one run)
This is the natural MSI **feature** idiom — selectable checkboxes, **not** a
mutually-exclusive "server vs desktop" radio (which is exactly what would have blocked
the "I want both" case):

| Component | What it does | Default |
|---|---|---|
| **Media server engine** | installs `xuva-server.exe`, ffmpeg/ffprobe, embedded SPA. The core. | always installed |
| **Run as a background service** | registers the Windows Service (LocalSystem, start at boot). The always-on appliance role. | optional |
| **Desktop tray app** | installs the Electron tray/launcher shell + per-user run-at-login. The "watch here" convenience. | optional |

Resulting deployments, all from **one install run**:
- *Laptop / personal watch-here:* engine + tray app, no service.
- *Headless appliance:* engine + service, no tray.
- *Always-on box you also watch on:* engine + service **and** tray app — **one** server
  (the service), with the tray attached to it (see the critical rule below).

### Critical rule: the tray shell attaches, never spawns a duplicate
A media server engine is singular per machine — one library DB, one scanner, one
`:8097` listener. Two would collide on port, database, and media. So the Electron tray
shell must **attach, not spawn**:

> On launch, if a Xuva service is already running on this box, the tray shell simply
> opens its window pointed at the local `http://127.0.0.1:8097`. Only if **no** service
> is present does it spawn its own `xuva-server.exe` child (today's desktop behavior).

This is the same shape Jellyfin/Emby/Plex present: a headless-capable server plus an
optional GUI helper — except unified behind one installer and one binary.

### Why one installer (reversing the earlier "two SKUs" decision)
The earlier draft rejected a single installer on the grounds that conditional
service-vs-desktop logic would bloat the fragile `installer.nsh`. That reasoning is
obsoleted by moving to **WiX/MSI and retiring NSIS** (§4): MSI models optional
components, service registration, and firewall rules **declaratively** — the conditional
complexity NSIS made painful is first-class in MSI. Against that one-time authoring
cost we get: one artifact to download (no "which one do I need?" fork), the "both"
deployment for free, role **switchable post-install** (toggle the service via
`xuva-server.exe service install|uninstall` or an MSI repair), and a single update
substrate (§5). The binary's `svc.IsWindowsService()` self-detection means both
behaviors live in one image with no duplication.

## 3. Native Windows Service support (the foundational enabler — DONE, step 1)

Implemented via `golang.org/x/sys/windows/svc`. When `svc.IsWindowsService()` is true,
the binary runs via `svc.Run` with a handler that maps SCM Stop/Shutdown onto the
existing graceful `http.Server.Shutdown` path; otherwise it takes the current console
path. **No behavior change for dev or the tray-spawned child** — a spawned child is not
SCM-launched, so it stays on the console path exactly as before. Start/shutdown logic is
shared through a single `serverRuntime{}` (`startRuntime`/`shutdown`) used by both paths.

Management subcommands (convenience + scripting; the MSI normally owns the service
declaratively via §4):

- `xuva-server.exe service run` — entry point the SCM invokes.
- `xuva-server.exe service install|uninstall|start|stop` — also the post-install toggle
  for switching a machine between "with service" and "tray-only" without reinstalling.

The service runs with **no console window**, writes only to the file logger
(`XUVA_LOG_DIR`), never blocks on stdin, and assumes no desktop. Under the SCM,
`XUVA_RUNTIME_HOME` defaults to `C:\ProgramData\Xuva` when unset (state stays out of
Program Files; the MSI sets it explicitly).

### Install layout
- Binaries: `C:\Program Files\Xuva\` (`xuva-server.exe`, `bin\ffmpeg.exe`,
  `bin\ffprobe.exe`, embedded web SPA, and — if the tray component is selected — the
  Electron shell).
- Runtime state: `C:\ProgramData\Xuva\` (`data\`, `logs\`, `transcode\`, `metadata\`,
  `cache\`, `temp\`, `downloads\`, `trailers\`). SYSTEM can write it; it survives upgrades.
- Service (when enabled): display name "Xuva Media Server", start type Automatic
  (Delayed Start a candidate to avoid boot contention), recovery = restart on failure.

## 4. Installer: WiX / MSI (single installer, retire NSIS)

One WiX/MSI replaces **both** the current NSIS desktop installer and the previously-planned
separate server MSI. MSI is chosen because the requirements — optional components,
transactional service upgrade, prompt-free in-app upgrade, enterprise deployability — are
exactly what MSI gives for free and what NSIS forced us to hand-roll fragilely.

WiX authoring outline:
- **Features / components** (§2): *engine* (always); *service* (optional —
  `ServiceInstall`/`ServiceControl`); *tray app* (optional — Electron shell + per-user
  run-at-login shortcut). The UI exposes these as checkboxes.
- **`ServiceInstall` / `ServiceControl`** (only when the service feature is selected):
  register "Xuva Media Server" as LocalSystem, Automatic; start on install, stop+remove
  on uninstall. Windows orchestrates **stop → replace files → start** during upgrades,
  **transactionally with rollback** on failure.
- **Updater registration** (§5): a SYSTEM-context updater (scheduled task or the service
  itself) so updates are prompt-free in every mode.
- **WiX Firewall extension**: declare the inbound rules (§7) — replaces the `netsh`
  shell-outs; clean removal on uninstall.
- **ARP / metadata**: publisher, version, icon, help link, Add/Remove Programs entry,
  repair support (repair is also how you add/remove the service or tray component later).
- **Unattended properties** for headless/enterprise deployment:
  ```
  msiexec /i xuva.msi /qn /norestart ^
    ADDLOCAL=Engine,Service XUVA_HTTP_PORT=8097 XUVA_DATA_DIR="D:\XuvaData"
  ```
- **Upgrade**: stable `UpgradeCode`; major-upgrade keyed on version so a newer MSI
  auto-removes the old and installs the new. This is the substrate for §5.

MSI constraints we accept: WiX toolchain added to build + CI; only one `msiexec` at a
time (fine for self-update).

## 5. In-app upgrades — prompt-free, every mode

In-app upgrade is a **required** product capability, and must be **prompt-free** whether
the machine runs the service, the tray app, or both. A single per-machine MSI changes how
we get there versus the old per-user-NSIS plan, so the mechanism is explicit here.

**The one nuance the single installer introduces.** Under the old two-SKU plan, prompt-free
tray updates came "for free" because the tray app installed **per-user** in
`%LOCALAPPDATA%`, which a non-admin user can overwrite. A single **per-machine** MSI lives
in `Program Files`, which a normal user **cannot** silently overwrite — so a tray-only user
applying an update would otherwise hit a UAC prompt. We solve this the way Chrome/Edge do:

> **Install a SYSTEM-context updater once, at install time** (the single, normal UAC consent
> that any installer requires). Thereafter that updater applies `msiexec` upgrades silently,
> regardless of which components are active and regardless of who is logged in.

Concretely:
1. **Check** for a newer release (GitHub Releases or our endpoint).
2. **Download** the MSI; verify SHA-256 (Authenticode later, once signing exists).
3. **Apply** via the SYSTEM-context updater:
   - **Service present:** the service is already SYSTEM and can launch
     `msiexec /i xuva-vX.Y.Z.msi /qn /norestart /l*v <log>` itself — but **detached /
     breakaway** so the updater outlives the service stop mid-upgrade. Leaning toward a
     small installer-registered **"Xuva Updater" scheduled task** (runs as SYSTEM) for
     clean lifetime decoupling rather than a self-detached child.
   - **Tray-only (no service):** the same installer-registered SYSTEM scheduled task
     applies the silent `msiexec` upgrade. No per-update UAC. The tray UI just surfaces
     "update available / installing / restart to finish."
4. **MSI major-upgrade** stops the service (if any), swaps files, restarts it; rolls back
   on failure.

Net: **one** UAC at install (normal for any app), **zero** UAC on every subsequent update,
in all modes. The updater task is the single mechanism; mode only changes who triggers it.

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

### Solution: per-library credentialed SMB (independent of run mode)
1. During library configuration the operator enters a UNC path (`\\NAS\Media`) plus
   optional username/password in the web UI.
2. Server **validates** by connecting with explicit credentials
   (`WNetAddConnection2` or an SMB client that takes per-connection creds).
3. Credentials are **persisted encrypted at rest** — Windows **DPAPI (machine
   scope)** is the natural fit; abstract behind a `SecretStore` interface so
   Linux/Docker get an equivalent (libsecret / file+key).
4. All access to that library uses its stored creds, regardless of how the server
   runs (service or tray-spawned child). Works identically on workgroup and domain.

This makes remote shares "just work" and **supersedes**, when the server runs as a
service, the assumption in [filesystem-access.md](filesystem-access.md) that NAS access
relies on the signed-in user session. A tray-spawned child still has the user session
available, but the credentialed path is preferred uniformly so behavior doesn't depend on
how the engine happens to be running on a given box.

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

Because firewall provisioning is now owned by the MSI (once, declaratively) rather than
per-update `netsh`, it is off the update hot path entirely — a prerequisite for the
prompt-free updates in §5.

## 8. Build pipeline & CI (roadmap item B)

- `packaging/windows/build-package.ps1` already builds `xuva-server.exe` + bundles
  ffmpeg/ffprobe and publishes the SPA. Replace the Electron/NSIS packaging target with
  the **single WiX MSI** target consuming that same compiled server + (optionally) the
  Electron shell payload.
- **CI**: add a Windows runner that builds **the MSI** on PRs touching
  installer/packaging files — closing the "green PR, broken tag build" class that bit
  v0.0.34 three times (the motivation for roadmap item B). Extend
  `package-verification.md` / `release-acceptance.ps1` to cover: MSI silent install (each
  component combination) → service auto-start → health → tray attach-not-spawn →
  upgrade-in-place (prompt-free) → rollback → uninstall (no residue).

## 9. Sequenced implementation plan

1. **Go-native Windows Service mode** in `xuva-server.exe` (svc.Run + subcommands;
   headless-clean). **DONE (step 1)** — shared `serverRuntime{}`, `service_windows.go` /
   `service_other.go`, SCM `XUVA_RUNTIME_HOME` default. Awaiting real-box SCM verification.
2. **Secret store + credentialed SMB access** (DPAPI machine scope) and the
   **credentialed UNC browse/validate API**. Server feature, testable independent of
   packaging. **(step 2 — next)**
3. **Single WiX MSI** with selectable components (engine / service / tray), Firewall
   extension, ARP, unattended properties, UpgradeCode — and the **tray attach-not-spawn**
   behavior in the Electron shell.
4. **Prompt-free in-app upgrade** via the installer-registered SYSTEM updater task
   (check → verify → silent `msiexec`), exercised in both service and tray-only modes.
5. **CI Windows MSI build + verification** (roadmap B), then real-box acceptance before
   tagging.

(The old plan's separate "de-elevate Desktop NSIS updates" step is gone — there is no
NSIS path anymore; prompt-free updates fall out of step 4 for every mode.)

## 10. Open questions (to resolve during implementation)

- Updater lifetime model: detached `msiexec` child vs MSI-installed "Xuva Updater"
  scheduled task (leaning task — required anyway for prompt-free tray-only updates).
- Delayed-start vs Automatic for the service (boot contention vs availability).
- `ProgramData\Xuva` ACL hardening (who beyond SYSTEM/Administrators can read
  data/secrets).
- Tray attach handshake details: how the tray shell reliably detects a running service
  vs. a stale port (health probe + service query) before deciding attach vs spawn.
- Whether to also offer a service-account install path later for shops that prefer
  native share auth over per-library creds (kept as a future option; not built now).
- Cross-platform `SecretStore` parity (Linux/Docker) so the credential feature isn't
  Windows-only.
