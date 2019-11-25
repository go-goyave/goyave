param (
    [Parameter(Mandatory=$true)][string]$projectName
)
New-Item -ItemType Directory -Path $projectName
Invoke-WebRequest "https://github.com/System-Glitch/goyave-template/archive/master.zip" -OutFile "temp.zip"
Expand-Archive -Path "temp.zip" -DestinationPath $projectName
Remove-Item -Path "temp.zip"
$include = ("*.go","*.mod","*.json")
Get-ChildItem -Path $projectName -Include $include -Recurse | ForEach-Object  { 
    (Get-Content $_).Replace("goyave_template", $projectName) | Set-Content $_ 
}