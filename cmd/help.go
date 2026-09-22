package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// helpAuthors are listed in the AUTHORS section of the help output.
var helpAuthors = []string{"ZHAO CHUNQI <zcq.qiqi@gmail.com>"}

func init() {
	// Render help (and usage shown on errors) in a sectioned layout:
	//
	//	NAME: / USAGE: / VERSION: / AUTHORS: / DESCRIPTION:
	//	COMMANDS: / OPTIONS: / GLOBAL OPTIONS:
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		writeHelp(cmd, cmd.OutOrStdout())
	})
	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		writeHelp(cmd, cmd.ErrOrStderr())
		return nil
	})
}

// helpCommandName is cobra's built-in help command.
const helpCommandName = "help"

// ensureHelpCommand makes sure cmd has cobra's help command registered and
// adds an "h" alias to it, so COMMANDS lists "help, h" and "git-open h"
// works.
func ensureHelpCommand(cmd *cobra.Command) {
	cmd.InitDefaultHelpCmd()
	for _, sub := range cmd.Commands() {
		if sub.Name() == helpCommandName && len(sub.Aliases) == 0 {
			sub.Aliases = []string{"h"}
		}
	}
}

// helpRow is one line of the help output: a left column (command names or
// flags) and an optional right column (descriptions). Both columns are
// aligned per section, padded to the widest left cell plus two spaces.
type helpRow struct {
	left  string
	right string
}

// writeHelp renders cmd's help in the sectioned layout.
func writeHelp(cmd *cobra.Command, w io.Writer) {
	// Make sure the --help flag exists even when help is rendered directly
	// (e.g. from tests) without going through Command.execute.
	cmd.InitDefaultHelpFlag()
	if !cmd.HasParent() {
		// Ensure the help/completion subcommands are registered so the
		// COMMANDS section is identical no matter how help was triggered.
		ensureHelpCommand(cmd)
		cmd.InitDefaultCompletionCmd()
	}

	name := cmd.CommandPath()
	if short := strings.TrimSpace(cmd.Short); short != "" {
		fmt.Fprintf(w, "NAME:\n   %s - %s\n\n", name, short)
	} else {
		fmt.Fprintf(w, "NAME:\n   %s\n\n", name)
	}

	fmt.Fprintf(w, "USAGE:\n   %s\n\n", usageText(cmd))

	if !cmd.HasParent() {
		writeVersion(w)
		writeAuthors(w)
	}

	if long := strings.TrimSpace(cmd.Long); long != "" && long != cmd.Short {
		writeIndented(w, "DESCRIPTION:", long)
	}

	if rows := commandRows(cmd); len(rows) > 0 {
		writeRows(w, "COMMANDS:", rows)
	}

	if cmd.HasParent() {
		if rows := flagRows(cmd.LocalFlags()); len(rows) > 0 {
			writeRows(w, "OPTIONS:", rows)
		}
		if rows := flagRows(cmd.InheritedFlags()); len(rows) > 0 {
			writeRows(w, "GLOBAL OPTIONS:", rows)
		}
	} else if rows := flagRows(cmd.LocalFlags()); len(rows) > 0 {
		writeRows(w, "GLOBAL OPTIONS:", rows)
	}
}

// usageText builds the USAGE line for cmd.
func usageText(cmd *cobra.Command) string {
	// Ensure the auto-generated --help flag exists so that leaf commands
	// always report "[command options]", no matter how help was triggered.
	cmd.InitDefaultHelpFlag()
	path := cmd.CommandPath()

	switch {
	case !cmd.HasParent() && cmd.HasAvailableSubCommands():
		return path + " [global options] [command [command options]]"
	case !cmd.HasParent():
		return path + " [global options] [arguments...]"
	case cmd.HasAvailableSubCommands():
		return path + " [command [command options]]"
	default:
		return path + " [command options] [arguments...]"
	}
}

// writeVersion writes the VERSION section of the root help.
func writeVersion(w io.Writer) {
	version := Version
	if CommitHash != "" && CommitHash != "none" {
		version += " (rev:" + CommitHash + ")"
	}
	fmt.Fprintf(w, "VERSION:\n   %s\n\n", version)
}

// writeAuthors writes the AUTHORS section of the root help.
func writeAuthors(w io.Writer) {
	if len(helpAuthors) == 0 {
		return
	}
	fmt.Fprint(w, "AUTHORS:\n")
	for _, author := range helpAuthors {
		fmt.Fprintf(w, "   %s\n", author)
	}
	fmt.Fprintln(w)
}

// writeIndented writes a titled section whose text is indented by three
// spaces on every line (used for DESCRIPTION).
func writeIndented(w io.Writer, title, text string) {
	fmt.Fprintf(w, "%s\n", title)
	for _, line := range strings.Split(text, "\n") {
		fmt.Fprintf(w, "   %s\n", line)
	}
	fmt.Fprintln(w)
}

// writeRows writes a titled section of rows, aligning the right column to
// the widest left cell plus two spaces.
func writeRows(w io.Writer, title string, rows []helpRow) {
	width := 0
	for _, row := range rows {
		if n := len([]rune(row.left)); n > width {
			width = n
		}
	}

	fmt.Fprintf(w, "%s\n", title)
	for _, row := range rows {
		if row.right == "" {
			fmt.Fprintf(w, "   %s\n", row.left)
			continue
		}
		fmt.Fprintf(w, "   %-*s  %s\n", width, row.left, row.right)
	}
	fmt.Fprintln(w)
}

// commandRows builds the COMMANDS rows for the subcommands of cmd.
func commandRows(cmd *cobra.Command) []helpRow {
	var rows []helpRow
	for _, sub := range cmd.Commands() {
		// IsAvailableCommand excludes the built-in help command, which is
		// still listed in the help output (e.g. "help, h").
		if !sub.IsAvailableCommand() && sub.Name() != "help" {
			continue
		}
		rows = append(rows, helpRow{left: sub.NameAndAliases(), right: sub.Short})
	}
	return rows
}

// flagRows builds help rows for the visible (non-hidden) flags of fs.
func flagRows(fs *pflag.FlagSet) []helpRow {
	var rows []helpRow
	if fs == nil {
		return rows
	}
	fs.VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		rows = append(rows, flagRow(f))
	})
	return rows
}

// flagRow formats a single flag as "   --long, -s value   usage": long name
// first, then the shorthand.
func flagRow(f *pflag.Flag) helpRow {
	left := "--" + f.Name
	if f.Shorthand != "" {
		left += ", -" + f.Shorthand
	}
	if f.NoOptDefVal == "" {
		left += " value"
	}

	usage := f.Usage
	if f.Name == "help" {
		usage = "show help"
	}
	if def := flagDefaultText(f); def != "" {
		usage += def
	}
	return helpRow{left: left, right: usage}
}

// flagDefaultText renders the "(default ...)" suffix for flags with a
// non-zero default value, omitting it for the usual zero values.
func flagDefaultText(f *pflag.Flag) string {
	switch f.DefValue {
	case "", "false", "0", "[]", "map[]", "<nil>":
		return ""
	}
	if strings.HasPrefix(f.Value.Type(), "string") {
		return fmt.Sprintf(" (default %q)", f.DefValue)
	}
	return fmt.Sprintf(" (default %s)", f.DefValue)
}
