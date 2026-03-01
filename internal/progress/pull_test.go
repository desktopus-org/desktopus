package progress

import (
	"strings"
	"testing"
)

// nonTTYWriter wraps a strings.Builder so isTerminal returns false.
type nonTTYWriter struct{ b strings.Builder }

func (w *nonTTYWriter) Write(p []byte) (int, error) { return w.b.Write(p) }
func (w *nonTTYWriter) String() string               { return w.b.String() }

// --- non-TTY mode ---

func TestRendererNonTTY_MeaningfulEventsOnly(t *testing.T) {
	w := &nonTTYWriter{}
	r := New(w)

	r.Update("abc123", "Pulling fs layer", "")
	r.Update("abc123", "Waiting", "")
	r.Update("abc123", "Downloading", "[=>   ] 1MB/50MB")
	r.Update("abc123", "Verifying Checksum", "")
	r.Update("abc123", "Download complete", "")
	r.Update("abc123", "Extracting", "")
	r.Update("abc123", "Pull complete", "")
	r.Print("Digest: sha256:deadbeef")
	r.Print("Status: Downloaded newer image for ubuntu:latest")

	got := w.String()

	if !strings.Contains(got, "abc123: Download complete") {
		t.Error("missing 'Download complete'")
	}
	if !strings.Contains(got, "abc123: Pull complete") {
		t.Error("missing 'Pull complete'")
	}
	if !strings.Contains(got, "Digest: sha256:deadbeef") {
		t.Error("missing digest line")
	}
	if !strings.Contains(got, "Status: Downloaded newer image") {
		t.Error("missing status line")
	}

	for _, noise := range []string{"Pulling fs layer", "Waiting", "[=>   ]", "Verifying Checksum", "Extracting"} {
		if strings.Contains(got, noise) {
			t.Errorf("noisy event %q should be suppressed", noise)
		}
	}
}

func TestRendererNonTTY_AlreadyExists(t *testing.T) {
	w := &nonTTYWriter{}
	r := New(w)

	r.Update("abc123", "Already exists", "")
	r.Print("Digest: sha256:deadbeef")
	r.Print("Status: Image is up to date for ubuntu:latest")

	got := w.String()

	if !strings.Contains(got, "abc123: Already exists") {
		t.Error("missing 'Already exists'")
	}
	if !strings.Contains(got, "Status: Image is up to date") {
		t.Error("missing up-to-date status")
	}
}

func TestRendererNonTTY_FlushBeforePullEventsDoesNotSuppressOutput(t *testing.T) {
	// streamBuildOutput calls Flush() on every stream event (e.g. "Step 1/9 : FROM...")
	// before any pull events arrive. Flush() must not silence subsequent Update() calls.
	w := &nonTTYWriter{}
	r := New(w)

	r.Flush() // simulates a stream event arriving before any pull events
	r.Update("abc123", "Pull complete", "")
	r.Print("Digest: sha256:deadbeef")

	got := w.String()

	if !strings.Contains(got, "abc123: Pull complete") {
		t.Error("Pull complete should appear even after an early Flush()")
	}
	if !strings.Contains(got, "Digest: sha256:deadbeef") {
		t.Error("Digest line should appear")
	}
}

// --- TTY mode ---

func TestRendererTTY_SpinnerOnActiveLayer(t *testing.T) {
	w := &nonTTYWriter{}
	r := newForTest(w, true)

	r.Update("abc123", "Waiting", "")

	got := w.String()

	// A spinner character from the braille set should appear.
	found := false
	for _, frame := range spinnerFrames {
		if strings.Contains(got, frame) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected a spinner character on active layer, got: %q", got)
	}
	if !strings.Contains(got, "abc123: Waiting") {
		t.Error("missing layer status")
	}
}

func TestRendererTTY_NospinnerOnDoneLayer(t *testing.T) {
	w := &nonTTYWriter{}
	r := newForTest(w, true)

	r.Update("abc123", "Pull complete", "")

	got := w.String()

	for _, frame := range spinnerFrames {
		if strings.Contains(got, frame) {
			t.Errorf("completed layer should not show spinner, got: %q", got)
		}
	}
	if !strings.Contains(got, "abc123: Pull complete") {
		t.Error("missing completed layer status")
	}
}

func TestRendererTTY_SpinnerOnDownloadingWithProgressBar(t *testing.T) {
	w := &nonTTYWriter{}
	r := newForTest(w, true)

	r.Update("abc123", "Downloading", "[====>   ] 45MB/89MB")

	got := w.String()

	// Spinner should appear alongside the progress bar.
	found := false
	for _, frame := range spinnerFrames {
		if strings.Contains(got, frame) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected spinner on downloading layer, got: %q", got)
	}
	if !strings.Contains(got, "[====>   ] 45MB/89MB") {
		t.Error("progress bar should still appear")
	}
}

func TestRendererTTY_PullingFromHeaderNoSpinner(t *testing.T) {
	w := &nonTTYWriter{}
	r := newForTest(w, true)

	r.Update("ubuntu-xfce", "Pulling from linuxserver/webtop", "")

	got := w.String()

	for _, frame := range spinnerFrames {
		if strings.Contains(got, frame) {
			t.Errorf("'Pulling from' header should not show spinner, got: %q", got)
		}
	}
	if !strings.Contains(got, "Pulling from linuxserver/webtop") {
		t.Error("missing 'Pulling from' header")
	}
}

func TestRendererTTY_FlushPrintsStaticOutput(t *testing.T) {
	w := &nonTTYWriter{}
	r := newForTest(w, true)

	r.Update("abc123", "Pull complete", "")
	r.Update("def456", "Waiting", "")

	// Clear ANSI noise for assertion purposes.
	r.Flush()

	got := w.String()
	// After Flush, both layers should appear as plain "id: status" lines.
	if !strings.Contains(got, "abc123: Pull complete") {
		t.Error("missing static Pull complete line after Flush")
	}
	if !strings.Contains(got, "def456: Waiting") {
		t.Error("missing static Waiting line after Flush")
	}
}

// --- isMeaningfulStatus ---

func TestIsMeaningfulStatus(t *testing.T) {
	meaningful := []string{
		"Pull complete",
		"Already exists",
		"Download complete",
		"Pulling from linuxserver/webtop",
		"Digest: sha256:abc",
		"Status: Downloaded newer image",
	}
	for _, s := range meaningful {
		if !isMeaningfulStatus(s) {
			t.Errorf("%q should be meaningful", s)
		}
	}

	noisy := []string{
		"Pulling fs layer",
		"Waiting",
		"Downloading",
		"Extracting",
		"Verifying Checksum",
	}
	for _, s := range noisy {
		if isMeaningfulStatus(s) {
			t.Errorf("%q should not be meaningful", s)
		}
	}
}
