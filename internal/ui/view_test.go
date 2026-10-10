package ui_test

import (
	"testing"

	"github.com/egoist/mygo/ui"
	appui "github.com/jellydn/macsweep/internal/ui"
)

func TestSetupViewDoesNotOfferActiveCleanup(t *testing.T) {
	tt := ui.NewTester(appui.View, 640, 400)
	for _, label := range []string{"Scan Mac", "Move to Trash"} {
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
		if tt.Focused(label) {
			t.Fatalf("setup control %q is active", label)
		}
	}
	if !tt.HasText("Scanning and cleanup are not available yet") {
		t.Fatal("setup build does not explain its limits")
	}
	if urls := tt.OpenedURLs(); len(urls) != 0 {
		t.Fatalf("setup interaction opened external URLs: %v", urls)
	}
}
