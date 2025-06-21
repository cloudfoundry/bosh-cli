module github.com/cloudfoundry/bosh-cli/v7

go 1.23.0

require (
	code.cloudfoundry.org/clock v1.40.0
	code.cloudfoundry.org/workpool v0.0.0-20241210013132-62cbb12e809b
	github.com/cheggaaa/pb/v3 v3.1.7
	github.com/cloudfoundry/bosh-agent/v2 v2.753.0
	github.com/cloudfoundry/bosh-davcli v0.0.423
	github.com/cloudfoundry/bosh-gcscli v0.0.307
	github.com/cloudfoundry/bosh-s3cli v0.0.367
	github.com/cloudfoundry/bosh-utils v0.0.545
	github.com/cloudfoundry/config-server v0.1.248
	github.com/cloudfoundry/socks5-proxy v0.2.154
	github.com/cppforlife/go-patch v0.2.0
	github.com/cppforlife/go-semi-semantic v0.0.0-20160921010311-576b6af77ae4
	github.com/dustin/go-humanize v1.0.1
	github.com/fatih/color v1.18.0
	github.com/golang/mock v1.6.0
	github.com/gopacket/gopacket v1.3.1
	github.com/hashicorp/go-multierror v1.1.1
	github.com/jessevdk/go-flags v1.6.1
	github.com/mattn/go-isatty v0.0.20
	github.com/maxbrunsfeld/counterfeiter/v6 v6.11.2
	github.com/onsi/ginkgo/v2 v2.23.4
	github.com/onsi/gomega v1.37.0
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/spf13/cobra v1.9.1
	github.com/vito/go-interact v1.0.2
	golang.org/x/crypto v0.39.0
	golang.org/x/text v0.26.0
	golang.org/x/tools v0.34.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cel.dev/expr v0.24.0 // indirect
	cloud.google.com/go v0.121.2 // indirect
	cloud.google.com/go/auth v0.16.2 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.7.0 // indirect
	cloud.google.com/go/iam v1.5.2 // indirect
	cloud.google.com/go/monitoring v1.24.2 // indirect
	cloud.google.com/go/storage v1.55.0 // indirect
	code.cloudfoundry.org/tlsconfig v0.29.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.29.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.53.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.53.0 // indirect
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/aws/aws-sdk-go v1.55.7 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/charlievieth/fs v0.0.3 // indirect
	github.com/cloudfoundry/go-socks5 v0.0.0-20250423223041-4ad5fea42851 // indirect
	github.com/cncf/xds/go v0.0.0-20250501225837-2ac532fd4443 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.32.4 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-jose/go-jose/v4 v4.0.5 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20250607225305-033d6d78b36a // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.14.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/pivotal-cf/paraphernalia v0.0.0-20180203224945-a64ae2051c20 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/spiffe/go-spiffe/v2 v2.5.0 // indirect
	github.com/zeebo/errs v1.4.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.36.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.61.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.61.0 // indirect
	go.opentelemetry.io/otel v1.36.0 // indirect
	go.opentelemetry.io/otel/metric v1.36.0 // indirect
	go.opentelemetry.io/otel/sdk v1.36.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.36.0 // indirect
	go.opentelemetry.io/otel/trace v1.36.0 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	golang.org/x/mod v0.25.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/term v0.32.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	google.golang.org/api v0.238.0 // indirect
	google.golang.org/genproto v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/grpc v1.73.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
