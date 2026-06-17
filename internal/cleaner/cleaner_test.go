package cleaner

import (
	"testing"
)

func TestCleanOptionsDefault(t *testing.T) {
	var opts CleanOptions

	if opts.DryRun {
		t.Error("default DryRun should be false")
	}
	if opts.SafeMode {
		t.Error("default SafeMode should be false")
	}
}

func TestCleanOptionsSafeMode(t *testing.T) {
	opts := CleanOptions{
		SafeMode: true,
	}

	if !opts.SafeMode {
		t.Error("SafeMode should be true")
	}
}

func TestCleanOptionsDryRun(t *testing.T) {
	opts := CleanOptions{
		DryRun: true,
	}

	if !opts.DryRun {
		t.Error("DryRun should be true")
	}
}

func TestCleanOptionsRecoverable(t *testing.T) {
	opts := CleanOptions{
		Recoverable: true,
	}

	if !opts.Recoverable {
		t.Error("Recoverable should be true")
	}
}
