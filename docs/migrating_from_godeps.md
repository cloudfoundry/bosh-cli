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

Vendor everything:

```
vendor init
vendor add -status ext
```

Disable govet. Since it will govet internal which we don't want.

Remove internal from test packages in bin/test-unit -skipPackage="acceptance,integration,internal"

We decided to use vendored ginkgo in our CI, so we vendor it explicitly, and re-vendor gomega so that is updates imports in ginkgo.

```
vendor add github.com/onsi/ginkgo/ginkgo
vendor add github.com/onsi/ginkgo/ginkgo/...
vendor add -status ext
```

Update install-ginkgo to install from internal dependecy:

```
$bin/go install ./internal/github.com/onsi/ginkgo/ginkgo
```

Clean everything from GOPATH:

```
cd $GOPATH
find src -type d -mindepth 2 -maxdepth 2 | grep -v 'cloudfoundry' | grep -v 'kardianos' | xargs rm -rf
find src -type d -mindepth 3 -maxdepth 3 | grep -v 'cloudfoundry/bosh-init' | grep -v 'kardianos' | xargs rm -rf
```

Run `bin/test`