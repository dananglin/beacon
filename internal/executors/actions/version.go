package actions

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/info"
)

type Version struct {
	*flag.FlagSet

	showFullVersion bool
}

func NewVersion() *Version {
	name := "version"

	version := Version{
		FlagSet: flag.NewFlagSet(name, flag.ExitOnError),
	}

	version.BoolVar(&version.showFullVersion, "full", false, "Print the applications full build information")

	return &version
}

func (a *Version) Execute(args []string) error {
	if err := a.Parse(args); err != nil {
		return fmt.Errorf("(version) flag parsing error: %w", err)
	}

	a.printVersion()

	return nil
}

func (a *Version) printVersion() {
	if !a.showFullVersion {
		fmt.Fprintf(os.Stdout, "%s %s\n", info.ApplicationName, info.BinaryVersion)

		return
	}

	var builder strings.Builder

	builder.WriteString(info.ApplicationName + "\n\n")

	tableWriter := tabwriter.NewWriter(&builder, 0, 4, 1, ' ', 0)

	_, _ = tableWriter.Write([]byte("Version:" + "\t" + info.BinaryVersion + "\n"))
	_, _ = tableWriter.Write([]byte("Git commit:" + "\t" + info.GitCommit + "\n"))
	_, _ = tableWriter.Write([]byte("Go version:" + "\t" + info.GoVersion + "\n"))
	_, _ = tableWriter.Write([]byte("Build date:" + "\t" + info.BuildTime + "\n"))

	_ = tableWriter.Flush()

	_, _ = os.Stdout.WriteString(builder.String())
}
