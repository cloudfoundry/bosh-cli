package index

import "testing"

func TestIndexFileComponentsBeforeFilesystem(t *testing.T) {
	for _, name := range []string{"../outside", "..", `..\outside`, "/outside"} {
		index := FSIndex{}
		if _, _, err := index.Find(name, "fingerprint"); err == nil {
			t.Fatalf("accepted index lookup %q", name)
		}
		if _, _, err := index.Add(name, "fingerprint", "archive", "sha1"); err == nil {
			t.Fatalf("accepted index write %q", name)
		}
	}
}
