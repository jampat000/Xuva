; Windows Firewall rules for Xuva LAN access and local discovery.
; These rules are scoped to the packaged server executable and avoid Public networks.

!macro XuvaAddFirewallRules
  DetailPrint "Configuring Windows Firewall rules for Xuva..."
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall delete rule name="Xuva Server HTTP" program="$INSTDIR\resources\runtime\xuva-server.exe"'
  Pop $0
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall delete rule name="Xuva Local Discovery mDNS" program="$INSTDIR\resources\runtime\xuva-server.exe"'
  Pop $0
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall add rule name="Xuva Server HTTP" dir=in action=allow program="$INSTDIR\resources\runtime\xuva-server.exe" protocol=TCP localport=8097 profile=private,domain enable=yes description="Allow Xuva web and API access from trusted LAN devices."'
  Pop $0
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall add rule name="Xuva Local Discovery mDNS" dir=in action=allow program="$INSTDIR\resources\runtime\xuva-server.exe" protocol=UDP localport=5353 profile=private,domain enable=yes description="Allow Xuva Bonjour/mDNS discovery from trusted LAN devices."'
  Pop $0
!macroend

!macro XuvaRemoveFirewallRules
  DetailPrint "Removing Windows Firewall rules for Xuva..."
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall delete rule name="Xuva Server HTTP" program="$INSTDIR\resources\runtime\xuva-server.exe"'
  Pop $0
  nsExec::ExecToLog '"$SYSDIR\netsh.exe" advfirewall firewall delete rule name="Xuva Local Discovery mDNS" program="$INSTDIR\resources\runtime\xuva-server.exe"'
  Pop $0
!macroend

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
!macroend

!macro customInstall
  !insertmacro XuvaAddFirewallRules
!macroend

!macro customUnInstall
  !insertmacro XuvaRemoveFirewallRules
!macroend
