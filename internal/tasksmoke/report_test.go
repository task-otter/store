// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// TestWriteReportContainsStatuses exercises WriteReportContainsStatuses.
func TestWriteReportContainsStatuses(t *testing.T) {
	t.Parallel()

	output := new(bytes.Buffer)
	err := writeReport(output, sampleStatusRows())
	requireNoErr(t, err)
	requireReportHasStatus(t, output.String(), statusPass)
	requireReportHasStatus(t, output.String(), statusSkip)
}

func sampleStatusRows() []*taskResult {
	return []*taskResult{
		{
			Module:   testEcho,
			Task:     testPing,
			Status:   statusPass,
			Duration: time.Second,
			Err:      nil,
			Output:   emptyString,
		},
		{
			Module:   testGit,
			Task:     namePush,
			Status:   statusSkip,
			Duration: emptyLength,
			Err:      nil,
			Output:   reasonDestructive,
		},
	}
}

func requireReportHasStatus(t *testing.T, report, status string) {
	t.Helper()

	requireSame(t, strings.Contains(report, status), true)
}

// TestHasFailuresDetectsFail exercises HasFailuresDetectsFail.
func TestHasFailuresDetectsFail(t *testing.T) {
	t.Parallel()

	requireSame(t, hasFailures([]*taskResult{{Status: statusFail}}), true)
	requireSame(t, !hasFailures([]*taskResult{{Status: statusPass}}), true)
}

// TestWriteReportRejectsClosedWriter exercises WriteReportRejectsClosedWriter.
func TestWriteReportRejectsClosedWriter(t *testing.T) {
	t.Parallel()

	err := writeReport(failWriter{}, []*taskResult{
		{
			Module:   testEcho,
			Task:     testPing,
			Status:   statusPass,
			Duration: emptyLength,
			Err:      nil,
			Output:   emptyString,
		},
	})
	requireErr(t, err)
}

// TestDurationTextRounds exercises DurationTextRounds.
func TestDurationTextRounds(t *testing.T) {
	t.Parallel()

	requireEqual(t, durationText(time.Second), testOneSecond)
}

// TestWriteAlignedHeaderError exercises WriteAlignedHeaderError.
func TestWriteAlignedHeaderError(t *testing.T) {
	t.Parallel()

	err := writeAligned(&reportSinkStub{}, []*taskResult{
		{Module: testEcho, Task: testPing, Status: statusPass},
	})
	requireErr(t, err)
}

// TestWriteAlignedRowError exercises WriteAlignedRowError.
func TestWriteAlignedRowError(t *testing.T) {
	t.Parallel()

	err := writeAligned(&reportSinkStub{remain: exitFail}, []*taskResult{
		{Module: testEcho, Task: testPing, Status: statusPass},
	})
	requireErr(t, err)
}

// TestWriteAlignedFlushError exercises WriteAlignedFlushError.
func TestWriteAlignedFlushError(t *testing.T) {
	t.Parallel()

	err := writeAligned(
		&reportSinkStub{remain: nameSplitParts, flushErr: sentinelErr()},
		[]*taskResult{
			{Module: testEcho, Task: testPing, Status: statusPass},
		},
	)
	requireErr(t, err)
}
