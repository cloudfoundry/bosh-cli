trap {
  write-error $_
  exit 1
}

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
$releases = Invoke-RestMethod 'https://go.dev/dl/?mode=json'
$release = $releases | Where-Object { $_.stable -and $_.version -match '^go1\.' } | Select-Object -First 1
$installer = $release.files | Where-Object { $_.os -eq 'windows' -and $_.arch -eq 'amd64' -and $_.kind -eq 'installer' } | Select-Object -First 1
if ($null -eq $installer -or $installer.sha256 -notmatch '^[0-9a-fA-F]{64}$' -or $installer.filename -notmatch '^go1\.[0-9.]+\.windows-amd64\.msi$') {
  throw 'Go release metadata does not contain a valid Windows installer and SHA256.'
}
$GO_VERSION = $release.version

Write-Host "Checking if Go needs to be installed or updated..."
if ((Get-Command 'go.exe' -ErrorAction SilentlyContinue) -ne $null) {
  $version = "$(go.exe version)"
  if ($version -match "go version $GO_VERSION windows\/amd64") {
    Write-Host "Golang $GO_VERSION already installed, skipping download."
    exit 0
  }
}

Write-Host "Installing $GO_VERSION"
Invoke-WebRequest "https://go.dev/dl/$($installer.filename)" -UseBasicParsing -OutFile go.msi
$actualHash = (Get-FileHash -LiteralPath go.msi -Algorithm SHA256).Hash
if ($actualHash -ne $installer.sha256) {
  Remove-Item -LiteralPath go.msi -Force
  throw 'Go installer SHA256 does not match the release metadata.'
}
$p = Start-Process -FilePath 'msiexec' -ArgumentList '/passive /norestart /i go.msi' -Wait -PassThru
if ($p.ExitCode -ne 0) {
  throw "Golang MSI installation process returned error code: $($p.ExitCode)"
}
Write-Host "Successfully installed go version: $(go version)"
