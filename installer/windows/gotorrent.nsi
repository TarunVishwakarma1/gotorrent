; ============================================================================
; GoTorrent — Windows Installer Script
; Built with NSIS (Nullsoft Scriptable Install System)
; ============================================================================

!include "MUI2.nsh"

; ---- Version (passed via /DVERSION=x.y.z on the command line) ----
!ifndef VERSION
  !define VERSION "1.0.0"
!endif

; ---- General Settings ----
Name "GoTorrent ${VERSION}"
OutFile "GoTorrent-${VERSION}-Setup.exe"
Unicode True
InstallDir "$PROGRAMFILES64\GoTorrent"
InstallDirRegKey HKLM "Software\GoTorrent" "InstallDir"
RequestExecutionLevel admin

; ---- MUI Settings ----
!define MUI_ABORTWARNING

; ---- Installer Pages ----
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "..\..\LICENSE"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

; ---- Uninstaller Pages ----
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

; ---- Language ----
!insertmacro MUI_LANGUAGE "English"

; ============================================================================
; Installer
; ============================================================================
Section "GoTorrent" SecMain
  SectionIn RO ; required — cannot be deselected

  SetOutPath "$INSTDIR"
  File "..\..\GoTorrent.exe"

  ; Persist install location
  WriteRegStr HKLM "Software\GoTorrent" "InstallDir" "$INSTDIR"

  ; Create uninstaller
  WriteUninstaller "$INSTDIR\Uninstall.exe"

  ; ---- Start Menu shortcuts ----
  CreateDirectory "$SMPROGRAMS\GoTorrent"
  CreateShortcut "$SMPROGRAMS\GoTorrent\GoTorrent.lnk" "$INSTDIR\GoTorrent.exe"
  CreateShortcut "$SMPROGRAMS\GoTorrent\Uninstall GoTorrent.lnk" "$INSTDIR\Uninstall.exe"

  ; ---- Desktop shortcut ----
  CreateShortcut "$DESKTOP\GoTorrent.lnk" "$INSTDIR\GoTorrent.exe"

  ; ---- Add / Remove Programs entry ----
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\GoTorrent" \
    "DisplayName" "GoTorrent"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\GoTorrent" \
    "DisplayVersion" "${VERSION}"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\GoTorrent" \
    "Publisher" "Tarun Vishwakarma"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\GoTorrent" \
    "UninstallString" "$INSTDIR\Uninstall.exe"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\GoTorrent" \
    "InstallLocation" "$INSTDIR"
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\GoTorrent" \
    "NoModify" 1
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\GoTorrent" \
    "NoRepair" 1
SectionEnd

; ============================================================================
; Uninstaller
; ============================================================================
Section "Uninstall"
  ; Remove application files
  Delete "$INSTDIR\GoTorrent.exe"
  Delete "$INSTDIR\Uninstall.exe"
  RMDir "$INSTDIR"

  ; Remove shortcuts
  Delete "$SMPROGRAMS\GoTorrent\GoTorrent.lnk"
  Delete "$SMPROGRAMS\GoTorrent\Uninstall GoTorrent.lnk"
  RMDir "$SMPROGRAMS\GoTorrent"
  Delete "$DESKTOP\GoTorrent.lnk"

  ; Remove registry keys
  DeleteRegKey HKLM "Software\GoTorrent"
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\GoTorrent"
SectionEnd
