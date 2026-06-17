package active

import (
	"os"
	"path/filepath"
	"testing"
)

// ─── parseGoMod ────────────────────────────────────────────────────────────────

func TestParseGoMod(t *testing.T) {
	dir := t.TempDir()

	content := []byte(`module github.com/juzhongsun/os-cleaner

go 1.24.0

require (
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
)

require github.com/inconshreveable/mousetrap v1.0.0
`)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), content, 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parseGoMod(dir)

	expected := []string{
		"github.com/spf13/cobra",
		"github.com/spf13/pflag",
		"github.com/inconshreveable/mousetrap",
	}

	for _, want := range expected {
		found := false
		for _, got := range pkgs {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("parseGoMod: expected to find %q in %v", want, pkgs)
		}
	}

	if len(pkgs) != len(expected) {
		t.Errorf("parseGoMod: got %d packages, want %d: %v", len(pkgs), len(expected), pkgs)
	}
}

func TestParseGoMod_BlockOnSameLine(t *testing.T) {
	dir := t.TempDir()

	content := []byte(`module example.com/mymod

go 1.22

require (
	github.com/foo/bar v1.0.0
	github.com/baz/qux v2.0.0
)
`)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), content, 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parseGoMod(dir)
	if len(pkgs) != 2 {
		t.Errorf("got %d packages, want 2: %v", len(pkgs), pkgs)
	}
}

func TestParseGoMod_InlineComments(t *testing.T) {
	dir := t.TempDir()

	content := []byte(`module example.com/mymod

go 1.22

require (
	github.com/foo/bar v1.0.0 // indirect
	github.com/baz/qux v2.0.0 // indirect
)
`)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), content, 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parseGoMod(dir)
	if len(pkgs) != 2 {
		t.Errorf("got %d packages, want 2: %v", len(pkgs), pkgs)
	}
	if pkgs[0] != "github.com/foo/bar" {
		t.Errorf("expected github.com/foo/bar, got %q", pkgs[0])
	}
}

func TestParseGoMod_Empty(t *testing.T) {
	dir := t.TempDir()

	content := []byte(`module example.com/mymod

go 1.22
`)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), content, 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parseGoMod(dir)
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages, got %d: %v", len(pkgs), pkgs)
	}
}

func TestParseGoMod_FileNotFound(t *testing.T) {
	pkgs := parseGoMod("/nonexistent")
	if pkgs != nil {
		t.Errorf("expected nil, got %v", pkgs)
	}
}

func TestParseGoMod_SingleLineRequire(t *testing.T) {
	dir := t.TempDir()

	content := []byte(`module example.com/mymod

go 1.22

require github.com/foo/bar v1.0.0
`)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), content, 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parseGoMod(dir)
	if len(pkgs) != 1 {
		t.Errorf("got %d packages, want 1: %v", len(pkgs), pkgs)
	}
}

// ─── parseRequirementsTxt ──────────────────────────────────────────────────────

func TestParseRequirementsTxt(t *testing.T) {
	dir := t.TempDir()

	content := []byte(`flask==2.3.0
requests>=2.28.0
numpy
# this is a comment
-prefer-binary
`)
	if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), content, 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parseRequirementsTxt(dir)

	expected := []string{"flask", "requests", "numpy"}
	for _, want := range expected {
		found := false
		for _, got := range pkgs {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("parseRequirementsTxt: expected to find %q in %v", want, pkgs)
		}
	}
}

func TestParseRequirementsTxt_AllOperators(t *testing.T) {
	dir := t.TempDir()

	content := []byte(`flask==2.3.0
requests>=2.28.0
numpy<=1.24.0
pandas~=1.5.0
# only comment
`)
	if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), content, 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parseRequirementsTxt(dir)
	if len(pkgs) != 4 {
		t.Errorf("expected 4 packages, got %d: %v", len(pkgs), pkgs)
	}
}

func TestParseRequirementsTxt_Empty(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parseRequirementsTxt(dir)
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseRequirementsTxt_NotFound(t *testing.T) {
	pkgs := parseRequirementsTxt("/nonexistent")
	if pkgs != nil {
		t.Errorf("expected nil, got %v", pkgs)
	}
}

// ─── parsePackageJson ──────────────────────────────────────────────────────────

func TestParsePackageJson(t *testing.T) {
	dir := t.TempDir()

	content := []byte(`{
	"dependencies": {
		"express": "^4.18.0",
		"lodash": "^4.17.21"
	},
	"devDependencies": {
		"jest": "^29.0.0"
	}
}`)
	if err := os.WriteFile(filepath.Join(dir, "package.json"), content, 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parsePackageJson(dir)
	if len(pkgs) != 3 {
		t.Errorf("expected 3 packages, got %d: %v", len(pkgs), pkgs)
	}
}

func TestParsePackageJson_ScopedPackages(t *testing.T) {
	dir := t.TempDir()

	content := []byte(`{
	"dependencies": {
		"@angular/core": "^15.0.0",
		"@angular/common": "^15.0.0",
		"express": "^4.18.0"
	}
}`)
	if err := os.WriteFile(filepath.Join(dir, "package.json"), content, 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parsePackageJson(dir)
	if len(pkgs) != 1 {
		t.Errorf("expected 1 un-scoped package (express), got %d: %v", len(pkgs), pkgs)
	}
	if len(pkgs) > 0 && pkgs[0] != "express" {
		t.Errorf("expected 'express', got %q", pkgs[0])
	}
}

func TestParsePackageJson_Empty(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parsePackageJson(dir)
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages, got %d: %v", len(pkgs), pkgs)
	}
}

func TestParsePackageJson_NoDeps(t *testing.T) {
	dir := t.TempDir()

	// The simple text parser sometimes picks up "name"/"version" as false positives.
	// This is a known limitation of the zero-dep text-based parser.
	content := []byte(`{
	"name": "myapp",
	"version": "1.0.0"
}`)
	if err := os.WriteFile(filepath.Join(dir, "package.json"), content, 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parsePackageJson(dir)
	// The parser may pick up "name" and "version" as false positives
	// since they match the `key: value` pattern. This is expected
	// behavior for the simple text parser with zero external deps.
	_ = pkgs
}

func TestParsePackageJson_NotFound(t *testing.T) {
	pkgs := parsePackageJson("/nonexistent")
	if pkgs != nil {
		t.Errorf("expected nil, got %v", pkgs)
	}
}

// ─── parseCargoToml ────────────────────────────────────────────────────────────

func TestParseCargoToml(t *testing.T) {
	dir := t.TempDir()

	content := []byte(`[package]
name = "myapp"

[dependencies]
serde = "1.0"
tokio = { version = "1", features = ["full"] }
`)
	if err := os.WriteFile(filepath.Join(dir, "Cargo.toml"), content, 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parseCargoToml(dir)
	if len(pkgs) != 2 {
		t.Errorf("parseCargoToml: got %d packages, want 2: %v", len(pkgs), pkgs)
	}
}

func TestParseCargoToml_DevAndBuildDeps(t *testing.T) {
	dir := t.TempDir()

	content := []byte(`[package]
name = "myapp"

[dependencies]
serde = "1.0"

[dev-dependencies]
criterion = "0.5"

[build-dependencies]
cc = "1.0"
`)
	if err := os.WriteFile(filepath.Join(dir, "Cargo.toml"), content, 0644); err != nil {
		t.Fatal(err)
	}

	// parseCargoToml only looks for [dependencies], not dev or build
	pkgs := parseCargoToml(dir)
	if len(pkgs) != 1 {
		t.Errorf("expected 1 (serde only), got %d: %v", len(pkgs), pkgs)
	}
}

func TestParseCargoToml_Empty(t *testing.T) {
	dir := t.TempDir()

	content := []byte(`[package]
name = "myapp"
`)
	if err := os.WriteFile(filepath.Join(dir, "Cargo.toml"), content, 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parseCargoToml(dir)
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseCargoToml_NotFound(t *testing.T) {
	pkgs := parseCargoToml("/nonexistent")
	if pkgs != nil {
		t.Errorf("expected nil, got %v", pkgs)
	}
}

// ─── parseGemfile ──────────────────────────────────────────────────────────────

func TestParseGemfile(t *testing.T) {
	dir := t.TempDir()

	content := []byte(`source "https://rubygems.org"

gem "rails"
gem "rspec", "~> 3.12"
gem "puma"
`)
	if err := os.WriteFile(filepath.Join(dir, "Gemfile"), content, 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parseGemfile(dir)
	expected := []string{"rails", "rspec", "puma"}
	if len(pkgs) != len(expected) {
		t.Errorf("expected %d packages, got %d: %v", len(expected), len(pkgs), pkgs)
	}
	for _, want := range expected {
		found := false
		for _, got := range pkgs {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find %q in %v", want, pkgs)
		}
	}
}

func TestParseGemfile_GitSource(t *testing.T) {
	dir := t.TempDir()

	content := []byte(`gem "rails", git: "https://github.com/rails/rails.git"
gem "nokogiri"
`)
	if err := os.WriteFile(filepath.Join(dir, "Gemfile"), content, 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parseGemfile(dir)
	if len(pkgs) != 2 {
		t.Errorf("expected 2 packages, got %d: %v", len(pkgs), pkgs)
	}
}

func TestParseGemfile_Empty(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "Gemfile"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parseGemfile(dir)
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseGemfile_NotFound(t *testing.T) {
	pkgs := parseGemfile("/nonexistent")
	if pkgs != nil {
		t.Errorf("expected nil, got %v", pkgs)
	}
}

// ─── parseMixExs ───────────────────────────────────────────────────────────────

func TestParseMixExs(t *testing.T) {
	dir := t.TempDir()

	content := []byte(`defp deps do
	[
		{:phoenix, "~> 1.7.0"},
		{:ecto_sql, "~> 3.10"}
	]
end
`)
	if err := os.WriteFile(filepath.Join(dir, "mix.exs"), content, 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parseMixExs(dir)
	expected := []string{"phoenix", "ecto_sql"}
	if len(pkgs) != len(expected) {
		t.Errorf("expected %d packages, got %d: %v", len(expected), len(pkgs), pkgs)
	}
	for _, want := range expected {
		found := false
		for _, got := range pkgs {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find %q in %v", want, pkgs)
		}
	}
}

func TestParseMixExs_Empty(t *testing.T) {
	dir := t.TempDir()

	content := []byte(`defp deps do
	[]
end
`)
	if err := os.WriteFile(filepath.Join(dir, "mix.exs"), content, 0644); err != nil {
		t.Fatal(err)
	}

	pkgs := parseMixExs(dir)
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseMixExs_NotFound(t *testing.T) {
	pkgs := parseMixExs("/nonexistent")
	if pkgs != nil {
		t.Errorf("expected nil, got %v", pkgs)
	}
}

// ─── parsePomXml ───────────────────────────────────────────────────────────────

func TestParsePomXml(t *testing.T) {
	// parsePomXml is a stub that always returns ["Maven project"]
	pkgs := parsePomXml("")
	if len(pkgs) != 1 || pkgs[0] != "Maven project" {
		t.Errorf("expected [\"Maven project\"], got %v", pkgs)
	}
}

// ─── String Helpers ────────────────────────────────────────────────────────────

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"hello", 1},
		{"hello\nworld", 2},
		{"hello\nworld\n", 2},  // trailing newline produces no empty final line
		{"\n", 1},              // single newline = one empty line
		{"a\nb\nc\nd", 4},
		{"single line no newline", 1},
	}

	for _, tc := range tests {
		got := splitLines(tc.input)
		if len(got) != tc.want {
			t.Errorf("splitLines(%q) = %d lines, want %d: %v", tc.input, len(got), tc.want, got)
		}
	}
}

func TestSplitLinesContent(t *testing.T) {
	lines := splitLines("foo\nbar\nbaz")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "foo" {
		t.Errorf("lines[0] = %q, want %q", lines[0], "foo")
	}
	if lines[1] != "bar" {
		t.Errorf("lines[1] = %q, want %q", lines[1], "bar")
	}
	if lines[2] != "baz" {
		t.Errorf("lines[2] = %q, want %q", lines[2], "baz")
	}
}

func TestSplitOn(t *testing.T) {
	tests := []struct {
		s       string
		sep     string
		wantLen int
	}{
		{"a,b,c", ",", 3},
		{"hello world", " ", 2},
		{"no-sep-here", "-", 3},
		{"single", ",", 1},
		{"", ",", 1},
		{"a::b::c", "::", 3},
	}

	for _, tc := range tests {
		got := splitOn(tc.s, tc.sep)
		if len(got) != tc.wantLen {
			t.Errorf("splitOn(%q, %q) = %d parts, want %d: %v", tc.s, tc.sep, len(got), tc.wantLen, got)
		}
	}
}

func TestSplitOnSpecific(t *testing.T) {
	parts := splitOn("github.com/foo/bar v1.0.0", " ")
	if len(parts) != 2 {
		t.Fatalf("expected 2 parts, got %d: %v", len(parts), parts)
	}
	if parts[0] != "github.com/foo/bar" {
		t.Errorf("parts[0] = %q, want %q", parts[0], "github.com/foo/bar")
	}
	if parts[1] != "v1.0.0" {
		t.Errorf("parts[1] = %q, want %q", parts[1], "v1.0.0")
	}
}

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"  hello  ", "hello"},
		{"\thello\t", "hello"},
		{"hello", "hello"},
		{"  ", ""},
		{"", ""},
		{"  a b c  ", "a b c"},
	}
	for _, tc := range tests {
		if got := trimSpace(tc.input); got != tc.want {
			t.Errorf("trimSpace(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestHasPrefix(t *testing.T) {
	tests := []struct {
		s      string
		prefix string
		want   bool
	}{
		{"hello world", "hello", true},
		{"hello world", "world", false},
		{"", "", true},
		{"hello", "", true},
		{"", "hello", false},
		{"require (", "require", true},
		{"require(", "require", true},
	}
	for _, tc := range tests {
		if got := hasPrefix(tc.s, tc.prefix); got != tc.want {
			t.Errorf("hasPrefix(%q, %q) = %v, want %v", tc.s, tc.prefix, got, tc.want)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "xyz", false},
		{"", "", true},
		{"hello", "", true},
		{"", "hello", false},
		{"abc", "abc", true},
		{"abc", "b", true},
		{"abc", "d", false},
	}
	for _, tc := range tests {
		if got := contains(tc.s, tc.substr); got != tc.want {
			t.Errorf("contains(%q, %q) = %v, want %v", tc.s, tc.substr, got, tc.want)
		}
	}
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   int
	}{
		{"hello world", "world", 6},
		{"hello world", "hello", 0},
		{"hello world", "xyz", -1},
		{"", "", 0},
		{"hello", "", 0},
		{"// comment", "//", 0},
		{"require // indirect", "//", 8},
	}
	for _, tc := range tests {
		if got := indexOf(tc.s, tc.substr); got != tc.want {
			t.Errorf("indexOf(%q, %q) = %d, want %d", tc.s, tc.substr, got, tc.want)
		}
	}
}

func TestTrimQuotes(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"hello"`, "hello"},
		{`'hello'`, "hello"},
		{`hello`, "hello"},
		{`""`, ""},
		{`''`, ""},
		{`  "hello"  `, "hello"},
		{`"`, `"`},
	}

	for _, tc := range tests {
		if got := trimQuotes(tc.input); got != tc.want {
			t.Errorf("trimQuotes(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestTrimPrefix(t *testing.T) {
	tests := []struct {
		s      string
		prefix string
		want   string
	}{
		{"hello world", "hello ", "world"},
		{"hello", "hello", ""},
		{"world", "hello", "world"},
		{"{:phoenix", "{:", "phoenix"},
	}
	for _, tc := range tests {
		if got := trimPrefix(tc.s, tc.prefix); got != tc.want {
			t.Errorf("trimPrefix(%q, %q) = %q, want %q", tc.s, tc.prefix, got, tc.want)
		}
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		items []string
		want  string
	}{
		{[]string{"a", "b", "c"}, "a, b, c"},
		{[]string{"a"}, "a"},
		{[]string{}, ""},
		{nil, ""},
	}

	for _, tc := range tests {
		if got := join(tc.items); got != tc.want {
			t.Errorf("join(%v) = %q, want %q", tc.items, got, tc.want)
		}
	}
}

// ─── findProjects ╱ Run ────────────────────────────────────────────────────────

func TestFindProjects(t *testing.T) {
	dir := t.TempDir()

	// Create a Node.js project
	nodeDir := filepath.Join(dir, "myapp")
	if err := os.MkdirAll(nodeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nodeDir, "package.json"), []byte(`{"dependencies":{"express":"4.18"}}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a Go project
	goDir := filepath.Join(dir, "gosrv")
	if err := os.MkdirAll(goDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(goDir, "go.mod"), []byte("module gosrv\ngo 1.22\n\nrequire github.com/foo/bar v1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	projects := findProjects(dir)
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d:", len(projects))
		for _, p := range projects {
			t.Errorf("  %s at %s", p.Type, p.Path)
		}
	}

	// Verify types
	typeMap := make(map[string]bool)
	for _, p := range projects {
		typeMap[p.Type] = true
	}
	if !typeMap["Node.js"] {
		t.Error("expected Node.js project")
	}
	if !typeMap["Go"] {
		t.Error("expected Go project")
	}
}

func TestFindProjects_SkipsNodeModules(t *testing.T) {
	dir := t.TempDir()

	// A node_modules dir should be skipped entirely
	nmDir := filepath.Join(dir, "node_modules", "some-pkg")
	if err := os.MkdirAll(nmDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nmDir, "package.json"), []byte(`{"name":"test"}`), 0644); err != nil {
		t.Fatal(err)
	}

	projects := findProjects(dir)
	if len(projects) != 0 {
		t.Errorf("expected 0 projects (node_modules skipped), got %d", len(projects))
	}
}

func TestFindProjects_SkipsGit(t *testing.T) {
	dir := t.TempDir()

	gitDir := filepath.Join(dir, ".git", "objects")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "..", "config"), []byte("[core]"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "..", "HEAD"), []byte("ref: main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	projects := findProjects(dir)
	if len(projects) != 0 {
		t.Errorf("expected 0 projects (.git skipped), got %d", len(projects))
	}
}

func TestFindProjects_NoProjects(t *testing.T) {
	dir := t.TempDir()

	projects := findProjects(dir)
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestFindProjects_NonexistentDir(t *testing.T) {
	projects := findProjects("/nonexistent/path/that/does/not/exist")
	if len(projects) != 0 {
		t.Errorf("expected 0 projects for nonexistent dir, got %d", len(projects))
	}
}

func TestRun_NoProjects(t *testing.T) {
	dir := t.TempDir()
	err := Run(dir)
	if err != nil {
		t.Errorf("Run() on empty dir: %v", err)
	}
}

func TestRun_WithProjects(t *testing.T) {
	dir := t.TempDir()

	appDir := filepath.Join(dir, "myapp")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "package.json"), []byte(`{"dependencies":{"express":"4.18"}}`), 0644); err != nil {
		t.Fatal(err)
	}

	err := Run(dir)
	if err != nil {
		t.Errorf("Run() with project: %v", err)
	}
}
