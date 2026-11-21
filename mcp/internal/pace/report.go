package pace

import (
	"fmt"
	"strings"
	"time"
)

func evaluateSession(sess Session, now time.Time) PaceEvaluation {
	elapsedSec := secondsBetween(sess.LastEditAt, now)
	remSec := max(0, sess.TimeLimitSec-elapsedSec)
	remCalls := max(0, sess.CallLimit-sess.NonEditCalls)

	reasons := []string{}
	if elapsedSec >= sess.TimeLimitSec {
		reasons = append(reasons, fmt.Sprintf("elapsed time (%ds) reached time limit (%ds)", elapsedSec, sess.TimeLimitSec))
	}
	if sess.NonEditCalls >= sess.CallLimit {
		reasons = append(reasons, fmt.Sprintf("non-edit tool calls (%d) reached call limit (%d)", sess.NonEditCalls, sess.CallLimit))
	}
	if sess.ConsecutiveFailures >= cyclicFailureLimit {
		reasons = append(reasons, fmt.Sprintf("cyclic failure loop detected (%d consecutive failures >= %d)", sess.ConsecutiveFailures, cyclicFailureLimit))
	}

	status := PaceStatusOK
	report := ""
	if len(reasons) > 0 {
		status = PaceStatusExceeded
		report = renderCheckpointReport(sess, elapsedSec, reasons)
	}

	return PaceEvaluation{
		OK:               true,
		PaceID:           sess.PaceID,
		Status:           status,
		ElapsedSec:       elapsedSec,
		TimeLimitSec:     sess.TimeLimitSec,
		RemainingSec:     remSec,
		NonEditCalls:     sess.NonEditCalls,
		CallLimit:        sess.CallLimit,
		RemainingCalls:   remCalls,
		LastEditedFile:   sess.LastEditedFile,
		Reasons:          reasons,
		CheckpointReport: report,
		Session:          sess,
	}
}

func renderCheckpointReport(sess Session, elapsedSec int, reasons []string) string {
	lastFile := sess.LastEditedFile
	if lastFile == "" {
		lastFile = "None yet"
	}
	note := sanitizeSingleLine(sess.LastNote)
	if note == "" {
		note = "Exploratory steps in progress"
	}
	var b strings.Builder
	b.WriteString("### ⏸️ Paced Execution Checkpoint\n\n")
	_, _ = fmt.Fprintf(&b, "**Current Threshold**: %ds / %d non-edit calls (Current: %ds elapsed, %d calls)\n",
		sess.TimeLimitSec, sess.CallLimit, elapsedSec, sess.NonEditCalls)
	_, _ = fmt.Fprintf(&b, "**Last File Edited**: %s\n\n", lastFile)
	b.WriteString("#### 1. Current Status & Findings\n")
	_, _ = fmt.Fprintf(&b, "- %s\n\n", note)
	b.WriteString("#### 2. Reason for Delay / Obstacle\n")
	for _, r := range reasons {
		_, _ = fmt.Fprintf(&b, "- %s\n", sanitizeSingleLine(r))
	}
	b.WriteString("\n#### 3. Proposed Next Actions\n")
	b.WriteString("- **Option A (Recommended)**: Proceed directly to the target file edit.\n")
	b.WriteString("- **Option B**: Adjust pacing thresholds and continue investigation.\n")
	b.WriteString("- **Option C**: Pause and revise the implementation plan.\n\n")
	b.WriteString("#### 4. User Judgment Needed\nHow would you like to proceed?\n")
	return b.String()
}

func renderStallReport(entry WatchEntry, elapsedSec, idleSec int, cpuPct float64, rssKB int64, reasons []string) string {
	cmdDesc := sanitizeSingleLine(entry.Command)
	if cmdDesc == "" {
		cmdDesc = fmt.Sprintf("PID %d", entry.PID)
	}
	var b strings.Builder
	b.WriteString("### ⚠️ Command Execution Stall Report\n\n")
	_, _ = fmt.Fprintf(&b, "**Monitored Command / Task**: `%s`\n", cmdDesc)
	_, _ = fmt.Fprintf(&b, "**Execution Mode**: `%s` (Check Interval: %ds)\n", entry.Mode, entry.CheckIntervalSec)
	_, _ = fmt.Fprintf(&b, "**Elapsed Time**: %ds (Idle: %ds)\n\n", elapsedSec, idleSec)
	b.WriteString("#### 1. Observed Liveness Signals\n")
	_, _ = fmt.Fprintf(&b, "- **Resource Usage**: CPU %.1f%%, RSS %d KB\n", cpuPct, rssKB)
	for _, r := range reasons {
		_, _ = fmt.Fprintf(&b, "- **Observation**: %s\n", sanitizeSingleLine(r))
	}
	b.WriteString("\n#### 2. Action Options for User\n")
	b.WriteString("- **Option A (Abort & Diagnose)**: Terminate the command/task and inspect logs.\n")
	b.WriteString("- **Option B (Extend Wait)**: Continue waiting for another check interval.\n")
	b.WriteString("- **Option C (Manual Inspection)**: Pause for user manual intervention.\n\n")
	b.WriteString("#### 3. User Judgment Needed\nDo you want to terminate this command/task, or continue waiting?\n")
	return b.String()
}

func sanitizeSingleLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}
