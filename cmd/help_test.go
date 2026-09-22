package cmd

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// captureHelp redirects rootCmd's stdout and stderr while fn runs and
// returns whatever was written to each stream.
func captureHelp(t *testing.T, fn func()) (stdout, stderr string) {
	t.Helper()
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	rootCmd.SetOut(out)
	rootCmd.SetErr(errOut)
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	}()
	fn()
	return out.String(), errOut.String()
}

// helpRowLine is one parsed row of a help section: its left cell (command
// name or flag), its right cell (description) and the column at which the
// description starts (0-based, in runes).
type helpRowLine struct {
	left     string
	right    string
	rightCol int
}

// helpSection returns the parsed rows of a titled help section.
func helpSection(t *testing.T, output, title string) []helpRowLine {
	t.Helper()
	start := strings.Index(output, title+"\n")
	if start < 0 {
		t.Fatalf("help output missing section %q:\n%s", title, output)
	}
	rest := output[start+len(title)+1:]
	end := strings.Index(rest, "\n\n")
	if end < 0 {
		end = len(rest)
	}

	re := regexp.MustCompile(`^   (\S.*?)(  +)(\S.*)$`)
	var rows []helpRowLine
	for _, line := range strings.Split(rest[:end], "\n") {
		if line == "" {
			continue
		}
		m := re.FindStringSubmatch(line)
		if m == nil {
			// Rows whose description is empty are printed without padding.
			rows = append(rows, helpRowLine{left: strings.TrimSpace(line)})
			continue
		}
		rows = append(rows, helpRowLine{
			left:     m[1],
			right:    m[3],
			rightCol: 3 + len([]rune(m[1])) + len(m[2]),
		})
	}
	if len(rows) == 0 {
		t.Fatalf("section %q has no rows:\n%s", title, output)
	}
	return rows
}

// assertAligned checks that every description in a section starts in the
// same column: each left cell is padded to the widest one plus two spaces.
func assertAligned(t *testing.T, title string, rows []helpRowLine) {
	t.Helper()
	width := 0
	for _, row := range rows {
		if n := len([]rune(row.left)); n > width {
			width = n
		}
	}
	want := 3 + width + 2
	for _, row := range rows {
		if row.right == "" {
			continue // no description to align
		}
		if row.rightCol != want {
			t.Errorf("%s: %q starts at column %d, want %d", title, row.right, row.rightCol, want)
		}
	}
}

func Test_rootHelpLayout(t *testing.T) {
	origVersion, origCommit := Version, CommitHash
	defer func() {
		Version = origVersion
		CommitHash = origCommit
	}()
	Version = "2.7.0"
	CommitHash = "bdcd306"

	stdout, stderr := captureHelp(t, func() {
		if err := rootCmd.Help(); err != nil {
			t.Errorf("rootCmd.Help() error = %v", err)
		}
	})

	if stderr != "" {
		t.Errorf("help must go to stdout, but stderr got: %q", stderr)
	}

	// Sections appear in order.
	sections := []string{
		"NAME:\n",
		"USAGE:\n",
		"VERSION:\n",
		"AUTHORS:\n",
		"DESCRIPTION:\n",
		"COMMANDS:\n",
		"GLOBAL OPTIONS:\n",
	}
	prev := -1
	for _, section := range sections {
		idx := strings.Index(stdout, section)
		if idx < 0 {
			t.Errorf("help output missing section %q:\n%s", section, stdout)
			continue
		}
		if idx < prev {
			t.Errorf("section %q appears out of order:\n%s", section, stdout)
		}
		prev = idx
	}

	wants := []string{
		"NAME:\n   git-open - Open the current Git repository in your browser\n",
		"USAGE:\n   git-open [global options] [command [command options]]\n",
		"VERSION:\n   2.7.0 (rev:bdcd306)\n",
		"AUTHORS:\n   ZHAO CHUNQI <zcq.qiqi@gmail.com>\n",
		"   help, h",
		"   repo        Print the repository name",
		"   version     Show program version information",
	}
	for _, want := range wants {
		if !strings.Contains(stdout, want) {
			t.Errorf("help output missing %q:\n%s", want, stdout)
		}
	}

	// Flags are written long-name first: "--help, -h", not "-h, --help".
	if strings.Contains(stdout, "-h, --help") || strings.Contains(stdout, "-v, --version") {
		t.Errorf("flags must be rendered long-name first:\n%s", stdout)
	}
	for _, want := range []string{"--help, -h", "--version, -v", "--chdir, -C", "--plain, -p"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("help output missing flag %q:\n%s", want, stdout)
		}
	}
	if !regexp.MustCompile(`(?m)^   --help, -h +show help$`).MatchString(stdout) {
		t.Errorf("--help row should read \"show help\":\n%s", stdout)
	}

	assertAligned(t, "COMMANDS", helpSection(t, stdout, "COMMANDS:"))
	assertAligned(t, "GLOBAL OPTIONS", helpSection(t, stdout, "GLOBAL OPTIONS:"))
}

func Test_rootHelpCommandDescriptions(t *testing.T) {
	stdout, _ := captureHelp(t, func() {
		if err := rootCmd.Help(); err != nil {
			t.Errorf("rootCmd.Help() error = %v", err)
		}
	})

	for _, row := range helpSection(t, stdout, "COMMANDS:") {
		if row.right == "" {
			t.Errorf("command %q has no description", row.left)
		}
	}
}

func Test_subcommandHelpLayout(t *testing.T) {
	stdout, stderr := captureHelp(t, func() {
		if err := repoCmd.Help(); err != nil {
			t.Errorf("repoCmd.Help() error = %v", err)
		}
	})

	if stderr != "" {
		t.Errorf("help must go to stdout, but stderr got: %q", stderr)
	}

	wants := []string{
		"NAME:\n   git-open repo - Print the repository name\n",
		"USAGE:\n   git-open repo [command options] [arguments...]\n",
		"OPTIONS:\n   --help, -h  show help\n",
		"GLOBAL OPTIONS:\n",
	}
	for _, want := range wants {
		if !strings.Contains(stdout, want) {
			t.Errorf("help output missing %q:\n%s", want, stdout)
		}
	}

	// VERSION and AUTHORS only appear in the root help.
	for _, unwanted := range []string{"VERSION:", "AUTHORS:"} {
		if strings.Contains(stdout, unwanted) {
			t.Errorf("subcommand help must not contain %q:\n%s", unwanted, stdout)
		}
	}

	assertAligned(t, "OPTIONS", helpSection(t, stdout, "OPTIONS:"))
	assertAligned(t, "GLOBAL OPTIONS", helpSection(t, stdout, "GLOBAL OPTIONS:"))
}

func Test_usageGoesToStderr(t *testing.T) {
	stdout, stderr := captureHelp(t, func() {
		if err := rootCmd.Usage(); err != nil {
			t.Errorf("rootCmd.Usage() error = %v", err)
		}
	})

	if stdout != "" {
		t.Errorf("usage must go to stderr, but stdout got: %q", stdout)
	}
	if !strings.Contains(stderr, "NAME:\n   git-open - ") {
		t.Errorf("stderr usage missing NAME section:\n%s", stderr)
	}
}

func findSubCommand(t *testing.T, name string) *cobra.Command {
	t.Helper()
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == name {
			return sub
		}
	}
	rootCmd.InitDefaultHelpCmd()
	rootCmd.InitDefaultCompletionCmd()
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == name {
			return sub
		}
	}
	t.Fatalf("subcommand %q not found", name)
	return nil
}

func Test_usageText(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
		want string
	}{
		{
			name: "root",
			cmd:  rootCmd,
			want: "git-open [global options] [command [command options]]",
		},
		{
			name: "leaf subcommand",
			cmd:  repoCmd,
			want: "git-open repo [command options] [arguments...]",
		},
		{
			name: "nested command",
			cmd:  findSubCommand(t, "completion"),
			want: "git-open completion [command [command options]]",
		},
		{
			name: "nested leaf",
			cmd:  findSubCommand(t, "help"),
			want: "git-open help [command options] [arguments...]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := usageText(tt.cmd); got != tt.want {
				t.Errorf("usageText(%s) = %q, want %q", tt.cmd.CommandPath(), got, tt.want)
			}
		})
	}
}

// Test_writeHelpBareCommand renders help for a synthetic command tree to
// cover the branches the real commands never hit: a NAME without a
// description, a subcommand without a description, and a hidden
// subcommand that must be skipped.
func Test_writeHelpBareCommand(t *testing.T) {
	root := &cobra.Command{Use: "demo"}
	root.CompletionOptions.DisableDefaultCmd = true
	root.AddCommand(&cobra.Command{Use: "plain", Run: func(*cobra.Command, []string) {}})
	root.AddCommand(&cobra.Command{Use: "secret", Hidden: true})

	buf := new(bytes.Buffer)
	writeHelp(root, buf)
	out := buf.String()

	if !strings.Contains(out, "NAME:\n   demo\n") {
		t.Errorf("NAME section should omit the description:\n%s", out)
	}
	if !strings.Contains(out, "USAGE:\n   demo [global options] [command [command options]]\n") {
		t.Errorf("USAGE section mismatch:\n%s", out)
	}

	rows := helpSection(t, out, "COMMANDS:")
	var listed []string
	for _, row := range rows {
		listed = append(listed, row.left)
		if row.left == "plain" && row.right != "" {
			t.Errorf("command without Short must have an empty description, got %q", row.right)
		}
	}
	if slicesContains(listed, "secret") {
		t.Errorf("hidden command must not be listed: %v", listed)
	}
	if !slicesContains(listed, "plain") {
		t.Errorf("COMMANDS missing %q: %v", "plain", listed)
	}
	assertAligned(t, "COMMANDS", rows)
}

// Test_writeHelpRootWithoutSubCommands covers the USAGE wording for a root
// command that has no subcommands.
func Test_writeHelpRootWithoutSubCommands(t *testing.T) {
	root := &cobra.Command{Use: "lonely", Short: "A command without subcommands"}
	root.CompletionOptions.DisableDefaultCmd = true

	buf := new(bytes.Buffer)
	writeHelp(root, buf)
	out := buf.String()

	if !strings.Contains(out, "USAGE:\n   lonely [global options] [arguments...]\n") {
		t.Errorf("USAGE section mismatch:\n%s", out)
	}
	if strings.Contains(out, "COMMANDS:") {
		t.Errorf("COMMANDS section must be omitted:\n%s", out)
	}
}

// Test_writeAuthorsEmpty covers the empty AUTHORS list.
func Test_writeAuthorsEmpty(t *testing.T) {
	orig := helpAuthors
	helpAuthors = nil
	defer func() { helpAuthors = orig }()

	buf := new(bytes.Buffer)
	writeAuthors(buf)
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty authors, got %q", buf.String())
	}
}

// Test_flagRows covers flag formatting: nil flag sets, hidden flags,
// flags without a shorthand, and default-value suffixes.
func Test_flagRows(t *testing.T) {
	if rows := flagRows(nil); len(rows) != 0 {
		t.Errorf("flagRows(nil) = %v, want no rows", rows)
	}

	fs := pflag.NewFlagSet("demo", pflag.ContinueOnError)
	fs.String("token", "abc", "api token")
	fs.Int("count", 3, "retry count")
	fs.Bool("flag", false, "a flag")
	fs.String("secret", "s3cr3t", "hidden flag")
	if err := fs.MarkHidden("secret"); err != nil {
		t.Fatalf("MarkHidden: %v", err)
	}

	rows := flagRows(fs)
	byLeft := make(map[string]string, len(rows))
	for _, row := range rows {
		if strings.Contains(row.left, "secret") {
			t.Errorf("hidden flag must not be listed: %q", row.left)
		}
		byLeft[row.left] = row.right
	}

	wants := map[string]string{
		"--token value": `api token (default "abc")`, // string default, quoted
		"--count value": "retry count (default 3)",   // non-string default
		"--flag":        "a flag",                    // zero default, no suffix
	}
	for left, want := range wants {
		got, ok := byLeft[left]
		if !ok {
			t.Errorf("missing flag row %q in %v", left, byLeft)
			continue
		}
		if got != want {
			t.Errorf("flag %q: description = %q, want %q", left, got, want)
		}
	}
}

func slicesContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
