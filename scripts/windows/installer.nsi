; GoPCA Windows Installer Script
; Requires NSIS 3.0 or later

;--------------------------------
; General Settings

!define PRODUCT_NAME "GoPCA"
!define PRODUCT_PUBLISHER "bitjungle"
!define PRODUCT_WEB_SITE "https://github.com/bitjungle/gopca"
!define PRODUCT_UNINST_KEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}"
!define PRODUCT_UNINST_ROOT_KEY "HKLM"

; Version will be set from command line via /DVERSION=x.x.x
!ifndef VERSION
  !define VERSION "0.0.0"
!endif

;--------------------------------
; Includes

!include "MUI2.nsh"
!include "nsDialogs.nsh"
!include "LogicLib.nsh"
!include "WinVer.nsh"
!include "x64.nsh"

;--------------------------------
; Interface Settings

!define MUI_ABORTWARNING
!define MUI_ICON "${NSISDIR}\Contrib\Graphics\Icons\orange-install.ico"
!define MUI_UNICON "${NSISDIR}\Contrib\Graphics\Icons\orange-uninstall.ico"
!define MUI_HEADERIMAGE
!define MUI_HEADERIMAGE_RIGHT
!define MUI_HEADERIMAGE_BITMAP "${NSISDIR}\Contrib\Graphics\Header\orange-r.bmp"
!define MUI_WELCOMEFINISHPAGE_BITMAP "${NSISDIR}\Contrib\Graphics\Wizard\orange.bmp"

;--------------------------------
; Installer Pages

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "..\..\LICENSE"
!insertmacro MUI_PAGE_COMPONENTS
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

;--------------------------------
; Uninstaller Pages

!insertmacro MUI_UNPAGE_WELCOME
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

;--------------------------------
; Languages

!insertmacro MUI_LANGUAGE "English"

;--------------------------------
; Installer Information

Name "${PRODUCT_NAME} ${VERSION}"
OutFile "..\..\build\windows-installer\GoPCA-Setup-v${VERSION}.exe"
InstallDir "$PROGRAMFILES64\${PRODUCT_NAME}"
InstallDirRegKey HKLM "Software\${PRODUCT_NAME}" "Install_Dir"
RequestExecutionLevel admin
ShowInstDetails show
ShowUnInstDetails show

;--------------------------------
; Version Information

; VIProductVersion requires X.X.X.X format
; Parse version components - handle both regular (0.9.6) and test versions (0.9.6-test)
!searchparse /noerrors "${VERSION}" "" VersionMajor "." VersionMinor "." VersionPatch "-" VersionSuffix

; If no hyphen found (regular version), parse without suffix
!ifndef VersionPatch
  !searchparse "${VERSION}" "" VersionMajor "." VersionMinor "." VersionPatch
!endif

; Always provide VIProductVersion in X.X.X.X format (required by NSIS)
VIProductVersion "${VersionMajor}.${VersionMinor}.${VersionPatch}.0"
VIAddVersionKey "ProductName" "${PRODUCT_NAME}"
VIAddVersionKey "ProductVersion" "${VERSION}"
VIAddVersionKey "CompanyName" "${PRODUCT_PUBLISHER}"
VIAddVersionKey "LegalCopyright" "Copyright (c) 2025 ${PRODUCT_PUBLISHER}"
VIAddVersionKey "FileDescription" "${PRODUCT_NAME} Installer"
VIAddVersionKey "FileVersion" "${VERSION}"

;--------------------------------
; Component Variables

Var AddToPath

;--------------------------------
; Functions

Function .onInit
  ; Check if 64-bit Windows
  ${If} ${RunningX64}
    SetRegView 64
  ${Else}
    MessageBox MB_OK|MB_ICONSTOP "This installer requires 64-bit Windows."
    Abort
  ${EndIf}
  
  ; Initialize variables
  StrCpy $AddToPath "0"
FunctionEnd

Function AddToSystemPath
  ; Add bin directory to PATH
  Push "$INSTDIR\bin"
  Call AddToPath
FunctionEnd

Function un.RemoveFromSystemPath
  ; Remove bin directory from PATH
  Push "$INSTDIR\bin"
  Call un.RemoveFromPath
FunctionEnd

; Function to add directory to PATH
Function AddToPath
  Exch $0
  Push $1
  Push $2
  Push $3
  
  ; Get current PATH
  ReadRegStr $1 HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "PATH"
  
  ; Check if directory already in PATH
  Push "$1;"
  Push "$0;"
  Call StrStr
  Pop $2
  StrCmp $2 "" 0 already_in_path
  Push "$1;"
  Push "$0\;"
  Call StrStr
  Pop $2
  StrCmp $2 "" 0 already_in_path
  
  ; Add to PATH
  StrCpy $3 "$1;$0"
  WriteRegExpandStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "PATH" "$3"
  
  ; Notify system of change
  SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=5000
  
  already_in_path:
  Pop $3
  Pop $2
  Pop $1
  Pop $0
FunctionEnd

; Function to remove directory from PATH
Function un.RemoveFromPath
  Exch $0
  Push $1
  Push $2
  Push $3
  Push $4
  Push $5
  Push $6
  
  ; Get current PATH
  ReadRegStr $1 HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "PATH"
  StrCpy $5 $1 1 -1
  StrCmp $5 ";" +2
  StrCpy $1 "$1;"
  Push $1
  Push "$0;"
  Call un.StrStr
  Pop $2
  StrCmp $2 "" unRemoveFromPath_done
  
  ; Remove from PATH
  StrLen $3 "$0;"
  StrLen $4 $2
  StrCpy $5 $1 -$4
  StrCpy $6 $2 "" $3
  StrCpy $3 "$5$6"
  StrCpy $5 $3 1 -1
  StrCmp $5 ";" 0 +2
  StrCpy $3 $3 -1
  WriteRegExpandStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "PATH" "$3"
  
  ; Notify system of change
  SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=5000
  
  unRemoveFromPath_done:
  Pop $6
  Pop $5
  Pop $4
  Pop $3
  Pop $2
  Pop $1
  Pop $0
FunctionEnd

; String search function
Function StrStr
  Exch $R1
  Exch
  Exch $R2
  Push $R3
  Push $R4
  Push $R5
  StrLen $R3 $R1
  StrCpy $R4 0
  
  loop:
    StrCpy $R5 $R2 $R3 $R4
    StrCmp $R5 $R1 done
    StrCmp $R5 "" done
    IntOp $R4 $R4 + 1
    Goto loop
  
  done:
    StrCpy $R1 $R2 "" $R4
    Pop $R5
    Pop $R4
    Pop $R3
    Pop $R2
    Exch $R1
FunctionEnd

Function un.StrStr
  Exch $R1
  Exch
  Exch $R2
  Push $R3
  Push $R4
  Push $R5
  StrLen $R3 $R1
  StrCpy $R4 0
  
  loop:
    StrCpy $R5 $R2 $R3 $R4
    StrCmp $R5 $R1 done
    StrCmp $R5 "" done
    IntOp $R4 $R4 + 1
    Goto loop
  
  done:
    StrCpy $R1 $R2 "" $R4
    Pop $R5
    Pop $R4
    Pop $R3
    Pop $R2
    Exch $R1
FunctionEnd

;--------------------------------
; Installer Sections

Section "GoPCA Desktop Application" SEC_GOPCA_DESKTOP
  SectionIn RO ; Required section
  
  SetOutPath "$INSTDIR"
  
  ; Copy GoPCA Desktop executable (required)
  File "..\..\build\windows-installer\GoPCA.exe"
  
  ; Create Start Menu shortcut
  CreateDirectory "$SMPROGRAMS\${PRODUCT_NAME}"
  CreateShortCut "$SMPROGRAMS\${PRODUCT_NAME}\GoPCA Desktop.lnk" "$INSTDIR\GoPCA.exe"
  
  ; Create Desktop shortcut (optional)
  CreateShortCut "$DESKTOP\GoPCA Desktop.lnk" "$INSTDIR\GoPCA.exe"
  
  ; Write installation info to registry
  WriteRegStr HKLM "Software\${PRODUCT_NAME}" "Install_Dir" "$INSTDIR"
  WriteRegStr HKLM "Software\${PRODUCT_NAME}" "Version" "${VERSION}"
  
  ; Write uninstaller info
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "DisplayName" "${PRODUCT_NAME}"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "UninstallString" "$INSTDIR\uninstall.exe"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "DisplayIcon" "$INSTDIR\GoPCA.exe"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "DisplayVersion" "${VERSION}"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "Publisher" "${PRODUCT_PUBLISHER}"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "URLInfoAbout" "${PRODUCT_WEB_SITE}"
  
  ; Estimate size (simplified - use a fixed estimate)
  WriteRegDWORD ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "EstimatedSize" "50000"
  
  ; Create uninstaller
  WriteUninstaller "$INSTDIR\uninstall.exe"
SectionEnd

Section "GoCSV Editor" SEC_GOCSV
  SectionIn RO ; Required section
  
  SetOutPath "$INSTDIR"
  
  ; Copy GoCSV executable (required)
  File "..\..\build\windows-installer\GoCSV.exe"
  
  ; Create Start Menu shortcut
  CreateShortCut "$SMPROGRAMS\${PRODUCT_NAME}\GoCSV Editor.lnk" "$INSTDIR\GoCSV.exe"
  
  ; Optional: Register CSV file association
  ; WriteRegStr HKCR ".csv" "" "GoCSV.Document"
  ; WriteRegStr HKCR "GoCSV.Document" "" "CSV Document"
  ; WriteRegStr HKCR "GoCSV.Document\DefaultIcon" "" "$INSTDIR\GoCSV.exe,0"
  ; WriteRegStr HKCR "GoCSV.Document\shell\open\command" "" '"$INSTDIR\GoCSV.exe" "%1"'
SectionEnd

Section "PCA Command Line Tool" SEC_PCA_CLI
  SectionIn RO ; Required section
  
  SetOutPath "$INSTDIR\bin"
  
  ; Copy PCA CLI executable (required)
  File /oname=pca.exe "..\..\build\windows-installer\pca-windows-amd64.exe"
  
  ; Set flag to add to PATH
  StrCpy $AddToPath "1"
SectionEnd

Section "Add CLI to System PATH" SEC_ADD_PATH
  ${If} $AddToPath == "1"
    Call AddToSystemPath
    
    ; Create batch file for easy CLI access
    FileOpen $0 "$INSTDIR\bin\pca.bat" w
    FileWrite $0 "@echo off$\r$\n"
    FileWrite $0 '"$INSTDIR\bin\pca.exe" %*$\r$\n'
    FileClose $0
  ${EndIf}
SectionEnd

Section "Start Menu Shortcuts" SEC_SHORTCUTS
  CreateShortCut "$SMPROGRAMS\${PRODUCT_NAME}\Uninstall ${PRODUCT_NAME}.lnk" "$INSTDIR\uninstall.exe"
  CreateShortCut "$SMPROGRAMS\${PRODUCT_NAME}\${PRODUCT_NAME} Website.lnk" "${PRODUCT_WEB_SITE}"
SectionEnd

;--------------------------------
; Section Descriptions

!insertmacro MUI_FUNCTION_DESCRIPTION_BEGIN
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_GOPCA_DESKTOP} "GoPCA Desktop - Interactive PCA analysis application (Required)"
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_GOCSV} "GoCSV Desktop - CSV file manipulation and data preparation tool (Required)"
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_PCA_CLI} "pca CLI - Command-line interface for PCA analysis and automation (Required)"
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_ADD_PATH} "Add the pca CLI to your system PATH for easy command-line access"
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_SHORTCUTS} "Create shortcuts in the Start Menu"
!insertmacro MUI_FUNCTION_DESCRIPTION_END

;--------------------------------
; Uninstaller Section

Section "Uninstall"
  ; Remove from PATH if it was added
  Call un.RemoveFromSystemPath
  
  ; Delete executables
  Delete "$INSTDIR\GoPCA.exe"
  Delete "$INSTDIR\GoCSV.exe"
  Delete "$INSTDIR\bin\pca.exe"
  Delete "$INSTDIR\bin\pca.bat"
  Delete "$INSTDIR\uninstall.exe"
  
  ; Remove directories
  RMDir "$INSTDIR\bin"
  RMDir "$INSTDIR"
  
  ; Delete shortcuts
  Delete "$DESKTOP\GoPCA Desktop.lnk"
  Delete "$SMPROGRAMS\${PRODUCT_NAME}\*.lnk"
  RMDir "$SMPROGRAMS\${PRODUCT_NAME}"
  
  ; Delete registry keys
  DeleteRegKey HKLM "Software\${PRODUCT_NAME}"
  DeleteRegKey ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}"
  
  ; Optional: Remove file associations
  ; DeleteRegKey HKCR ".csv"
  ; DeleteRegKey HKCR "GoCSV.Document"
SectionEnd

