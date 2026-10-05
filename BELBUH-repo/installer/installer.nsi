Unicode true
!define MULTIUSER_EXECUTIONLEVEL Highest
!define MULTIUSER_INSTALLMODE_DEFAULT_CURRENTUSER
!define MULTIUSER_INSTALLMODE_COMMANDLINE
!define MULTIUSER_MUI
!define MULTIUSER_INSTALLMODE_INSTDIR_REGISTRY_KEY "Software\CSL\BELHUB"
!define MULTIUSER_INSTALLMODE_INSTDIR_REGISTRY_VALUENAME "InstallDir"
!include "MultiUser.nsh"
!include "MUI2.nsh"

Name "BELHUB 3.2"
OutFile "BELHUB-3.2-Setup.exe"
BrandingText "ООО Компьютерная безопасность"
Icon "assets\BELHUB.ico"
UninstallIcon "assets\BELHUB.ico"
ShowInstDetails show
ShowUninstDetails show

!define MUI_ABORTWARNING
!define MUI_ICON "assets\BELHUB.ico"
!define MUI_UNICON "assets\BELHUB.ico"
!insertmacro MUI_PAGE_WELCOME
!insertmacro MULTIUSER_PAGE_INSTALLMODE
!insertmacro MUI_PAGE_COMPONENTS
!insertmacro MUI_PAGE_INSTFILES
!define MUI_FINISHPAGE_RUN "$INSTDIR\BELHUB-3.2.exe"
!define MUI_FINISHPAGE_RUN_TEXT "Запустить BELHUB 3.2"
!insertmacro MUI_PAGE_FINISH
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_LANGUAGE "Russian"

Function .onInit
  !insertmacro MULTIUSER_INIT
FunctionEnd

Section "BELHUB 3.2 (обязательно)" SEC_MAIN
  SectionIn RO
  ${If} $MultiUser.InstallMode == "AllUsers"
    StrCpy $INSTDIR "$PROGRAMFILES64\BELHUB"
  ${Else}
    StrCpy $INSTDIR "$LOCALAPPDATA\Programs\BELHUB"
  ${EndIf}
  SetOutPath "$INSTDIR"
  File "BELHUB-3.2.exe"
  File "BELHUB-Widget.exe"
  File "BELHUB-3.2.html"
  File "widget.html"
  CreateDirectory "$SMPROGRAMS\BELHUB"
  CreateShortcut "$SMPROGRAMS\BELHUB\BELHUB 3.2.lnk" "$INSTDIR\BELHUB-3.2.exe" "" "$INSTDIR\BELHUB-3.2.exe" 0
  CreateShortcut "$SMPROGRAMS\BELHUB\BELHUB мини-виджет.lnk" "$INSTDIR\BELHUB-Widget.exe" "" "$INSTDIR\BELHUB-Widget.exe" 0
  CreateShortcut "$SMPROGRAMS\BELHUB\Удалить BELHUB.lnk" "$INSTDIR\Uninstall.exe"
  CreateShortcut "$DESKTOP\BELHUB 3.2.lnk" "$INSTDIR\BELHUB-3.2.exe" "" "$INSTDIR\BELHUB-3.2.exe" 0
  CreateShortcut "$DESKTOP\BELHUB Виджет.lnk" "$INSTDIR\BELHUB-Widget.exe" "" "$INSTDIR\BELHUB-Widget.exe" 0
  WriteRegStr HKCU "Software\Classes\belhub" "" "URL:BELHUB Local Protocol"
  WriteRegStr HKCU "Software\Classes\belhub" "URL Protocol" ""
  WriteRegStr HKCU "Software\Classes\belhub\DefaultIcon" "" "$INSTDIR\BELHUB-3.2.exe,0"
  WriteRegStr HKCU "Software\Classes\belhub\shell\open\command" "" '"$INSTDIR\BELHUB-3.2.exe" "%1"'
  WriteUninstaller "$INSTDIR\Uninstall.exe"
  WriteRegStr SHCTX "Software\CSL\BELHUB" "InstallDir" "$INSTDIR"
  WriteRegStr SHCTX "Software\Microsoft\Windows\CurrentVersion\Uninstall\BELHUB" "DisplayName" "BELHUB 3.2"
  WriteRegStr SHCTX "Software\Microsoft\Windows\CurrentVersion\Uninstall\BELHUB" "Publisher" "ООО Компьютерная безопасность"
  WriteRegStr SHCTX "Software\Microsoft\Windows\CurrentVersion\Uninstall\BELHUB" "DisplayVersion" "3.2"
  WriteRegStr SHCTX "Software\Microsoft\Windows\CurrentVersion\Uninstall\BELHUB" "DisplayIcon" "$INSTDIR\BELHUB-3.2.exe"
  WriteRegStr SHCTX "Software\Microsoft\Windows\CurrentVersion\Uninstall\BELHUB" "UninstallString" "$INSTDIR\Uninstall.exe"
SectionEnd

Section /o "Запускать мини-виджет вместе с Windows" SEC_AUTO
  WriteRegStr SHCTX "Software\Microsoft\Windows\CurrentVersion\Run" "BELHUB Widget" '"$INSTDIR\BELHUB-Widget.exe"'
SectionEnd

Section "Uninstall"
  DeleteRegValue SHCTX "Software\Microsoft\Windows\CurrentVersion\Run" "BELHUB Widget"
  DeleteRegKey HKCU "Software\Classes\belhub"
  DeleteRegKey SHCTX "Software\Microsoft\Windows\CurrentVersion\Uninstall\BELHUB"
  DeleteRegKey SHCTX "Software\CSL\BELHUB"
  Delete "$DESKTOP\BELHUB 3.2.lnk"
  Delete "$DESKTOP\BELHUB Виджет.lnk"
  RMDir /r "$SMPROGRAMS\BELHUB"
  Delete "$INSTDIR\BELHUB-3.2.exe"
  Delete "$INSTDIR\BELHUB-Widget.exe"
  Delete "$INSTDIR\BELHUB-3.2.html"
  Delete "$INSTDIR\widget.html"
  Delete "$INSTDIR\Uninstall.exe"
  RMDir "$INSTDIR"
SectionEnd

Function un.onInit
  !insertmacro MULTIUSER_UNINIT
FunctionEnd
