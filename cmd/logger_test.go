/*
 * Minio Cloud Storage (C) 2015 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Sirupsen/logrus"
)

// Tests callerLocation.
func TestCallerLocation(t *testing.T) {
	currentLocation := func() string { return callerLocation() }
	gotLocation := currentLocation()
	expectedLocation := "[logger_test.go:31:TestCallerLocation()]"
	if gotLocation != expectedLocation {
		t.Errorf("expected : %s, got : %s", expectedLocation, gotLocation)
	}
}

// Tests error logger.
func TestLogger(t *testing.T) {
	var buffer bytes.Buffer
	var fields logrus.Fields
	log.Out = &buffer
	log.Formatter = new(logrus.JSONFormatter)

	errorIf(errors.New("Fake error"), "Failed with error.")
	err := json.Unmarshal(buffer.Bytes(), &fields)
	if err != nil {
		t.Fatal(err)
	}
	if fields["level"] != "error" {
		t.Fatalf("Expected error, got %s", fields["level"])
	}
	msg, ok := fields["cause"]
	if !ok {
		t.Fatal("Cause field missing")
	}
	if msg != "Fake error" {
		t.Fatal("Cause field has unexpected message", msg)
	}
}
