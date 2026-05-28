; Custom NSIS hooks for the Xuva Windows installer.
;
; Responsibilities:
;   1. Provision Windows Firewall rules at install time (so LAN access just
;      works, no manual "as administrator" netsh dance for the user). Rules
;      are scoped to the packaged server exe AND to the chosen ports — the
;      port-scoped rule survives install-path changes (drive change,
;      uninstall+reinstall to a different folder, dev-mode runs from the
;      source tree).
;   2. Present a "Create desktop shortcut" checkbox on a custom wizard page
;      so users who don't want clutter on their desktop can opt out. The
;      checkbox defaults to checked because most users do want it; the
;      power-user opt-out is the new behaviour this hook unlocks.
;
; The custom page uses electron-builder's documented `customPageAfterChangeDir`
; hook — the page is injected by assistedInstaller.nsh between the directory
; chooser and InstFiles. createDesktopShortcut is set to false in package.json
; so the built-in unconditional shortcut creation is disabled; we recreate it
; here only when the user opted in.

!include nsDialogs.nsh
!include LogicLib.nsh

; All three Vars are read/written only in the installer pass: the dialog +
; checkbox vars are touched inside XuvaShortcutPageShow/Leave, and
; XuvaCreateDesktopShortcut is set in customInit + read in customInstall —
; both of which electron-builder's installer.nsi only `!insertmacro`s under
; `!ifndef BUILD_UNINSTALLER`. NSIS warning 6001 ("Variable not referenced or
; never set, wasting memory") is treated as an error by electron-builder's
; strict mode and was what killed the v0.0.24 / v0.0.25 release builds, so
; we gate the declarations the same way.
!ifndef BUILD_UNINSTALLER
  Var XuvaShortcutDialog
  Var XuvaShortcutCheckbox
  Var XuvaCreateDesktopShortcut
!endif

; ── Firewall rules ──────────────────────────────────────────────────────────

!macro XuvaAddFirewallRules
  DetailPrint "Configuring Windows Firewall rules for Xuva..."
  ; Wipe any existing rules of the same names so reinstalls / path changes
  ; don't accumulate stale entries pointing at the previous install dir.
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall delete rule name="Xuva Server HTTP"'
  Pop $0
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall delete rule name="Xuva Server HTTP (Port)"'
  Pop $0
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall delete rule name="Xuva Local Discovery mDNS"'
  Pop $0
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall delete rule name="Xuva Local Discovery mDNS (Port)"'
  Pop $0

  ; Program-scoped rule for the packaged server. Tightest scope — only the
  ; signed Xuva binary at this install location is allowed.
  ;
  ; profile=any covers Public networks too, because Windows often
  ; misclassifies a home LAN as Public (especially after a network change
  ; or on a fresh adapter) and the user then can't reach Xuva from their
  ; phone / tablet / other PC without manually re-classifying the network.
  ; Confirmed against the live media-server: with profile=private,domain
  ; only, the LAN was unreachable until the user manually added Public to
  ; the rule. The program-scoped guard means only the signed Xuva binary
  ; at this install path is allowed in — the rule isn't a generic "open
  ; port 8097 to the world", it's specifically "let traffic in to *this*
  ; xuva-server.exe", which keeps the exposure tight.
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall add rule name="Xuva Server HTTP" dir=in action=allow program="$INSTDIR\resources\runtime\xuva-server.exe" protocol=TCP localport=8097 profile=any enable=yes description="Allow Xuva web and API access on all network profiles (program-scoped)."'
  Pop $0
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall add rule name="Xuva Local Discovery mDNS" dir=in action=allow program="$INSTDIR\resources\runtime\xuva-server.exe" protocol=UDP localport=5353 profile=any enable=yes description="Allow Xuva Bonjour/mDNS discovery on all network profiles (program-scoped)."'
  Pop $0

  ; Port-scoped fallback rule. Survives install-path changes between releases
  ; — without this, an upgrade that lands in a slightly different folder
  ; would leave the user unable to reach the new binary on the LAN until
  ; they reran the installer. Scoped to Private+Domain only (not Public)
  ; because this rule has no program guard, so a "any program on port 8097"
  ; rule on a coffee-shop wifi would be a footgun. The program-scoped pair
  ; above covers the Public case for the actual Xuva binary.
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall add rule name="Xuva Server HTTP (Port)" dir=in action=allow protocol=TCP localport=8097 profile=private,domain enable=yes description="Allow inbound TCP 8097 (Xuva web/API) on private and domain networks."'
  Pop $0
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall add rule name="Xuva Local Discovery mDNS (Port)" dir=in action=allow protocol=UDP localport=5353 profile=private,domain enable=yes description="Allow inbound UDP 5353 (mDNS/Bonjour) on private and domain networks."'
  Pop $0
!macroend

!macro XuvaRemoveFirewallRules
  DetailPrint "Removing Windows Firewall rules for Xuva..."
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall delete rule name="Xuva Server HTTP"'
  Pop $0
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall delete rule name="Xuva Server HTTP (Port)"'
  Pop $0
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall delete rule name="Xuva Local Discovery mDNS"'
  Pop $0
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall delete rule name="Xuva Local Discovery mDNS (Port)"'
  Pop $0
!macroend

; ── UAC: require admin so we can add firewall rules ─────────────────────────

!macro customInit
  ${IfNot} ${UAC_IsAdmin}
    !insertmacro UAC_RunElevated
    ${Switch} $0
      ${Case} 0
        ${Break}
      ${Case} 1223
        Quit
      ${Default}
        MessageBox mb_IconStop|mb_TopMost|mb_SetForeground "Unable to request administrator rights for Xuva setup. Windows Firewall rules cannot be provisioned."
        Quit
    ${EndSwitch}
    Quit
  ${EndIf}

  ; Default the desktop-shortcut choice to ON. The custom page below
  ; overwrites this if the user unchecks the box.
  StrCpy $XuvaCreateDesktopShortcut "1"
!macroend

; ── Custom page: "Create desktop shortcut" checkbox ─────────────────────────
; Wired via the electron-builder hook customPageAfterChangeDir. The page is
; injected by assistedInstaller.nsh between the directory chooser and the
; install progress, which is the natural place for "options about how to
; install" toggles.

!macro customPageAfterChangeDir
  Page custom XuvaShortcutPageShow XuvaShortcutPageLeave
!macroend

; The two Page functions below are only referenced from `customPageAfterChangeDir`
; which is only inserted into the installer build, never the uninstaller.
; NSIS's `warning 6010: install function "..." not referenced - zeroing code` is
; treated as an error by electron-builder, so we have to gate the function
; declarations on `!ifndef BUILD_UNINSTALLER` (defined when building the
; uninstaller stub). Without this guard, v0.0.24/v0.0.25 release builds
; failed the uninstaller compile pass.

!ifndef BUILD_UNINSTALLER

Function XuvaShortcutPageShow
  ; NOTE: don't use !insertmacro MUI_HEADER_TEXT here — electron-builder
  ; !includes this script BEFORE MUI2.nsh is loaded by installer.nsi, so
  ; the macro is undefined at parse time. The header keeps whatever text
  ; the previous wizard page set (Directory chooser → "Choose Install
  ; Location") which is acceptable; the page label below makes its own
  ; purpose obvious.

  nsDialogs::Create 1018
  Pop $XuvaShortcutDialog
  ${If} $XuvaShortcutDialog == error
    Abort
  ${EndIf}

  ${NSD_CreateLabel} 0 0 100% 24u "Xuva will always be added to your Start menu. Tick the box below if you'd also like a shortcut on your Desktop."
  Pop $0

  ${NSD_CreateCheckbox} 0 32u 100% 12u "Create a Desktop shortcut for Xuva"
  Pop $XuvaShortcutCheckbox
  ${If} $XuvaCreateDesktopShortcut == "1"
    ${NSD_Check} $XuvaShortcutCheckbox
  ${EndIf}

  nsDialogs::Show
FunctionEnd

Function XuvaShortcutPageLeave
  ${NSD_GetState} $XuvaShortcutCheckbox $0
  ${If} $0 == ${BST_CHECKED}
    StrCpy $XuvaCreateDesktopShortcut "1"
  ${Else}
    StrCpy $XuvaCreateDesktopShortcut "0"
  ${EndIf}
FunctionEnd

!endif ; BUILD_UNINSTALLER guard for the install-only Page functions

; ── Install / Uninstall hooks ───────────────────────────────────────────────

!macro customInstall
  !insertmacro XuvaAddFirewallRules

  ; createDesktopShortcut is disabled in package.json so electron-builder
  ; doesn't unconditionally create the shortcut. Honour the user's checkbox
  ; choice from the custom page above instead. ${SHORTCUT_NAME} and $appExe
  ; are populated by electron-builder's NSIS macros.
  ${If} $XuvaCreateDesktopShortcut == "1"
    DetailPrint "Creating Desktop shortcut for Xuva..."
    CreateShortCut "$DESKTOP\${SHORTCUT_NAME}.lnk" "$appExe" "" "$appExe" 0 "" "" "${APP_DESCRIPTION}"
    ClearErrors
    WinShell::SetLnkAUMI "$DESKTOP\${SHORTCUT_NAME}.lnk" "${APP_ID}"
  ${Else}
    DetailPrint "Skipping Desktop shortcut (unchecked by user)."
  ${EndIf}

  ; ── Install receipt ──────────────────────────────────────────────────────
  ; Write a small JSON receipt to ProgramData so we can observe whether this
  ; hook actually ran and at what privilege level. The previous "install over
  ; the top" verification was useless because the user had already added a
  ; firewall rule manually — checking the rule's existence couldn't
  ; distinguish "installer fired customInstall" from "user's manual rule
  ; survived". The receipt's presence + content unambiguously says the hook
  ; ran. The server can also read it later to surface install diagnostics in
  ; the settings UI.
  ;
  ; ProgramData persists across upgrades and is writeable by an elevated
  ; installer. We write a stable filename (no version in path) so the latest
  ; receipt always overwrites prior runs.
  DetailPrint "Writing install receipt to ProgramData..."
  ; Use %PROGRAMDATA% env var to get the canonical path (C:\ProgramData).
  ; $APPDATA relative traversal (e.g. $APPDATA\..\..) resolves incorrectly
  ; for per-user installs because $APPDATA is the Roaming profile dir and
  ; only two levels up from there is the user's home, not the machine-wide
  ; ProgramData root.
  ReadEnvStr $6 "PROGRAMDATA"
  ${If} $6 == ""
    StrCpy $6 "C:\ProgramData"
  ${EndIf}
  CreateDirectory "$6\Xuva"
  ; Capture admin state and current timestamp via NSIS macros + system call.
  ${If} ${UAC_IsAdmin}
    StrCpy $9 "true"
  ${Else}
    StrCpy $9 "false"
  ${EndIf}
  ; %SYSTEMTIME% style ISO timestamp — we don't have ISO out of the box in
  ; NSIS but the order y/m/d h:m is sortable and good enough for debugging.
  System::Call 'kernel32::GetSystemTime(p)p.r8'
  ; Bail on the timestamp if the call shape varies — receipt is best-effort.
  ClearErrors
  FileOpen $7 "$6\Xuva\install-receipt.json" w
  ${IfNot} ${Errors}
    FileWrite $7 '{$\r$\n'
    FileWrite $7 '  "schema": 1,$\r$\n'
    FileWrite $7 '  "installerVersion": "${VERSION}",$\r$\n'
    FileWrite $7 '  "installDir": "$INSTDIR",$\r$\n'
    FileWrite $7 '  "ranAsAdmin": $9,$\r$\n'
    FileWrite $7 '  "desktopShortcutRequested": $XuvaCreateDesktopShortcut,$\r$\n'
    FileWrite $7 '  "firewallRulesAttempted": true,$\r$\n'
    FileWrite $7 '  "shortcutName": "${SHORTCUT_NAME}"$\r$\n'
    FileWrite $7 '}$\r$\n'
    FileClose $7
  ${EndIf}
!macroend

!macro customUnInstall
  !insertmacro XuvaRemoveFirewallRules

  ; Best-effort: remove the Desktop shortcut we may have created. Safe to
  ; call even if it doesn't exist (Delete just sets an error flag we ignore).
  Delete "$DESKTOP\${SHORTCUT_NAME}.lnk"
  ClearErrors

  ; Drop the install receipt so a future fresh-install starts with a clean
  ; state and we don't leave forensic data lying around after uninstall.
  ReadEnvStr $0 "PROGRAMDATA"
  ${If} $0 == ""
    StrCpy $0 "C:\ProgramData"
  ${EndIf}
  Delete "$0\Xuva\install-receipt.json"
  ClearErrors
!macroend
