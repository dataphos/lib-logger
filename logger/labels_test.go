// Copyright 2024 Syntio Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger_test

import (
	"testing"

	"github.com/dataphos/lib-logger/logger"
)

func TestNewLabels(t *testing.T) {
	labels := logger.Labels{"key0": "val0", "key1": "val1"}

	if labels["key0"] != "val0" ||
		labels["key1"] != "val1" {
		t.Error("Missing information.")
	}
}

func TestLabels_Add(t *testing.T) {
	labels := logger.Labels{"key0": "val0", "key1": "val1"}
	labels.Add(logger.Labels{"key2": "val2", "key3": "val3"})

	if labels["key2"] != "val2" || labels["key3"] != "val3" {
		t.Error("Not added.")
	}
}

func TestLabels_AddWithLAlias(t *testing.T) {
	labels := logger.Labels{"key0": "val0", "key1": "val1"}
	labels.Add(logger.L{"key2": "val2", "key3": "val3"})

	if labels["key2"] != "val2" || labels["key3"] != "val3" {
		t.Error("Not added.")
	}
}

func TestLabels_Del(t *testing.T) {
	labels := logger.Labels{"key0": "val0", "key1": "val1"}
	labels.Del("key1")

	if _, ok := labels["key1"]; ok {
		t.Error("Not deleted.")
	}
}

func TestLabels_DelMultiple(t *testing.T) {
	labels := logger.Labels{"key0": "val0", "key1": "val1"}
	labels.Del("key0", "key1")

	if len(labels) != 0 {
		t.Error("Not deleted.")
	}
}

func TestLabels_Clone(t *testing.T) {
	original := logger.Labels{"key0": "val0"}
	clone := original.Clone()

	clone.Add(logger.Labels{"src": "clone"})

	original.Add(logger.Labels{"src": "original"})

	if clone["src"] != "clone" || original["src"] != "original" {
		t.Error("Clone not independent.")
	}

	if val, ok := clone["key0"]; !ok || val != "val0" {
		t.Error("Clone missing data.")
	}
}

func TestLabelsMethodChaining(t *testing.T) {
	labels := logger.Labels{"key0": "val0", "key1": "val1"}
	labels.Del("key1").Add(logger.L{"key2": "val2"})

	if _, ok := labels["key1"]; ok {
		t.Error("Not deleted.")
	}

	if val, ok := labels["key2"]; !ok || val != "val2" {
		t.Error("Not added.")
	}
}
