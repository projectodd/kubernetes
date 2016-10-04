// Copyright 2016 Red Hat, Inc, and individual contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubesh

import (
	"fmt"
	"testing"
)

func check(t *testing.T, exp []string, given string) {
	r, err := tokenize(given)
	if err != nil {
		t.Error(err)
	}

	success := true
	if len(exp) != len(r) {
		success = false
	} else {
		for i := range r {
			success = r[i] == exp[i]
		}
	}

	if !success {
		fmt.Printf("   Given: %#v\n", given)
		fmt.Printf("     Got: %#v\n", r)
		fmt.Printf("Expected: %#v\n", exp)
		t.Fail()
	}
}

func s(args ...string) []string {
	return args
}

func TestValidInput(t *testing.T) {
	check(t, s("abc"), "abc")
	check(t, s("ab", "c"), "ab c")
	check(t, s("ab c"), "ab\\ c")
	check(t, s("ab\rc"), "ab\rc")
	check(t, s("ab", "c"), "ab 'c'")
	check(t, s("ab", "'c"), "ab '\\'c'")
	check(t, s("ab", "c d"), "ab 'c d'")
	check(t, s("a", "b"), "a  \t b")
	check(t, s("ab", "\\'c"), `ab "\'c"`)
	check(t, s("ab", "\"c"), `ab "\"c"`)
	check(t, s("ab", "-x=c d"), "ab -x='c d'")
	check(t, s("ab", "-x=c d"), `ab -x="c d"`)
}

func TestBadInput(t *testing.T) {
	_, err := tokenize("a b \"c")
	if err == nil {
		t.Fail()
	}
}
