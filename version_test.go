package version

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// setBuildVars overrides the ldflags-settable variables for the duration of a
// test and restores the original values afterwards. Tests using it must not
// run in parallel, since they mutate package-level state.
func setBuildVars(t *testing.T, version, gitCommit, buildDate string) {
	t.Helper()
	origVersion, origCommit, origDate := Version, GitCommit, BuildDate
	t.Cleanup(func() {
		Version, GitCommit, BuildDate = origVersion, origCommit, origDate
	})
	Version, GitCommit, BuildDate = version, gitCommit, buildDate
}

// TestDefaults verifies the values used when no -ldflags are provided
// (e.g. `go run` / `go test`), and that runtime info is detected correctly.
func TestDefaults(t *testing.T) {
	got := get()
	want := info{
		Version:   "development",
		GitCommit: "none",
		BuildDate: "unknown",
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
	if got != want {
		t.Errorf("get() = %+v; want %+v", got, want)
	}
}

// TestGetReflectsBuildVars verifies get() reads the package variables at call
// time, so values injected at build time (or changed later) are reported.
func TestGetReflectsBuildVars(t *testing.T) {
	setBuildVars(t, "v1.2.3", "abc123", "2026-03-19T10:00:00Z")

	got := get()
	if got.Version != "v1.2.3" || got.GitCommit != "abc123" || got.BuildDate != "2026-03-19T10:00:00Z" {
		t.Errorf("get() = %+v; want build vars v1.2.3/abc123/2026-03-19T10:00:00Z", got)
	}
}

// TestInfoString verifies the exact multi-line format: field order, label
// padding, and no trailing newline.
func TestInfoString(t *testing.T) {
	tests := []struct {
		name string
		in   info
		want string
	}{
		{
			name: "all fields set",
			in: info{
				Version:   "v1.2.3",
				GitCommit: "abc123",
				BuildDate: "2026-03-19T10:00:00Z",
				GoVersion: "go1.26",
				Platform:  "linux/amd64",
			},
			want: "Version   : v1.2.3\n" +
				"GitCommit : abc123\n" +
				"BuildDate : 2026-03-19T10:00:00Z\n" +
				"GoVersion : go1.26\n" +
				"Platform  : linux/amd64",
		},
		{
			name: "empty fields",
			in:   info{},
			want: "Version   : \n" +
				"GitCommit : \n" +
				"BuildDate : \n" +
				"GoVersion : \n" +
				"Platform  : ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.string(); got != tt.want {
				t.Errorf("info.string() =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

// TestInfoStringCoversAllFields guards against a field being added to info
// without appearing in the output (one line per field).
func TestInfoStringCoversAllFields(t *testing.T) {
	const numFields = 5 // keep in sync with the info struct
	lines := strings.Split(info{}.string(), "\n")
	if len(lines) != numFields {
		t.Errorf("info.string() produced %d lines; want %d (one per field)", len(lines), numFields)
	}
}

// TestInfoShort verifies short() contains only the version, aligned with the
// value column of string().
func TestInfoShort(t *testing.T) {
	in := info{Version: "v1.2.3", GitCommit: "abc123", Platform: "linux/amd64"}

	got := in.short()
	if want := "Version:    v1.2.3"; got != want {
		t.Errorf("info.short() = %q; want %q", got, want)
	}

	// The value must start at the same column as in string(), so both
	// outputs line up when printed one after the other.
	firstLine := strings.SplitN(in.string(), "\n", 2)[0]
	if strings.Index(got, "v1.2.3") != strings.Index(firstLine, "v1.2.3") {
		t.Errorf("short() value column misaligned with string(): %q vs %q", got, firstLine)
	}
}

// TestHelpers verifies the exported helpers reflect the current build vars.
func TestHelpers(t *testing.T) {
	setBuildVars(t, "v9.9.9", "deadbeef", "2026-01-01T00:00:00Z")

	wantString := fmt.Sprintf("Version   : v9.9.9\n"+
		"GitCommit : deadbeef\n"+
		"BuildDate : 2026-01-01T00:00:00Z\n"+
		"GoVersion : %s\n"+
		"Platform  : %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	if got := String(); got != wantString {
		t.Errorf("String() =\n%s\nwant\n%s", got, wantString)
	}

	if got, want := Short(), "Version:    v9.9.9"; got != want {
		t.Errorf("Short() = %q; want %q", got, want)
	}
}

// TestLdflagsInjection builds a real binary with -ldflags -X and checks the
// values are picked up. This guards the main contract of the package: the
// variables must stay settable at link time (e.g. not turned into consts or
// renamed, which would make -X silently do nothing).
func TestLdflagsInjection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping build-based test in -short mode")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain not found in PATH")
	}

	const pkg = "github.com/ghdrope/go-version"
	bin := filepath.Join(t.TempDir(), "version-debug")
	ldflags := strings.Join([]string{
		"-X " + pkg + ".Version=v1.2.3",
		"-X " + pkg + ".GitCommit=abc123",
		"-X " + pkg + ".BuildDate=2026-03-19T10:00:00Z",
	}, " ")

	build := exec.Command(goBin, "build", "-ldflags", ldflags, "-o", bin, "./hack")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}

	out, err := exec.Command(bin).CombinedOutput()
	if err != nil {
		t.Fatalf("running binary failed: %v\n%s", err, out)
	}
	for _, want := range []string{
		"Version   : v1.2.3",
		"GitCommit : abc123",
		"BuildDate : 2026-03-19T10:00:00Z",
		"Version:    v1.2.3",
	} {
		if !strings.Contains(string(out), want) {
			t.Errorf("binary output missing %q\noutput:\n%s", want, out)
		}
	}
}

func ExampleShort() {
	fmt.Println(Short())
	// Output: Version:    development
}
