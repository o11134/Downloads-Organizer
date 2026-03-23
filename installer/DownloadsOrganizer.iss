; Inno Setup script for Downloads Organizer

[Setup]
AppId={{2B9D9278-1021-4E95-A29A-7E12042A856A}
AppName=Downloads Organizer
AppVersion=1.0.0
AppPublisher=Downloads Organizer
DefaultDirName={localappdata}\Programs\DownloadsOrganizer
DefaultGroupName=Downloads Organizer
UninstallDisplayIcon={app}\DownloadsOrganizer.exe
Compression=lzma
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=lowest
OutputDir=..\dist\installer
OutputBaseFilename=DownloadsOrganizerSetup

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "startup"; Description: "Start with Windows"; GroupDescription: "Options:"; Flags: unchecked

[Files]
Source: "..\DownloadsOrganizer.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\Downloads Organizer"; Filename: "{app}\DownloadsOrganizer.exe"
Name: "{group}\Uninstall Downloads Organizer"; Filename: "{uninstallexe}"

[Run]
Filename: "{app}\DownloadsOrganizer.exe"; Description: "Launch Downloads Organizer"; Flags: nowait postinstall skipifsilent

[Registry]
Root: HKCU; Subkey: "Software\Microsoft\Windows\CurrentVersion\Run"; ValueType: string; ValueName: "DownloadsOrganizer"; ValueData: """{app}\DownloadsOrganizer.exe"""; Flags: uninsdeletevalue; Tasks: startup
