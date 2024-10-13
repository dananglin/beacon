package executors

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/info"
)

type versionExecutor struct {
	*flag.FlagSet

	showFullVersion bool
}

func executeVersionCommand(args []string) error {
	executorName := "version"

	executor := versionExecutor{
		FlagSet: flag.NewFlagSet(executorName, flag.ExitOnError),
	}

	executor.BoolVar(&executor.showFullVersion, "full", false, "Print the applications full build information")

	if err := executor.Parse(args); err != nil {
		return fmt.Errorf("(%s) flag parsing error: %w", executorName, err)
	}

	executor.printVersion()

	return nil
}

func (e *versionExecutor) printVersion() {
	if !e.showFullVersion {
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
