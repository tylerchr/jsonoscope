package jsonoscope

import (
	"bytes"
	"io"
)

// Equal indicates whether the two Readers contain semantically
// identical JSON data by comparing the signatures of their root
// JSON objects.
func Equal(r1, r2 io.Reader) (bool, error) {

	var sig1, sig2 []byte

	_ = Recurse(r1, CustomVisitor{
		OnExit: func(path string, token Token, sig []byte) {
			if path == "." {
				sig1 = sig
			}
		},
	})

	_ = Recurse(r2, CustomVisitor{
		OnExit: func(path string, token Token, sig []byte) {
			if path == "." {
				sig2 = sig
			}
		},
	})

	return bytes.Equal(sig1, sig2), nil

}
