trap {
  write-error $_
  exit 1
}

$env:GOPATH = Join-Path -Path $PWD "gopath"
$env:PATH = $env:GOPATH + "/bin;C:/go/bin;" + $env:PATH

cd $env:GOPATH/src/github.com/cloudfoundry/bosh-cli

poweshell.exe bin/install-go.ps1

go.exe install github.com/cloudfoundry/bosh-cli/vendor/github.com/onsi/ginkgo/ginkgo
ginkgo.exe -race -trace integration

if ($LastExitCode -ne 0) {
  Write-Error $_
  exit 1
}
