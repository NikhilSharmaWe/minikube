/*
Copyright 2020 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package audit

import (
	"os"
	"os/user"
	"testing"

	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
)

func TestAudit(t *testing.T) {
	t.Run("username", func(t *testing.T) {
		u, err := user.Current()
		if err != nil {
			t.Fatal(err)
		}

		tests := []struct {
			userFlag string
			want     string
		}{
			{
				"testUser",
				"testUser",
			},
			{
				"",
				u.Username,
			},
		}

		for _, test := range tests {
			viper.Set(config.UserFlag, test.userFlag)

			got := userName()

			if got != test.want {
				t.Errorf("userFlag = %q; username() = %q; want %q", test.userFlag, got, test.want)
			}
		}
	})

	t.Run("args", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()

		tests := []struct {
			args []string
			want string
		}{
			{
				[]string{"minikube", "start"},
				"",
			},
			{
				[]string{"minikube", "start", "--user", "testUser"},
				"--user testUser",
			},
		}

		for _, test := range tests {
			os.Args = test.args

			got := args()

			if got != test.want {
				t.Errorf("os.Args = %q; args() = %q; want %q", os.Args, got, test.want)
			}
		}
	})

	t.Run("shouldLog", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()

		tests := []struct {
			args []string
			want bool
		}{
			{
				[]string{"minikube", "start"},
				true,
			},
			{
				[]string{"minikube", "delete"},
				true,
			},
			{
				[]string{"minikube", "status"},
				false,
			},
			{
				[]string{"minikube", "version"},
				false,
			},
			{
				[]string{"minikube"},
				false,
			},
			{
				[]string{"minikube", "delete", "--purge"},
				false,
			},
		}

		for _, test := range tests {
			os.Args = test.args

			got := shouldLog()

			if got != test.want {
				t.Errorf("os.Args = %q; shouldLog() = %t; want %t", os.Args, got, test.want)
			}
		}
	})

	t.Run("isDeletePurge", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()

		tests := []struct {
			args []string
			want bool
		}{
			{
				[]string{"minikube", "delete", "--purge"},
				true,
			},
			{
				[]string{"minikube", "delete"},
				false,
			},
			{
				[]string{"minikube", "start", "--purge"},
				false,
			},
			{
				[]string{"minikube"},
				false,
			},
		}

		for _, test := range tests {
			os.Args = test.args

			got := isDeletePurge()

			if got != test.want {
				t.Errorf("os.Args = %q; isDeletePurge() = %t; want %t", os.Args, got, test.want)
			}
		}
	})

	// Check if logging with limited args causes a panic
	t.Run("LogCommandStart", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"minikube", "start"}

		auditID, err := LogCommandStart()
		if auditID == "" {
			t.Fatal("audit ID should not be empty")
		}
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("LogCommandEnd", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"minikube"}

		auditID, err := LogCommandStart()
		if err != nil {
			t.Fatal("start failed")
		}
		err = LogCommandEnd(auditID)

		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("LogCommandEnd", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"minikube"}

		err := LogCommandEnd("non-existing-id")

		if err == nil {
			t.Fatal("function LogCommandEnd should return an error when a non-existing id is passed in it as an argument")
		}
	})
}
