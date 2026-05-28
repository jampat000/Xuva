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

Var XuvaShortcutDialog
Var XuvaShortcutCheckbox
Var XuvaCreateDesktopShortcut

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
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall add rule name="Xuva Server HTTP" dir=in action=allow program="$INSTDIR\resources\runtime\xuva-server.exe" protocol=TCP localport=8097 profile=private,domain enable=yes description="Allow Xuva web and API access from trusted LAN devices."'
  Pop $0
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall add rule name="Xuva Local Discovery mDNS" dir=in action=allow program="$INSTDIR\resources\runtime\xuva-server.exe" protocol=UDP localport=5353 profile=private,domain enable=yes description="Allow Xuva Bonjour/mDNS discovery from trusted LAN devices."'
  Pop $0

  ; Port-scoped fallback rule. Survives install-path changes between releases
  ; — without this, an upgrade that lands in a slightly different folder
  ; would leave the user unable to reach the new binary on the LAN until they
  ; reran the installer. Still scoped to Private+Domain only (not Public).
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

Function XuvaShortcutPageShow
  !insertmacro MUI_HEADER_TEXT "Choose shortcuts" "Decide where Xuva should appear after install."

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
!macroend

!macro customUnInstall
  !insertmacro XuvaRemoveFirewallRules

  ; Best-effort: remove the Desktop shortcut we may have created. Safe to
  ; call even if it doesn't exist (Delete just sets an error flag we ignore).
  Delete "$DESKTOP\${SHORTCUT_NAME}.lnk"
  ClearErrors
!macroend
