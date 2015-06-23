# Migrating from Godeps to Vendor

Restore godeps dependencies into your GOPATH, so that we vendor the same revisions:

```
godep restore
```

Remove Godeps:

```
rm -rf Godeps
```

Remove Godeps from GOPATH in bin/env

Get vendor tool:

```
go get github.com/kardianos/vendor
```

Currently the tool has an [issue](https://github.com/kardianos/vendor/issues/18) handling same dependencies from different packages. Until fix for it is merged pull changes from our fork:

```
goto vendor
git remote add mariash https://github.com/mariash/vendor
git fetch mariash master
git co mariash/master
go build -o $GOPATH/bin/vendor github.com/kardianos/vendor
```

Vendor everything:

```
vendor init
vendor add -status ext
```

Disable govet. Since it will govet internal which we don't want.

Remove internal from test packages in bin/test-unit -skipPackage="acceptance,integration,internal"

We decided to use vendored ginkgo in our CI.

```
vendor add github.com/onsi/ginkgo/...
vendor add github.com/onsi/gomega/...
```

Update install-ginkgo to install from internal dependecy:

```
$bin/go install ./internal/github.com/onsi/ginkgo
```

Clean everything from GOPATH:

```
cd $GOPATH
find src -type d -mindepth 2 -maxdepth 2 | grep -v 'cloudfoundry' | grep -v 'kardianos' | xargs rm -rf
find src -type d -mindepth 3 -maxdepth 3 | grep -v 'cloudfoundry/bosh-init' | grep -v 'kardianos' | xargs rm -rf
```

Run `bin/test`