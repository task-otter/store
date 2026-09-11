// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"flag"
	"fmt"
	"io"
	"os"
)

// Main is the tasksmoke CLI entrypoint. It returns a process exit code.
func Main(args []string, stdout, stderr io.Writer) int {
	flags, err := parseFlags(args, stderr)
	if err != nil {
		return nameSplitParts
	}

	return runMain(flags, stdout, stderr)
}

func argsWithoutProg(args []string) []string {
	if len(args) == emptyLength {
		return nil
	}

	return args[exitFail:]
}

func bindFlags(set *flag.FlagSet, flags *cliFlags) {
	set.StringVar(&flags.Module, flagModule, emptyString, moduleUsage)
	set.BoolVar(&flags.List, flagList, false, listUsage)
}

func newFlagSet(stderr io.Writer) *flag.FlagSet {
	set := flag.NewFlagSet(appName, flag.ContinueOnError)
	set.SetOutput(stderr)

	return set
}

func parseFlags(args []string, stderr io.Writer) (*cliFlags, error) {
	flags := new(cliFlags)
	set := newFlagSet(stderr)

	bindFlags(set, flags)

	parsed, err := parsedFlags(set, flags, args)
	if err != nil {
		return nil, fmt.Errorf("parse cli flags: %w", err)
	}

	return parsed, nil
}

func parsedFlags(set *flag.FlagSet, flags *cliFlags, args []string) (*cliFlags, error) {
	err := set.Parse(argsWithoutProg(args))
	if err != nil {
		return nil, fmt.Errorf("parse flags: %w", err)
	}

	return flags, nil
}

func printErr(stderr io.Writer, err error, code int) int {
	written, writeErr := fmt.Fprintf(stderr, "%v\n", err)
	if writeErr != nil {
		return code
	}

	if written == emptyLength {
		return code
	}

	return code
}

func reportExitCode(report *smokeReport) int {
	if hasFailures(report.Results) {
		return exitFail
	}

	return emptyLength
}

func runMain(flags *cliFlags, stdout, stderr io.Writer) int {
	root, err := detectRepoRootFrom(os.Getwd)

	return startCLI(&startedCLI{
		Err:    err,
		Flags:  flags,
		Root:   root,
		Stderr: stderr,
		Stdout: stdout,
	})
}

func runSmoke(run *cliRun) (*smokeReport, error) {
	report, err := runSuite(run.Engine, &runOptions{
		ListOnly: run.Flags.List,
		Module:   run.Flags.Module,
		RepoRoot: run.Root,
		Stderr:   run.Stderr,
		Stdout:   run.Stdout,
	})
	if err != nil {
		return nil, fmt.Errorf(errWrapFormat, errSmokeEngine, err)
	}

	return report, nil
}

func runWithRoot(run *cliRun) int {
	report, err := runSmoke(run)
	if err != nil {
		return printErr(run.Stderr, err, exitFail)
	}

	return writeAndExit(run.Stdout, run.Stderr, report)
}

func startCLI(started *startedCLI) int {
	if started.Err != nil {
		return printErr(started.Stderr, started.Err, exitFail)
	}

	return runWithRoot(&cliRun{
		Engine: newEngine(),
		Flags:  started.Flags,
		Root:   started.Root,
		Stdout: started.Stdout,
		Stderr: started.Stderr,
	})
}

func writeAndExit(stdout, stderr io.Writer, report *smokeReport) int {
	err := writeReport(stdout, report.Results)
	if err != nil {
		return printErr(stderr, err, exitFail)
	}

	return reportExitCode(report)
}
