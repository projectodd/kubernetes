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

package main

type TokenizeError string

func (e TokenizeError) Error() string {
	return string(e)
}

func tokenize(s string) ([]string, error) {
	inQuote, escape := false, false
	var curQuote rune
	tokens := []string{}
	curToken := []rune{}

	for _, r := range s {
		appendEscape, appendRune, appendToken := false, true, false

		switch r {
		case '\\':
			appendRune = escape
			escape = !escape
		case '\'', '"':
			appendRune = false
			switch {
			case escape:
				appendEscape = !(inQuote && curQuote == r)
				appendRune = true

			case inQuote:
				inQuote = curQuote != r

			default:
				inQuote = true
				curQuote = r
			}
			escape = false
		case ' ', '\t':
			appendRune = escape
			if inQuote {
				appendEscape = escape
				appendRune = true
			} else {
				appendToken = !escape && len(curToken) > 0
			}
			escape = false
		default:
			appendEscape = escape
			escape = false
		}

		if appendEscape {
			curToken = append(curToken, '\\')
		}
		if appendRune {
			curToken = append(curToken, r)
		}
		if appendToken {
			tokens = append(tokens, string(curToken))
			curToken = []rune{}
		}
	}

	if inQuote {
		return tokens, TokenizeError("Unexpected end of input; did you forget to close a quote?")
	}
	if len(curToken) > 0 {
		tokens = append(tokens, string(curToken))
	}

	return tokens, nil
}
