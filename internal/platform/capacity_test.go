package platform_test

import (
	"path/filepath"
	"testing"

	"github.com/jellydn/macsweep/internal/platform"
)

func TestVolumeCapacityAndUnavailablePath(t *testing.T) {
	dir := t.TempDir()
	volume, err := platform.Volume(dir)
	if err != nil {
		t.Fatal(err)
	}
	if volume.Total == 0 || volume.Available > volume.Total {
		t.Fatalf("invalid capacity: %+v", volume)
	}
	if _, err := platform.Volume(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("missing volume reported as available")
	}
}
