Unicode true

####
## Please note: Template replacements don't work in this file. They are provided with default defines like
## mentioned underneath.
## If the keyword is not defined, "wails_tools.nsh" will populate them with the values from ProjectInfo.
## If they are defined here, "wails_tools.nsh" will not touch them. This allows to use this project.nsi manually
## from outside of Wails for debugging and development of the installer.
##
## For development first make a wails nsis build to populate the "wails_tools.nsh":
## > wails build --target windows/amd64 --nsis
## Then you can call makensis on this file with specifying the path to your binary:
## For a AMD64 only installer:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app.exe
## For a ARM64 only installer:
## > makensis -DARG_WAILS_ARM64_BINARY=..\..\bin\app.exe
## For a installer with both architectures:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app-amd64.exe -DARG_WAILS_ARM64_BINARY=..\..\bin\app-arm64.exe
####
## The following information is taken from the ProjectInfo file, but they can be overwritten here.
####
## !define INFO_PROJECTNAME    "MyProject" # Default "{{.Name}}"
## !define INFO_COMPANYNAME    "MyCompany" # Default "{{.Info.CompanyName}}"
!define INFO_PRODUCTNAME    "Video Archiver" # Override default to use proper display name
## !define INFO_PRODUCTVERSION "1.0.0"     # Default "{{.Info.ProductVersion}}"
!define INFO_COPYRIGHT      "NotCoffee418" # Default "{{.Info.Copyright}}"
###
## !define PRODUCT_EXECUTABLE  "Application.exe"      # Default "${INFO_PROJECTNAME}.exe"
## !define UNINST_KEY_NAME     "UninstKeyInRegistry"  # Default "${INFO_COMPANYNAME}${INFO_PRODUCTNAME}"
####
!define REQUEST_EXECUTION_LEVEL "user"            # Default "admin"  see also https://nsis.sourceforge.io/Docs/Chapter4.html
####
## Include the wails tools
####
!include "wails_tools.nsh"



## Define custom macro for closing running processes
####
!macro wails.closeRunningProcesses
    DetailPrint "Closing running ${INFO_PRODUCTNAME} processes..."
    
    # Force close any running instances of the application
    # Use /F to force and /T to terminate child processes  
    # Ignore errors (process might not be running)
    ExecWait 'taskkill /F /IM "${PRODUCT_EXECUTABLE}" /T' $0
    
    # Small delay to let processes close cleanly
    Sleep 2000
!macroend

# The version information for this two must consist of 4 parts
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# Enable HiDPI support. https://nsis.sourceforge.io/Reference/ManifestDPIAware
ManifestDPIAware true

!include "MUI.nsh"
!include "FileFunc.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
# !define MUI_WELCOMEFINISHPAGE_BITMAP "resources\leftimage.bmp" #Include this to add a bitmap on the left side of the Welcome Page. Must be a size of 164x314
!define MUI_FINISHPAGE_RUN
!define MUI_FINISHPAGE_RUN_TEXT "Start ${INFO_PRODUCTNAME} UI"
!define MUI_FINISHPAGE_RUN_FUNCTION "LaunchUI"
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_ABORTWARNING # This will warn the user if they exit from the installer.

!insertmacro MUI_PAGE_WELCOME # Welcome to the installer page.
# !insertmacro MUI_PAGE_LICENSE "resources\eula.txt" # Adds a EULA page to the installer
!insertmacro MUI_PAGE_INSTFILES # Installing page.
!insertmacro MUI_PAGE_FINISH # Finished installation page.

!insertmacro MUI_UNPAGE_COMPONENTS # Uninstall components page
!insertmacro MUI_UNPAGE_INSTFILES # Uinstalling page

!insertmacro MUI_LANGUAGE "English" # Set the Language of the installer

## The following two statements can be used to sign the installer and the uninstaller. The path to the binaries are provided in %1
#!uninstfinalize 'signtool --file "%1"'
#!finalize 'signtool --file "%1"'

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe" # Name of the installer's file.
# Define installation directory
InstallDir "$LOCALAPPDATA\videoarchiver" # Install to hardcoded app data folder to match app expectations
ShowInstDetails show # This will always show the installation details.

Function .onInit
   !insertmacro wails.checkArchitecture
FunctionEnd

Function LaunchUI
    # Launch as current user, not admin
    SetOutPath $INSTDIR

    # Launch needs explorer.exe to launch non-admin process from admin installer
    # Also doesn't work without it in de-elevated installer for whatever reason while daemon does.
    # Just keep it like this.
    Exec '"$WINDIR\explorer.exe" "$INSTDIR\${PRODUCT_EXECUTABLE}"'
FunctionEnd

Section
    !insertmacro wails.setShellContext

    !insertmacro wails.closeRunningProcesses

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files

    # Create regular shortcuts
    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    
    # Registry key to start daemon on startup
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "${INFO_PRODUCTNAME} Daemon" '"$INSTDIR\${PRODUCT_EXECUTABLE}" --mode daemon'

    # Start daemon immediately (before the UI, essential)
    # Does not need explorer.exe for whatever reason
    Exec '"$INSTDIR\${PRODUCT_EXECUTABLE}" --mode daemon'


    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    # Custom install override for local user installs
    # REPLACE the wails macro call with direct function call
    # !insertmacro wails.writeUninstaller
    Call WriteUninstallerHKCU
SectionEnd

# Uninstall sections
Section "un.Uninstall Application" SEC_REMOVE_APP
    SectionIn RO # This section is mandatory
    
    !insertmacro wails.setShellContext

    !insertmacro wails.closeRunningProcesses

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}" # Remove the WebView2 DataPath

    # Remove application files, but preserve data files if user chose to keep them
    Push $INSTDIR
    Call un.RemoveApplicationFiles

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"
    # Registry key to start daemon on startup
    DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "${INFO_PRODUCTNAME} Daemon"
    
    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller
SectionEnd

Section /o "un.Remove User Data (config.json and *.sqlite files)" SEC_REMOVE_DATA
    # This section removes user data files (config.json and *.sqlite)
    Delete "$INSTDIR\config.json"
    Delete "$INSTDIR\*.sqlite"
    
    # Remove the installation directory if it's empty now
    RMDir $INSTDIR
SectionEnd

# Function to remove application files while preserving user data
Function un.RemoveApplicationFiles
    Pop $R0 # Installation directory
    
    # Check if user chose to keep data
    SectionGetFlags ${SEC_REMOVE_DATA} $R1
    IntOp $R1 $R1 & ${SF_SELECTED}
    
    # If user data removal is NOT selected, preserve data files
    IntCmp $R1 0 preserve_data remove_all preserve_data
    
    remove_all:
        # Remove everything
        RMDir /r "$R0"
        Goto done
    
    preserve_data:
        # Remove all files except config.json and *.sqlite
        Push "$R0"
        Call un.RemoveAllExceptData
    
    done:
FunctionEnd

# Function to remove all files except data files
Function un.RemoveAllExceptData
    Pop $R0 # Directory path
    
    # Remove the executable and other application files
    Delete "$R0\${PRODUCT_EXECUTABLE}"
    Delete "$R0\uninstall.exe"
    
    # Remove any subdirectories that are not user data
    # Note: We preserve the root directory and any user data files (*.sqlite, config.json)
    # This conservative approach only removes known application files
    
    # The directory will remain with user data files intact
FunctionEnd

# Set default selections for uninstall components
Function un.onInit
    # By default, keep user data (don't select the remove data section)
    # The remove application section is always selected (RO)
FunctionEnd


# ---- MACRO OVERWRITES FOR LOCAL USER ----
# ---- Overriodes wails_tools.nsh since we need to run as local user ----
# Override wails macros to use HKCU instead of HKLM for user-level installs
# Place this AFTER the !include "wails_tools.nsh" line
# Override wails macros to use HKCU instead of HKLM for user-level installs
# Place this AFTER the !include "wails_tools.nsh" line

# Conditionally undefine macros only if they exist
# HKCU Registry Overrides - PUT THIS AT THE VERY END OF YOUR .NSI FILE
# This overwrites the wails macro calls with custom implementations

# Redefine the registry key constant to use a different internal name
!define UNINST_KEY_HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UNINST_KEY_NAME}"

# Replace the macro calls in your sections with direct registry writes
# In your main Section, REPLACE this line:
#   !insertmacro wails.writeUninstaller
# WITH:
#   Call WriteUninstallerHKCU

# In your uninstall section, REPLACE this line:
#   !insertmacro wails.deleteUninstaller  
# WITH:
#   Call DeleteUninstallerHKCU

Function WriteUninstallerHKCU
    WriteUninstaller "$INSTDIR\uninstall.exe"

    WriteRegStr HKCU "${UNINST_KEY_HKCU}" "Publisher" "${INFO_COMPANYNAME}"
    WriteRegStr HKCU "${UNINST_KEY_HKCU}" "DisplayName" "${INFO_PRODUCTNAME}"
    WriteRegStr HKCU "${UNINST_KEY_HKCU}" "DisplayVersion" "${INFO_PRODUCTVERSION}"
    WriteRegStr HKCU "${UNINST_KEY_HKCU}" "DisplayIcon" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    WriteRegStr HKCU "${UNINST_KEY_HKCU}" "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
    WriteRegStr HKCU "${UNINST_KEY_HKCU}" "QuietUninstallString" "$\"$INSTDIR\uninstall.exe$\" /S"

    ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
    IntFmt $0 "0x%08X" $0
    WriteRegDWORD HKCU "${UNINST_KEY_HKCU}" "EstimatedSize" "$0"
FunctionEnd

Function un.DeleteUninstallerHKCU
    Delete "$INSTDIR\uninstall.exe"
    DeleteRegKey HKCU "${UNINST_KEY_HKCU}"
FunctionEnd