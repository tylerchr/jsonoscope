package jsonoscope

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestRecurse(t *testing.T) {

	cases := []struct {
		JSON       []byte
		Keys       int
		Tokens     map[string]Token
		Signatures map[string][]byte
	}{
		{
			JSON: []byte(`null`),
			Keys: 1,
			Tokens: map[string]Token{
				".": Null,
			},
			Signatures: map[string][]byte{
				".": MustDecodeString("2be88ca4242c76e8253ac62474851065032d6833"),
			},
		},
		{
			JSON: []byte(`1`),
			Keys: 1,
			Tokens: map[string]Token{
				".": Number,
			},
			Signatures: map[string][]byte{
				".": MustDecodeString("356a192b7913b04c54574d18c28d46e6395428ab"),
			},
		},
		{
			JSON: []byte(`true`),
			Keys: 1,
			Tokens: map[string]Token{
				".": Boolean,
			},
			Signatures: map[string][]byte{
				".": MustDecodeString("5ffe533b830f08a0326348a9160afafc8ada44db"),
			},
		},
		{
			JSON: []byte(`false`),
			Keys: 1,
			Tokens: map[string]Token{
				".": Boolean,
			},
			Signatures: map[string][]byte{
				".": MustDecodeString("7cb6efb98ba5972a9b5090dc2e517fe14d12cb04"),
			},
		},
		{
			JSON: []byte(`"hello"`),
			Keys: 1,
			Tokens: map[string]Token{
				".": String,
			},
			Signatures: map[string][]byte{
				".": MustDecodeString("a1f2fbfe2c4ad81749cd0380b735295d06f9d0c4"),
			},
		},

		// For array input, verify that children are visited and that order matters.
		{
			JSON: []byte(`[1, 2]`),
			Keys: 3,
			Tokens: map[string]Token{
				".":    Array,
				".[0]": Number,
				".[1]": Number,
			},
			Signatures: map[string][]byte{
				".":    MustDecodeString("58c6912831df431a52af3cd818caa352f60d8db0"),
				".[0]": MustDecodeString("356a192b7913b04c54574d18c28d46e6395428ab"),
				".[1]": MustDecodeString("da4b9237bacccdf19c0760cab7aec4a8359010b0"),
			},
		},
		{
			JSON: []byte(`[2, 1]`),
			Keys: 3,
			Tokens: map[string]Token{
				".":    Array,
				".[0]": Number,
				".[1]": Number,
			},
			Signatures: map[string][]byte{
				".":    MustDecodeString("6cd222ce3e5160b83760f62ab187087474d885f6"),
				".[0]": MustDecodeString("da4b9237bacccdf19c0760cab7aec4a8359010b0"),
				".[1]": MustDecodeString("356a192b7913b04c54574d18c28d46e6395428ab"),
			},
		},

		// For object input, verify that children are visited and that order does not matter.
		{
			JSON: []byte(`{ "Planet": "Earth", "Index": 3 }`),
			Keys: 3,
			Tokens: map[string]Token{
				".":       Object,
				".Planet": String,
				".Index":  Number,
			},
			Signatures: map[string][]byte{
				".":       MustDecodeString("a0fc0dbd5dd267db2f72cde45d83f5941f04abac"),
				".Planet": MustDecodeString("81c5bb3f2088f30b131671bddfa4db414eafdcfa"),
				".Index":  MustDecodeString("77de68daecd823babbb58edb1c8e14d7106e83bb"),
			},
		},
		{
			JSON: []byte(`{ "Index": 3, "Planet": "Earth" }`),
			Keys: 3,
			Tokens: map[string]Token{
				".":       Object,
				".Planet": String,
				".Index":  Number,
			},
			Signatures: map[string][]byte{
				".":       MustDecodeString("a0fc0dbd5dd267db2f72cde45d83f5941f04abac"),
				".Planet": MustDecodeString("81c5bb3f2088f30b131671bddfa4db414eafdcfa"),
				".Index":  MustDecodeString("77de68daecd823babbb58edb1c8e14d7106e83bb"),
			},
		},
	}

	for i, c := range cases {
		var keys int
		Recurse(bytes.NewReader(c.JSON), CustomVisitor{
			OnEnter: func(path string, token Token) {
				keys++
			},
			OnExit: func(path string, token Token, sig []byte) {

				// If an expected Token for this path was provided, verify it.
				if expectedToken, ok := c.Tokens[path]; ok && token != expectedToken {
					t.Errorf("[case %d] Unexpected token for key '%s': expected %s but got %s\n", i, path, expectedToken, token)
				}

				// If an expected Signature for this path was provided, verify it.
				if expectedSig, ok := c.Signatures[path]; ok && !bytes.Equal(sig, expectedSig) {
					t.Errorf("[case %d] Unexpected signature for key '%s': expected %x but got %x\n", i, path, expectedSig, sig)
				}

			},
		})

		if c.Keys > 0 && c.Keys != keys {
			t.Errorf("[case %d] Unexpected key count: expected %d but got %d\n", i, c.Keys, keys)
		}
	}

}

// MustDecodeString wraps hex.DecodeString but panics if a failure occurs. It
// simplifies test code where we're hardcoding a hex value we know won't fail.
func MustDecodeString(s string) []byte {
	data, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}
