# Go RPM Utils

[![Go Reference](https://pkg.go.dev/badge/github.com/sassoftware/go-rpmutils.svg)](https://pkg.go.dev/github.com/sassoftware/go-rpmutils)

go-rpmutils is a library written in [go](http://golang.org) for parsing and extracting content from [RPMs](http://www.rpm.org).

## Overview

go-rpmutils provides a few interfaces for handling RPM packages. There is a highlevel `Rpm` struct that provides access to the RPM header and [CPIO](https://en.wikipedia.org/wiki/Cpio) payload. The CPIO payload can be extracted to a filesystem location via the `ExpandPayload` function or through a Reader interface, similar to the [tar implementation](https://golang.org/pkg/archive/tar/) in the go standard library.

## Example

```go
// Opening a RPM file
f, err := os.Open("foo.rpm")
if err != nil {
    panic(err)
}
rpm, err := rpmutils.ReadRpm(f)
if err != nil {
    panic(err)
}
// Getting metadata
nevra, err := rpm.Header.GetNEVRA()
if err != nil {
    panic(err)
}
fmt.Println(nevra)
provides, err := rpm.Header.GetStrings(rpmutils.PROVIDENAME)
if err != nil {
    panic(err)
}
fmt.Println("Provides:")
for _, p := range provides {
    fmt.Println(p)
}
// Extracting payload
if err := rpm.ExpandPayload("destdir"); err != nil {
    panic(err)
}
```

## Contributing

1. Read contributor agreement
2. Fork it
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Commit your changes (`git commit -a`). Make sure to include a Signed-off-by line per the contributor agreement.
5. Push to the branch (`git push origin my-new-feature`)
6. Create new Pull Request

## License

go-rpmutils is released under the Apache 2.0 license. See [LICENSE](https://github.com/sassoftware/go-rpmutils/blob/master/LICENSE).
