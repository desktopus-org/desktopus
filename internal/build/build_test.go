package build

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"text/template"

	"github.com/desktopus-org/desktopus/internal/config"
	"github.com/desktopus-org/desktopus/internal/module"
)

// --- BuildContext ---

func TestBuildContextAddFile(t *testing.T) {
	bctx := NewBuildContext()

	if err := bctx.AddFile("test.txt", []byte("hello world"), 0644); err != nil {
		t.Fatalf("AddFile: %v", err)
	}

	reader, err := bctx.Reader()
	if err != nil {
		t.Fatalf("Reader: %v", err)
	}

	tr := tar.NewReader(reader)
	header, err := tr.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}

	if header.Name != "test.txt" {
		t.Errorf("expected name 'test.txt', got %q", header.Name)
	}
	if header.Size != 11 {
		t.Errorf("expected size 11, got %d", header.Size)
	}
	if header.Mode != 0644 {
		t.Errorf("expected mode 0644, got %o", header.Mode)
	}

	content, _ := io.ReadAll(tr)
	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(content))
	}
}

func TestBuildContextMultipleFiles(t *testing.T) {
	bctx := NewBuildContext()

	files := map[string]string{
		"Dockerfile":   "FROM ubuntu",
		"playbook.yml": "- hosts: localhost",
		"ansible.cfg":  "[defaults]",
	}

	for name, content := range files {
		if err := bctx.AddFile(name, []byte(content), 0644); err != nil {
			t.Fatalf("AddFile %s: %v", name, err)
		}
	}

	reader, err := bctx.Reader()
	if err != nil {
		t.Fatalf("Reader: %v", err)
	}

	tr := tar.NewReader(reader)
	found := make(map[string]bool)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		found[header.Name] = true
	}

	for name := range files {
		if !found[name] {
			t.Errorf("file %q not found in tar", name)
		}
	}
}

func TestBuildContextExecutableMode(t *testing.T) {
	bctx := NewBuildContext()
	if err := bctx.AddFile("script.sh", []byte("#!/bin/bash"), 0755); err != nil {
		t.Fatalf("AddFile: %v", err)
	}

	reader, err := bctx.Reader()
	if err != nil {
		t.Fatalf("Reader: %v", err)
	}

	tr := tar.NewReader(reader)
	header, err := tr.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}

	if header.Mode != 0755 {
		t.Errorf("expected mode 0755, got %o", header.Mode)
	}
}

func TestBuildContextAddFileFromReader(t *testing.T) {
	bctx := NewBuildContext()

	content := "content from reader"
	r := strings.NewReader(content)

	if err := bctx.AddFileFromReader("from-reader.txt", r, int64(len(content)), 0644); err != nil {
		t.Fatalf("AddFileFromReader: %v", err)
	}

	reader, err := bctx.Reader()
	if err != nil {
		t.Fatalf("Reader: %v", err)
	}

	tr := tar.NewReader(reader)
	header, err := tr.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}

	if header.Name != "from-reader.txt" {
		t.Errorf("expected name 'from-reader.txt', got %q", header.Name)
	}

	got, _ := io.ReadAll(tr)
	if string(got) != content {
		t.Errorf("expected %q, got %q", content, string(got))
	}
}

// --- generateDockerfile ---

func TestGenerateDockerfile(t *testing.T) {
	tmplFuncs := templateFuncs()
	tmpl, err := template.New("Dockerfile.tmpl").Funcs(tmplFuncs).ParseFS(templatesFS, "templates/Dockerfile.tmpl")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	cfg := &config.DesktopConfig{
		Name:        "test-desktop",
		Description: "A test desktop",
		Base:        config.BaseSpec{OS: "ubuntu", Desktop: "xfce"},
	}

	modules := []*module.Module{
		{Name: "chrome", SystemPkgs: []string{"wget", "curl"}},
	}

	result, err := generateDockerfile(tmpl, cfg, modules, 0)
	if err != nil {
		t.Fatalf("generateDockerfile: %v", err)
	}

	dockerfile := string(result)

	if !strings.Contains(dockerfile, "FROM lscr.io/linuxserver/webtop:ubuntu-xfce") {
		t.Error("missing FROM instruction")
	}
	if !strings.Contains(dockerfile, "wget") {
		t.Error("missing system package 'wget'")
	}
	if !strings.Contains(dockerfile, "curl") {
		t.Error("missing system package 'curl'")
	}
	if !strings.Contains(dockerfile, "ansible-playbook") {
		t.Error("missing ansible-playbook command")
	}
	if !strings.Contains(dockerfile, `org.desktopus.name="test-desktop"`) {
		t.Error("missing name label")
	}
	// No post-run scripts, so should NOT have the section
	if strings.Contains(dockerfile, "custom-cont-init.d") {
		t.Error("should not have post-run section without post-run scripts")
	}
	// Ubuntu should use apt-get and python3-apt
	if !strings.Contains(dockerfile, "apt-get") {
		t.Error("ubuntu should use apt-get")
	}
	if !strings.Contains(dockerfile, "python3-apt") {
		t.Error("ubuntu should install python3-apt")
	}
}

func TestGenerateDockerfileFedora(t *testing.T) {
	tmplFuncs := templateFuncs()
	tmpl, err := template.New("Dockerfile.tmpl").Funcs(tmplFuncs).ParseFS(templatesFS, "templates/Dockerfile.tmpl")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	cfg := &config.DesktopConfig{
		Name: "fedora-desktop",
		Base: config.BaseSpec{OS: "fedora", Desktop: "xfce"},
	}

	result, err := generateDockerfile(tmpl, cfg, nil, 0)
	if err != nil {
		t.Fatalf("generateDockerfile: %v", err)
	}

	dockerfile := string(result)

	if !strings.Contains(dockerfile, "FROM lscr.io/linuxserver/webtop:fedora-xfce") {
		t.Error("missing FROM instruction for fedora")
	}
	if !strings.Contains(dockerfile, "dnf install") {
		t.Error("fedora should use dnf install")
	}
	if !strings.Contains(dockerfile, "python3-dnf") {
		t.Error("fedora should install python3-dnf")
	}
	if !strings.Contains(dockerfile, "dnf remove") {
		t.Error("fedora should use dnf remove for cleanup")
	}
	if !strings.Contains(dockerfile, "dnf clean all") {
		t.Error("fedora should run dnf clean all")
	}
}

func TestGenerateDockerfileArch(t *testing.T) {
	tmplFuncs := templateFuncs()
	tmpl, err := template.New("Dockerfile.tmpl").Funcs(tmplFuncs).ParseFS(templatesFS, "templates/Dockerfile.tmpl")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	cfg := &config.DesktopConfig{
		Name: "arch-desktop",
		Base: config.BaseSpec{OS: "arch", Desktop: "i3"},
	}

	result, err := generateDockerfile(tmpl, cfg, nil, 0)
	if err != nil {
		t.Fatalf("generateDockerfile: %v", err)
	}

	dockerfile := string(result)

	if !strings.Contains(dockerfile, "FROM lscr.io/linuxserver/webtop:arch-i3") {
		t.Error("missing FROM instruction for arch")
	}
	if !strings.Contains(dockerfile, "pacman -Sy --noconfirm") {
		t.Error("arch should use pacman -Sy --noconfirm")
	}
	if !strings.Contains(dockerfile, "pacman -Rns --noconfirm ansible") {
		t.Error("arch should use pacman -Rns for cleanup")
	}
	if !strings.Contains(dockerfile, "pacman -Scc --noconfirm") {
		t.Error("arch should run pacman -Scc")
	}
}

func TestGenerateDockerfileAlpine(t *testing.T) {
	tmplFuncs := templateFuncs()
	tmpl, err := template.New("Dockerfile.tmpl").Funcs(tmplFuncs).ParseFS(templatesFS, "templates/Dockerfile.tmpl")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	cfg := &config.DesktopConfig{
		Name: "alpine-desktop",
		Base: config.BaseSpec{OS: "alpine", Desktop: "xfce"},
	}

	result, err := generateDockerfile(tmpl, cfg, nil, 0)
	if err != nil {
		t.Fatalf("generateDockerfile: %v", err)
	}

	dockerfile := string(result)

	if !strings.Contains(dockerfile, "FROM lscr.io/linuxserver/webtop:latest") {
		t.Error("missing FROM instruction for alpine (alpine-xfce resolves to latest)")
	}
	if !strings.Contains(dockerfile, "apk add --no-cache") {
		t.Error("alpine should use apk add --no-cache")
	}
	if !strings.Contains(dockerfile, "python3") {
		t.Error("alpine should install python3")
	}
	if !strings.Contains(dockerfile, "apk del ansible") {
		t.Error("alpine should use apk del for cleanup")
	}
}

func TestGenerateDockerfileEL(t *testing.T) {
	tmplFuncs := templateFuncs()
	tmpl, err := template.New("Dockerfile.tmpl").Funcs(tmplFuncs).ParseFS(templatesFS, "templates/Dockerfile.tmpl")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	cfg := &config.DesktopConfig{
		Name: "el-desktop",
		Base: config.BaseSpec{OS: "el", Desktop: "xfce"},
	}

	result, err := generateDockerfile(tmpl, cfg, nil, 0)
	if err != nil {
		t.Fatalf("generateDockerfile: %v", err)
	}

	dockerfile := string(result)

	if !strings.Contains(dockerfile, "FROM lscr.io/linuxserver/webtop:el-xfce") {
		t.Error("missing FROM instruction for el")
	}
	if !strings.Contains(dockerfile, "dnf install") {
		t.Error("el should use dnf install")
	}
	if !strings.Contains(dockerfile, "python3-dnf") {
		t.Error("el should install python3-dnf")
	}
}

func TestGenerateDockerfileWithPostRun(t *testing.T) {
	tmplFuncs := templateFuncs()
	tmpl, err := template.New("Dockerfile.tmpl").Funcs(tmplFuncs).ParseFS(templatesFS, "templates/Dockerfile.tmpl")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	cfg := &config.DesktopConfig{
		Name: "test",
		Base: config.BaseSpec{OS: "ubuntu", Desktop: "xfce"},
		PostRun: []config.PostRunScript{
			{Name: "setup", Script: "echo hi"},
		},
	}

	result, err := generateDockerfile(tmpl, cfg, nil, 0)
	if err != nil {
		t.Fatalf("generateDockerfile: %v", err)
	}

	if !strings.Contains(string(result), "custom-cont-init.d") {
		t.Error("should include post-run section")
	}
}

func TestGenerateDockerfileWithVerbosity(t *testing.T) {
	tmplFuncs := templateFuncs()
	tmpl, err := template.New("Dockerfile.tmpl").Funcs(tmplFuncs).ParseFS(templatesFS, "templates/Dockerfile.tmpl")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	cfg := &config.DesktopConfig{
		Name: "test",
		Base: config.BaseSpec{OS: "ubuntu", Desktop: "xfce"},
	}

	result, err := generateDockerfile(tmpl, cfg, nil, 3)
	if err != nil {
		t.Fatalf("generateDockerfile: %v", err)
	}

	if !strings.Contains(string(result), "-vvv") {
		t.Error("expected -vvv for verbosity 3")
	}
}

func TestGenerateDockerfileDeduplicatesPackages(t *testing.T) {
	tmplFuncs := templateFuncs()
	tmpl, err := template.New("Dockerfile.tmpl").Funcs(tmplFuncs).ParseFS(templatesFS, "templates/Dockerfile.tmpl")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	cfg := &config.DesktopConfig{
		Name: "test",
		Base: config.BaseSpec{OS: "ubuntu", Desktop: "xfce"},
	}

	modules := []*module.Module{
		{Name: "chrome", SystemPkgs: []string{"wget", "curl"}},
		{Name: "firefox", SystemPkgs: []string{"curl", "git"}},
	}

	result, err := generateDockerfile(tmpl, cfg, modules, 0)
	if err != nil {
		t.Fatalf("generateDockerfile: %v", err)
	}

	// curl should appear only once
	if strings.Count(string(result), "curl") != 1 {
		t.Error("expected 'curl' to appear exactly once (deduplication)")
	}
}

// --- generatePlaybook ---

func TestGeneratePlaybook(t *testing.T) {
	tmpl, err := template.New("playbook.yml.tmpl").ParseFS(templatesFS, "templates/playbook.yml.tmpl")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	modules := []*module.Module{
		{
			Name:    "chrome",
			Path:    "chrome",
			Builtin: true,
			Vars: map[string]module.Var{
				"chrome_channel": {Default: "stable"},
			},
		},
	}

	overrides := []map[string]interface{}{
		{"chrome_channel": "beta"},
	}

	result, err := generatePlaybook(tmpl, modules, overrides, "ubuntu", "desktopus", "/home/desktopus")
	if err != nil {
		t.Fatalf("generatePlaybook: %v", err)
	}

	playbook := string(result)

	if !strings.Contains(playbook, "localhost") {
		t.Error("missing hosts: localhost")
	}
	if !strings.Contains(playbook, "chrome") {
		t.Error("missing chrome module reference")
	}
	if !strings.Contains(playbook, "beta") {
		t.Error("expected override value 'beta'")
	}
}

func TestGeneratePlaybookDefaults(t *testing.T) {
	tmpl, err := template.New("playbook.yml.tmpl").ParseFS(templatesFS, "templates/playbook.yml.tmpl")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	modules := []*module.Module{
		{
			Name:    "chrome",
			Path:    "chrome",
			Builtin: true,
			Vars: map[string]module.Var{
				"chrome_channel": {Default: "stable"},
			},
		},
	}

	// No overrides
	result, err := generatePlaybook(tmpl, modules, nil, "ubuntu", "desktopus", "/home/desktopus")
	if err != nil {
		t.Fatalf("generatePlaybook: %v", err)
	}

	if !strings.Contains(string(result), "stable") {
		t.Error("expected default value 'stable'")
	}
}

func TestGeneratePlaybookOSSpecificTaskFile(t *testing.T) {
	tmpl, err := template.New("playbook.yml.tmpl").ParseFS(templatesFS, "templates/playbook.yml.tmpl")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	modules := []*module.Module{
		{
			Name:        "chrome",
			Path:        "chrome",
			Builtin:     true,
			OSTaskFiles: map[string]bool{"ubuntu": true},
		},
		{
			Name:    "vscode",
			Path:    "vscode",
			Builtin: true,
		},
	}

	result, err := generatePlaybook(tmpl, modules, nil, "ubuntu", "desktopus", "/home/desktopus")
	if err != nil {
		t.Fatalf("generatePlaybook: %v", err)
	}

	playbook := string(result)

	// chrome has ubuntu-specific tasks
	if !strings.Contains(playbook, "modules/chrome/tasks/ubuntu.yml") {
		t.Error("expected chrome to use tasks/ubuntu.yml for ubuntu build")
	}
	// vscode has no OS-specific tasks, should fall back
	if !strings.Contains(playbook, "modules/vscode/tasks/main.yml") {
		t.Error("expected vscode to fall back to tasks/main.yml")
	}
}

func TestGeneratePlaybookOSFallback(t *testing.T) {
	tmpl, err := template.New("playbook.yml.tmpl").ParseFS(templatesFS, "templates/playbook.yml.tmpl")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	modules := []*module.Module{
		{
			Name:        "chrome",
			Path:        "chrome",
			Builtin:     true,
			OSTaskFiles: map[string]bool{"ubuntu": true},
		},
	}

	// Building for fedora, but chrome only has ubuntu-specific tasks
	result, err := generatePlaybook(tmpl, modules, nil, "fedora", "desktopus", "/home/desktopus")
	if err != nil {
		t.Fatalf("generatePlaybook: %v", err)
	}

	playbook := string(result)

	if !strings.Contains(playbook, "modules/chrome/tasks/main.yml") {
		t.Error("expected chrome to fall back to tasks/main.yml for fedora build")
	}
}

// --- addPostRunScripts ---

func TestAddPostRunScripts(t *testing.T) {
	bctx := NewBuildContext()
	cfg := &config.DesktopConfig{
		PostRun: []config.PostRunScript{
			{Name: "setup-git", RunAs: "abc", Script: "git config --global user.name test"},
			{Name: "root-setup", RunAs: "root", Script: "apt-get update"},
		},
	}

	if err := addPostRunScripts(bctx, cfg); err != nil {
		t.Fatalf("addPostRunScripts: %v", err)
	}

	reader, err := bctx.Reader()
	if err != nil {
		t.Fatalf("Reader: %v", err)
	}

	tr := tar.NewReader(reader)
	files := make(map[string]string)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		data, _ := io.ReadAll(tr)
		files[header.Name] = string(data)
	}

	// Check first script (abc user, with s6-setuidgid)
	script1, ok := files["postrun/50-desktopus-setup-git.sh"]
	if !ok {
		t.Fatal("missing postrun/50-desktopus-setup-git.sh")
	}
	if !strings.Contains(script1, "s6-setuidgid abc") {
		t.Error("abc script should use s6-setuidgid")
	}
	if !strings.Contains(script1, "git config") {
		t.Error("missing script content")
	}

	// Check second script (root, no s6-setuidgid)
	script2, ok := files["postrun/51-desktopus-root-setup.sh"]
	if !ok {
		t.Fatal("missing postrun/51-desktopus-root-setup.sh")
	}
	if strings.Contains(script2, "s6-setuidgid") {
		t.Error("root script should NOT use s6-setuidgid")
	}
	if !strings.Contains(script2, "apt-get update") {
		t.Error("missing script content")
	}
}

func TestAddPostRunScriptsDefaultRunAs(t *testing.T) {
	bctx := NewBuildContext()
	cfg := &config.DesktopConfig{
		PostRun: []config.PostRunScript{
			{Name: "test", Script: "echo hi"}, // no RunAs specified
		},
	}

	if err := addPostRunScripts(bctx, cfg); err != nil {
		t.Fatalf("addPostRunScripts: %v", err)
	}

	reader, err := bctx.Reader()
	if err != nil {
		t.Fatalf("Reader: %v", err)
	}

	tr := tar.NewReader(reader)
	header, err := tr.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}

	data, _ := io.ReadAll(tr)
	// Default should be "desktopus" (effective user when no user is set)
	if !strings.Contains(string(data), "s6-setuidgid desktopus") {
		t.Errorf("default runas should be desktopus, script: %s", string(data))
	}
	if header.Mode != 0755 {
		t.Errorf("expected mode 0755, got %o", header.Mode)
	}
}

// --- addRuntimeFiles ---

func TestAddRuntimeFiles(t *testing.T) {
	bctx := NewBuildContext()
	cfg := &config.DesktopConfig{
		Files: []config.FileSpec{
			{Path: "/config/.bashrc", Content: "export PS1='\\u@\\h'", Mode: "0644"},
			{Path: "/config/.vimrc", Content: "set number"},
		},
	}

	if err := addRuntimeFiles(bctx, cfg); err != nil {
		t.Fatalf("addRuntimeFiles: %v", err)
	}

	reader, err := bctx.Reader()
	if err != nil {
		t.Fatalf("Reader: %v", err)
	}

	tr := tar.NewReader(reader)
	files := make(map[string]string)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		data, _ := io.ReadAll(tr)
		files[header.Name] = string(data)
	}

	// Check runtime files were added
	if _, ok := files["runtime-files/config/.bashrc"]; !ok {
		t.Error("missing runtime-files/config/.bashrc")
	}
	if _, ok := files["runtime-files/config/.vimrc"]; !ok {
		t.Error("missing runtime-files/config/.vimrc")
	}

	// Check provisioner script
	script, ok := files["99-desktopus-files.sh"]
	if !ok {
		t.Fatal("missing 99-desktopus-files.sh")
	}
	if !strings.Contains(script, "envsubst") {
		t.Error("provisioner should use envsubst")
	}
	if !strings.Contains(script, "/config/.bashrc") {
		t.Error("provisioner should reference .bashrc path")
	}
}

func TestAddRuntimeFilesEmpty(t *testing.T) {
	bctx := NewBuildContext()
	cfg := &config.DesktopConfig{} // no files

	if err := addRuntimeFiles(bctx, cfg); err != nil {
		t.Fatalf("addRuntimeFiles: %v", err)
	}

	reader, err := bctx.Reader()
	if err != nil {
		t.Fatalf("Reader: %v", err)
	}

	tr := tar.NewReader(reader)
	_, err = tr.Next()
	if err != io.EOF {
		t.Error("expected empty tar for no runtime files")
	}
}

// --- streamBuildOutput ---

func TestStreamBuildOutput(t *testing.T) {
	messages := []struct {
		Stream string `json:"stream,omitempty"`
		Error  string `json:"error,omitempty"`
	}{
		{Stream: "Step 1/5 : FROM ubuntu\n"},
		{Stream: " ---> abc123\n"},
		{Stream: "Successfully built abc123\n"},
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, m := range messages {
		if err := enc.Encode(m); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}

	var output bytes.Buffer
	err := streamBuildOutput(&buf, &output)
	if err != nil {
		t.Fatalf("streamBuildOutput: %v", err)
	}

	if !strings.Contains(output.String(), "Step 1/5") {
		t.Error("missing stream output")
	}
	if !strings.Contains(output.String(), "Successfully built") {
		t.Error("missing final stream output")
	}
}

func TestStreamBuildOutputError(t *testing.T) {
	messages := []struct {
		Stream string `json:"stream,omitempty"`
		Error  string `json:"error,omitempty"`
	}{
		{Stream: "Step 1/5 : FROM ubuntu\n"},
		{Error: "something went wrong"},
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, m := range messages {
		if err := enc.Encode(m); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}

	var output bytes.Buffer
	err := streamBuildOutput(&buf, &output)
	if err == nil {
		t.Error("expected error from build stream")
	}
	if !strings.Contains(err.Error(), "something went wrong") {
		t.Errorf("expected error message, got: %v", err)
	}
}

func TestStreamBuildOutputEmpty(t *testing.T) {
	var buf bytes.Buffer
	var output bytes.Buffer
	err := streamBuildOutput(&buf, &output)
	if err != nil {
		t.Fatalf("empty stream should not error: %v", err)
	}
}

func TestStreamBuildOutputPullProgress(t *testing.T) {
	// Docker emits pull progress events (status/id/progress) during FROM pulls.
	// Only meaningful state transitions should be shown; intermediate noise
	// (Waiting, Downloading, Extracting, Verifying Checksum) must be suppressed.
	input := `{"status":"Pulling from linuxserver/webtop","id":"ubuntu-xfce"}
{"status":"Pulling fs layer","progressDetail":{},"id":"abc12345"}
{"status":"Waiting","progressDetail":{},"id":"abc12345"}
{"status":"Downloading","progressDetail":{"current":100,"total":1000},"progress":"[=>  ]","id":"abc12345"}
{"status":"Verifying Checksum","progressDetail":{},"id":"abc12345"}
{"status":"Download complete","progressDetail":{},"id":"abc12345"}
{"status":"Extracting","progressDetail":{},"id":"abc12345"}
{"status":"Pull complete","progressDetail":{},"id":"abc12345"}
{"status":"Digest: sha256:deadbeef"}
{"status":"Status: Downloaded newer image for linuxserver/webtop:ubuntu-xfce"}
{"stream":"Step 1/9 : FROM lscr.io/linuxserver/webtop:ubuntu-xfce\n"}
`
	var output bytes.Buffer
	if err := streamBuildOutput(strings.NewReader(input), &output); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := output.String()

	// Meaningful events must appear
	if !strings.Contains(got, "ubuntu-xfce: Pulling from linuxserver/webtop") {
		t.Error("missing pull-from line")
	}
	if !strings.Contains(got, "abc12345: Pull complete") {
		t.Error("missing pull-complete line")
	}
	if !strings.Contains(got, "Digest: sha256:deadbeef") {
		t.Error("missing digest line")
	}
	if !strings.Contains(got, "Status: Downloaded newer image") {
		t.Error("missing status line")
	}

	// Noisy intermediate events must be suppressed
	if strings.Contains(got, "Pulling fs layer") {
		t.Error("'Pulling fs layer' should be suppressed")
	}
	if strings.Contains(got, "Waiting") {
		t.Error("'Waiting' should be suppressed")
	}
	if strings.Contains(got, "[=>  ]") {
		t.Error("download progress bar should be suppressed")
	}
	if strings.Contains(got, "Extracting") {
		t.Error("'Extracting' should be suppressed")
	}
	if strings.Contains(got, "Verifying Checksum") {
		t.Error("'Verifying Checksum' should be suppressed")
	}

	// Regular stream events must still appear
	if !strings.Contains(got, "Step 1/9") {
		t.Error("missing stream output")
	}
}

// --- generateDockerfile user creation ---

func TestGenerateDockerfileDefaultUser(t *testing.T) {
	tmplFuncs := templateFuncs()
	tmpl, err := template.New("Dockerfile.tmpl").Funcs(tmplFuncs).ParseFS(templatesFS, "templates/Dockerfile.tmpl")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	cfg := &config.DesktopConfig{
		Name: "test",
		Base: config.BaseSpec{OS: "ubuntu", Desktop: "xfce"},
		// User omitted → creates "desktopus" with home "/home/desktopus"
	}

	result, err := generateDockerfile(tmpl, cfg, nil, 0)
	if err != nil {
		t.Fatalf("generateDockerfile: %v", err)
	}

	dockerfile := string(result)
	if !strings.Contains(dockerfile, "RUN useradd -m -d /home/desktopus -s /bin/bash desktopus") {
		t.Errorf("expected useradd for default user, got:\n%s", dockerfile)
	}
}

func TestGenerateDockerfileAbcUser(t *testing.T) {
	tmplFuncs := templateFuncs()
	tmpl, err := template.New("Dockerfile.tmpl").Funcs(tmplFuncs).ParseFS(templatesFS, "templates/Dockerfile.tmpl")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	cfg := &config.DesktopConfig{
		Name: "test",
		Base: config.BaseSpec{OS: "ubuntu", Desktop: "xfce"},
		User: "abc",
	}

	result, err := generateDockerfile(tmpl, cfg, nil, 0)
	if err != nil {
		t.Fatalf("generateDockerfile: %v", err)
	}

	if strings.Contains(string(result), "RUN useradd") {
		t.Error("abc user should not emit RUN useradd (built-in user)")
	}
}

func TestGenerateDockerfileCustomUser(t *testing.T) {
	tmplFuncs := templateFuncs()
	tmpl, err := template.New("Dockerfile.tmpl").Funcs(tmplFuncs).ParseFS(templatesFS, "templates/Dockerfile.tmpl")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	cfg := &config.DesktopConfig{
		Name: "test",
		Base: config.BaseSpec{OS: "ubuntu", Desktop: "xfce"},
		User: "carlos",
	}

	result, err := generateDockerfile(tmpl, cfg, nil, 0)
	if err != nil {
		t.Fatalf("generateDockerfile: %v", err)
	}

	dockerfile := string(result)
	if !strings.Contains(dockerfile, "RUN useradd -m -d /home/carlos -s /bin/bash carlos") {
		t.Errorf("expected useradd for custom user, got:\n%s", dockerfile)
	}
}

// --- generatePlaybook user/home ---

func TestGeneratePlaybookDefaultUser(t *testing.T) {
	tmpl, err := template.New("playbook.yml.tmpl").ParseFS(templatesFS, "templates/playbook.yml.tmpl")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	result, err := generatePlaybook(tmpl, nil, nil, "ubuntu", "desktopus", "/home/desktopus")
	if err != nil {
		t.Fatalf("generatePlaybook: %v", err)
	}

	playbook := string(result)
	if !strings.Contains(playbook, "desktopus_user: desktopus") {
		t.Errorf("expected desktopus_user: desktopus, got:\n%s", playbook)
	}
	if !strings.Contains(playbook, "desktopus_home: /home/desktopus") {
		t.Errorf("expected desktopus_home: /home/desktopus, got:\n%s", playbook)
	}
}

func TestGeneratePlaybookCustomUser(t *testing.T) {
	tmpl, err := template.New("playbook.yml.tmpl").ParseFS(templatesFS, "templates/playbook.yml.tmpl")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	result, err := generatePlaybook(tmpl, nil, nil, "ubuntu", "carlos", "/home/carlos")
	if err != nil {
		t.Fatalf("generatePlaybook: %v", err)
	}

	playbook := string(result)
	if !strings.Contains(playbook, "desktopus_user: carlos") {
		t.Errorf("expected desktopus_user: carlos, got:\n%s", playbook)
	}
	if !strings.Contains(playbook, "desktopus_home: /home/carlos") {
		t.Errorf("expected desktopus_home: /home/carlos, got:\n%s", playbook)
	}
}

// --- addPostRunScripts default user ---

func TestAddPostRunScriptsDefaultUser(t *testing.T) {
	bctx := NewBuildContext()
	cfg := &config.DesktopConfig{
		// User omitted → effective user is "desktopus"
		PostRun: []config.PostRunScript{
			{Name: "test", Script: "echo hi"},
		},
	}

	if err := addPostRunScripts(bctx, cfg); err != nil {
		t.Fatalf("addPostRunScripts: %v", err)
	}

	reader, err := bctx.Reader()
	if err != nil {
		t.Fatalf("Reader: %v", err)
	}

	tr := tar.NewReader(reader)
	_, err = tr.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}

	data, _ := io.ReadAll(tr)
	if !strings.Contains(string(data), "s6-setuidgid desktopus") {
		t.Errorf("default runas should be desktopus, got:\n%s", string(data))
	}
}

// --- addRuntimeFiles default user ---

func TestAddRuntimeFilesDefaultUser(t *testing.T) {
	bctx := NewBuildContext()
	cfg := &config.DesktopConfig{
		// User omitted → effective user is "desktopus"
		Files: []config.FileSpec{
			{Path: "/home/desktopus/.bashrc", Content: "# bashrc"},
		},
	}

	if err := addRuntimeFiles(bctx, cfg); err != nil {
		t.Fatalf("addRuntimeFiles: %v", err)
	}

	reader, err := bctx.Reader()
	if err != nil {
		t.Fatalf("Reader: %v", err)
	}

	tr := tar.NewReader(reader)
	files := make(map[string]string)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		data, _ := io.ReadAll(tr)
		files[header.Name] = string(data)
	}

	script, ok := files["99-desktopus-files.sh"]
	if !ok {
		t.Fatal("missing 99-desktopus-files.sh")
	}
	if !strings.Contains(script, "chown desktopus:desktopus") {
		t.Errorf("provisioner should use chown desktopus:desktopus, got:\n%s", script)
	}
}

// --- templateFuncs ---

func TestTemplateFuncRepeat(t *testing.T) {
	funcs := templateFuncs()
	repeatFn := funcs["repeat"].(func(int, string) string)

	if got := repeatFn(3, "v"); got != "vvv" {
		t.Errorf("repeat(3, v) = %q, want 'vvv'", got)
	}
	if got := repeatFn(0, "v"); got != "" {
		t.Errorf("repeat(0, v) = %q, want ''", got)
	}
	if got := repeatFn(1, "abc"); got != "abc" {
		t.Errorf("repeat(1, abc) = %q, want 'abc'", got)
	}
}
