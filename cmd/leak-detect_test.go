// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"runtime/debug"
	"sort"
	"strings"
	"time"
)

const (
	// deadline (in seconds) up to which the go routine leak detection has to be retried.
	leakDetectDeadline = 5
	// pause time (in milliseconds) between each snapshot at the end of the go routine leak detection.
	leakDetectPauseTimeMs = 50
)

// LeakDetect - type with  methods for go routine leak detection.
type LeakDetect struct {
	relevantRoutines map[string]bool
}

// NewLeakDetect - Initialize a LeakDetector with the snapshot of relevant Go routines.
func NewLeakDetect() LeakDetect {
	snapshot := LeakDetect{
		relevantRoutines: make(map[string]bool),
	}
	for _, g := range pickRelevantGoroutines() {
		snapshot.relevantRoutines[g] = true
	}
	return snapshot
}

// CompareCurrentSnapshot - Compares the initial relevant stack trace with the current one (during the time of invocation).
func (initialSnapShot LeakDetect) CompareCurrentSnapshot() []string {
	var stackDiff []string
	for _, g := range pickRelevantGoroutines() {
		// Identify the Go routines those were not present in the initial snapshot.
		// In other words a stack diff.
		if !initialSnapShot.relevantRoutines[g] {
			stackDiff = append(stackDiff, g)
		}
	}
	return stackDiff
}

// DetectLeak - Creates a snapshot of runtime stack and compares it with the initial stack snapshot.
func (initialSnapShot LeakDetect) DetectLeak(t TestErrHandler) {
	if t.Failed() {
		return
	}
	// Loop, waiting for goroutines to shut down.
	// Wait up to 5 seconds, but finish as quickly as possible.
	deadline := time.Now().UTC().Add(leakDetectDeadline * time.Second)
	for {
		// get sack snapshot of relevant go routines.
		leaked := initialSnapShot.CompareCurrentSnapshot()
		// current stack snapshot matches the initial one, no leaks, return.
		if len(leaked) == 0 {
			return
		}
		// wait a test again will deadline.
		if time.Now().UTC().Before(deadline) {
			time.Sleep(leakDetectPauseTimeMs * time.Millisecond)
			continue
		}
		// after the deadline time report all the difference in the latest snapshot compared with the initial one.
		for _, g := range leaked {
			t.Errorf("Leaked goroutine: %v", g)
		}
		return
	}
}

// DetectTestLeak -  snapshots the currently running goroutines and returns a
// function to be run at the end of tests to see whether any
// goroutines leaked.
// Usage: `defer DetectTestLeak(t)()` in beginning line of benchmarks or unit tests.
func DetectTestLeak(t TestErrHandler) func() {
	initialStackSnapShot := NewLeakDetect()
	return func() {
		initialStackSnapShot.DetectLeak(t)
	}
}

// list of functions to be ignored from the stack trace.
// Leak detection is done when tests are run, should ignore the tests related functions,
// and other runtime functions while identifying leaks.
var ignoredStackFns = []string{
	"",
	// Below are the stacks ignored by the upstream leaktest code.
	"testing.Main(",
	"testing.tRunner(",
	"testing.tRunner(",
	"runtime.goexit",
	"created by runtime.gc",
	// ignore the snapshot function.
	// since the snapshot is taken here the entry will have the current function too.
	"pickRelevantGoroutines",
	"runtime.MHeap_Scavenger",
	"signal.signal_recv",
	"sigterm.handler",
	"runtime_mcall",
	"goroutine in C code",
}

// Identify whether the stack trace entry is part of ignoredStackFn .
func isIgnoredStackFn(stack string) (ok bool) {
	ok = true
	for _, stackFn := range ignoredStackFns {
		if !strings.Contains(stack, stackFn) {
			ok = false
			continue
		}
		break
	}
	return ok
}

// pickRelevantGoroutines returns all goroutines we care about for the purpose
// of leak checking. It excludes testing or runtime ones.
func pickRelevantGoroutines() (gs []string) {
	// get runtime stack buffer.
	buf := debug.Stack()
	// runtime stack of go routines will be listed with 2 blank spaces between each of them, so split on "\n\n" .
	for _, g := range strings.Split(string(buf), "\n\n") {
		// Again split on a new line, the first line of the second half contaisn the info about the go routine.
		sl := strings.SplitN(g, "\n", 2)
		if len(sl) != 2 {
			continue
		}
		stack := strings.TrimSpace(sl[1])
		// ignore the testing go routine.
		// since the tests will be invoking the leaktest it would contain the test go routine.
		if strings.HasPrefix(stack, "testing.RunTests") {
			continue
		}
		// Ignore the following go routines.
		// testing and run time go routines should be ignored, only the application generated go routines should be taken into account.
		if isIgnoredStackFn(stack) {
			continue
		}
		gs = append(gs, g)
	}
	sort.Strings(gs)
	return
}
