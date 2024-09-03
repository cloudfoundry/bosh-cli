module github.com/cloudfoundry/bosh-cli/v7

go 1.22.0

require (
	code.cloudfoundry.org/clock v1.8.0
	code.cloudfoundry.org/workpool v0.0.0-20240408164905-b6c2fa5a80e4
	github.com/cheggaaa/pb/v3 v3.1.5
	github.com/cloudfoundry/bosh-agent v0.0.61-0.20230301011448-4cfe06c13ad7
	github.com/cloudfoundry/bosh-davcli v0.0.363
	github.com/cloudfoundry/bosh-gcscli v0.0.246
	github.com/cloudfoundry/bosh-s3cli v0.0.314
	github.com/cloudfoundry/bosh-utils v0.0.486
	github.com/cloudfoundry/config-server v0.1.202
	github.com/cloudfoundry/socks5-proxy v0.2.122
	github.com/cppforlife/go-patch v0.2.0
	github.com/cppforlife/go-semi-semantic v0.0.0-20160921010311-576b6af77ae4
	github.com/dustin/go-humanize v1.0.1
	github.com/fatih/color v1.17.0
	github.com/golang/mock v1.6.0
	github.com/gopacket/gopacket v1.2.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/jessevdk/go-flags v1.6.1
	github.com/mattn/go-isatty v0.0.20
	github.com/maxbrunsfeld/counterfeiter/v6 v6.8.1
	github.com/onsi/ginkgo/v2 v2.20.2
	github.com/onsi/gomega v1.34.2
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/spf13/cobra v1.8.1
	github.com/vito/go-interact v1.0.1
	golang.org/x/crypto v0.26.0
	golang.org/x/text v0.17.0
	golang.org/x/tools v0.24.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloud.google.com/go v0.115.1 // indirect
	cloud.google.com/go/auth v0.9.2 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.4 // indirect
	cloud.google.com/go/compute/metadata v0.5.0 // indirect
	cloud.google.com/go/iam v1.2.0 // indirect
	cloud.google.com/go/storage v1.43.0 // indirect
	code.cloudfoundry.org/tlsconfig v0.2.0 // indirect
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/aws/aws-sdk-go v1.55.5 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/charlievieth/fs v0.0.3 // indirect
	github.com/cloudfoundry/go-socks5 v0.0.0-20240831012420-2590b55236ee // indirect
	github.com/creack/pty v1.1.9 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/pprof v0.0.0-20240829160300-da1f7e9f2b25 // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.3 // indirect
	github.com/googleapis/gax-go/v2 v2.13.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/pivotal-cf/paraphernalia v0.0.0-20180203224945-a64ae2051c20 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.54.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.54.0 // indirect
	go.opentelemetry.io/otel v1.29.0 // indirect
	go.opentelemetry.io/otel/metric v1.29.0 // indirect
	go.opentelemetry.io/otel/trace v1.29.0 // indirect
	golang.org/x/mod v0.20.0 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/oauth2 v0.22.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/term v0.23.0 // indirect
	golang.org/x/time v0.6.0 // indirect
	google.golang.org/api v0.195.0 // indirect
	google.golang.org/genproto v0.0.0-20240827150818-7e3bb234dfed // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240827150818-7e3bb234dfed // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/grpc v1.66.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)
