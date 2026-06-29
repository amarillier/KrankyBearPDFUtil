; Inno Setup script for KrankyBear PDF — ONE Windows installer for all three binaries.
;   pdfgui.exe    - primary GUI (Start Menu / optional desktop shortcut)
;   pdfutil.exe   - CLI (no shortcut; {app} can be added to PATH)
;   pdfviewer.exe - bundled helper viewer (no shortcut for now)
;
; Build on Windows after compile-windows.ps1 has produced bin\*-windows-amd64.exe:
;   "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" Inno\KrankyBearPDF.iss
; Version is kept in lock-step by ../setver.sh (#define MyAppVersion below).

#define MyAppName "KrankyBearPDF"
#define MyAppNiceName "KrankyBear PDF"
#define MyAppVersion "0.3.1"
#define MyAppPublisher "Allan Marillier, 2025-"
#define MyAppURL "https://github.com/amarillier/KrankyBearPDF"
#define MyGuiExe "pdfgui.exe"
#define MyCliExe "pdfutil.exe"
#define MyViewerExe "pdfviewer.exe"

[Setup]
; NOTE: AppId uniquely identifies this application. Do not reuse for other apps.
; Unique to KrankyBearPDF (a shared GUID makes Inno reuse another app's install dir).
AppId={{E9903AB2-F1F1-4DE7-A022-D0400A6AE72E}
AppName={#MyAppNiceName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
ArchitecturesAllowed=x64 arm64
ArchitecturesInstallIn64BitMode=x64 arm64
UninstallDisplayIcon={app}\{#MyGuiExe}
DefaultDirName={autopf}\{#MyAppName}
DisableDirPage=yes
DisableProgramGroupPage=yes
ChangesAssociations=yes
ChangesEnvironment=yes
LicenseFile=..\LICENSE
PrivilegesRequiredOverridesAllowed=dialog
OutputDir=..\installers
OutputBaseFilename=KrankyBearPDFSetup_{#MyAppVersion}
; No .ico ships in the repo; the GUI exe carries its embedded icon (go-winres).
; To brand the installer itself, generate assets\images\KrankyBearBeanieMultiColor.ico
; and uncomment the next line.
;SetupIconFile=..\assets\images\KrankyBearBeanieMultiColor.ico
Compression=lzma
SolidCompression=yes
WizardStyle=modern

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "addtopath"; Description: "Add install folder to PATH (lets you run pdfutil / pdfviewer from any terminal)"; GroupDescription: "Command line:"; Flags: unchecked

[Files]
; GUI is the primary app; CLI and viewer ride along in the same folder.
Source: "..\bin\pdfgui-windows-amd64.exe";    DestDir: "{app}"; DestName: "{#MyGuiExe}";    Flags: ignoreversion
Source: "..\bin\pdfutil-windows-amd64.exe";   DestDir: "{app}"; DestName: "{#MyCliExe}";    Flags: ignoreversion
Source: "..\bin\pdfviewer-windows-amd64.exe"; DestDir: "{app}"; DestName: "{#MyViewerExe}"; Flags: ignoreversion skipifsourcedoesntexist
Source: "..\ReleaseNotes.txt";                DestDir: "{app}"; Flags: isreadme
Source: "..\LICENSE";                         DestDir: "{app}"; DestName: "License.txt"; Flags: ignoreversion

[Registry]
; Add {app} to the (system or user) PATH when the addtopath task is selected.
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; \
    ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; \
    Tasks: addtopath; Check: NeedsAddPath('{app}'); Flags: preservestringtype

[Icons]
; Shortcut for the GUI only.
Name: "{autoprograms}\{#MyAppNiceName}"; Filename: "{app}\{#MyGuiExe}"
Name: "{autodesktop}\{#MyAppNiceName}";  Filename: "{app}\{#MyGuiExe}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#MyGuiExe}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppNiceName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent

[UninstallRun]
Filename: "{cmd}"; Parameters: "/C ""taskkill /im {#MyGuiExe} /f /t"; Flags: runhidden; RunOnceId: "KillGui"
Filename: "{cmd}"; Parameters: "/C ""taskkill /im {#MyViewerExe} /f /t"; Flags: runhidden; RunOnceId: "KillViewer"

[Code]
// Only append {app} to PATH if it isn't already present, to avoid duplicate entries.
function NeedsAddPath(Param: string): Boolean;
var
  OrigPath: string;
begin
  if not RegQueryStringValue(HKLM, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment', 'Path', OrigPath) then
  begin
    Result := True;
    exit;
  end;
  Result := Pos(';' + Uppercase(ExpandConstant(Param)) + ';', ';' + Uppercase(OrigPath) + ';') = 0;
end;
