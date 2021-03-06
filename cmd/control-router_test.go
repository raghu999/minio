/*
 * Minio Cloud Storage, (C) 2016 Minio, Inc.
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
	"net/url"
	"testing"
)

// Tests initialization of remote controller clients.
func TestInitRemoteControlClients(t *testing.T) {
	rootPath, err := newTestConfig("us-east-1")
	if err != nil {
		t.Fatal("Unable to initialize config", err)
	}
	defer removeAll(rootPath)

	testCases := []struct {
		isDistXL     bool
		srvCmdConfig serverCmdConfig
		totalClients int
	}{
		// Test - 1 no allocation if server config is not distributed XL.
		{
			isDistXL:     false,
			srvCmdConfig: serverCmdConfig{},
			totalClients: 0,
		},
		// Test - 2 two clients allocated with 4 disks with 2 disks on same node each.
		{
			isDistXL: true,
			srvCmdConfig: serverCmdConfig{
				endpoints: []*url.URL{{
					Scheme: "http",
					Host:   "10.1.10.1:9000",
					Path:   "/mnt/disk1",
				}, {
					Scheme: "http",
					Host:   "10.1.10.1:9000", Path: "/mnt/disk2",
				}, {
					Scheme: "http",
					Host:   "10.1.10.2:9000", Path: "/mnt/disk1",
				}, {
					Scheme: "http",
					Host:   "10.1.10.2:9000", Path: "/mnt/disk2"},
				},
			},
			totalClients: 2,
		},
		// Test - 3 4 clients allocated with 4 disks with 1 disk on each node.
		{
			isDistXL: true,
			srvCmdConfig: serverCmdConfig{
				endpoints: []*url.URL{{
					Scheme: "http",
					Host:   "10.1.10.1:9000", Path: "/mnt/disk1",
				}, {
					Scheme: "http",
					Host:   "10.1.10.2:9000", Path: "/mnt/disk2",
				}, {
					Scheme: "http",
					Host:   "10.1.10.3:9000", Path: "/mnt/disk1",
				}, {
					Scheme: "http",
					Host:   "10.1.10.4:9000", Path: "/mnt/disk2"},
				},
			},
			totalClients: 4,
		},
	}

	// Evaluate and validate all test cases.
	for i, testCase := range testCases {
		globalIsDistXL = testCase.isDistXL
		rclients := initRemoteControlClients(testCase.srvCmdConfig)
		if len(rclients) != testCase.totalClients {
			t.Errorf("Test %d, Expected %d, got %d RPC clients.", i+1, testCase.totalClients, len(rclients))
		}
	}
}
