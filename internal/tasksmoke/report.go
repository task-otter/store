// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

func durationText(duration time.Duration) string {
	return duration.Round(time.Millisecond).String()
}

func flushReport(sink reportSink) error {
	err := sink.Flush()
	if err != nil {
		return fmt.Errorf("flush smoke report: %w", err)
	}

	return nil
}

func formattedRow(result *taskResult) string {
	return fmt.Sprintf(
		tableRowFormat,
		result.Module,
		result.Task,
		result.Status,
		durationText(result.Duration),
	)
}

func hasFailures(results []*taskResult) bool {
	for i := range results {
		if results[i].Status == statusFail {
			return true
		}
	}

	return false
}

func newTableWriter(stdout io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(
		stdout,
		emptyLength,
		emptyLength,
		nameSplitParts,
		tabPadChar,
		emptyLength,
	)
}

func writeAligned(sink reportSink, results []*taskResult) error {
	err := writeReportRows(sink, results)
	if err != nil {
		return fmt.Errorf("write smoke report rows: %w", err)
	}

	flushErr := flushReport(sink)
	if flushErr != nil {
		return fmt.Errorf("flush aligned report: %w", flushErr)
	}

	return nil
}

func writeReport(stdout io.Writer, results []*taskResult) error {
	err := writeAligned(newTableWriter(stdout), results)
	if err != nil {
		return fmt.Errorf("write aligned smoke report: %w", err)
	}

	return nil
}

func writeReportRows(writer io.Writer, results []*taskResult) error {
	err := writeHeader(writer)
	if err != nil {
		return fmt.Errorf("write smoke report header: %w", err)
	}

	rowsErr := writeResultRows(writer, results)
	if rowsErr != nil {
		return fmt.Errorf("write smoke report rows body: %w", rowsErr)
	}

	return nil
}

func writeHeader(writer io.Writer) error {
	written, err := io.WriteString(writer, tableHeader)
	if err != nil {
		return fmt.Errorf(errReportHeader, err)
	}

	if written == emptyLength && tableHeader != emptyString {
		return fmt.Errorf(errReportHeader, io.ErrShortWrite)
	}

	return nil
}

func writeResultRow(writer io.Writer, result *taskResult) error {
	text := formattedRow(result)

	written, err := io.WriteString(writer, text)
	if err != nil {
		return fmt.Errorf(errReportRow, err)
	}

	if written == emptyLength && text != emptyString {
		return fmt.Errorf(errReportRow, io.ErrShortWrite)
	}

	return nil
}

func writeResultRows(writer io.Writer, results []*taskResult) error {
	for i := range results {
		err := writeResultRow(writer, results[i])
		if err != nil {
			return fmt.Errorf("write smoke report row: %w", err)
		}
	}

	return nil
}
