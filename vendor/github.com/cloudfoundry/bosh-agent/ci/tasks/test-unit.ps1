trap {
  write-error $_
  exit 1
}

$env:GOPATH = Join-Path -Path $PWD "gopath"
$env:PATH = $env:GOPATH + "/bin;C:/go/bin;C:/var/vcap/bosh/bin;" + $env:PATH

cd $env:GOPATH/src/github.com/cloudfoundry/bosh-agent

if ((Get-Command "go.exe" -ErrorAction SilentlyContinue) -eq $null)
{
  Write-Host "Installing Go 1.7.3!"
  Invoke-WebRequest https://storage.googleapis.com/golang/go1.7.3.windows-amd64.msi -OutFile go.msi

  $p = Start-Process -FilePath "msiexec" -ArgumentList "/passive /norestart /i go.msi" -Wait -PassThru

  if($p.ExitCode -ne 0)
  {
    throw "Golang MSI installation process returned error code: $($p.ExitCode)"
  }
  Write-Host "Go is installed!"
}

# Change constant for unit test process labels to prevent collisions with
# bosh deployed Job Supervisor
$file = Join-Path -Path $PWD jobsupervisor/windows_job_supervisor.go
(Get-Content $file).replace('serviceDescription = "vcap"', 'serviceDescription = "vcap_test"') | Set-Content $file

go.exe install github.com/cloudfoundry/bosh-agent/vendor/github.com/onsi/ginkgo/ginkgo
ginkgo.exe -r -race -keepGoing -skipPackage="integration,vendor"
if ($LastExitCode -ne 0)
{
    Write-Error $_
    exit 1
}
