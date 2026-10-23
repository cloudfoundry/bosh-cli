package util_test

import (
	"testing"

	"github.com/cloudfoundry/bosh-cli/v7/common/util"
)

func TestURLUserinfoDisplay(t *testing.T) {
	for _, test := range []struct{ input, want string }{
		{"HTTPS://username@host/path", "HTTPS://<redacted>:<redacted>@host/path"},
		{"HTTP://user:password@host/path", "HTTP://<redacted>:<redacted>@host/path"},
		{"https://host/path?user=username@host", "https://host/path?user=username@host"},
		{"https://user:password@one https://name@two", "https://<redacted>:<redacted>@one https://<redacted>:<redacted>@two"},
		{"https://name%40domain:p%3Ass@host", "https://<redacted>:<redacted>@host"},
	} {
		if got := util.RedactBasicAuth(test.input); got != test.want {
			t.Fatalf("got %q want %q", got, test.want)
		}
	}
}
