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

package logger

type Labels map[string]string

type L = Labels

// Add adds new and overwrites existing keys.
func (l Labels) Add(labels Labels) Labels {
	for key, val := range labels {
		l[key] = val
	}

	return l
}

// Del deletes keys.
func (l Labels) Del(keys ...string) Labels {
	for _, key := range keys {
		delete(l, key)
	}

	return l
}

// Clone deep copy.
func (l Labels) Clone() Labels {
	clone := Labels{}
	for key, val := range l {
		clone[key] = val
	}

	return clone
}
