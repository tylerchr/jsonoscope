package jsonoscope

import (
	"crypto/sha1"
	"encoding/json"
	"hash"
	"io"
	"strconv"
	"strings"
	"sync"
)

const (
	Null Token = 1 + iota
	Number
	Boolean
	String
	Array
	Object
)

type (
	recurser struct {
		dec     *json.Decoder // the source of JSON tokens
		visitor Visitor       // visitors to invoke as we traverse the JSON tree
		hasher  hash.Hash     // a hash function used to calculate signatures

		// sigpool pools buffers used for generating an object signature
		sigpool *sync.Pool

		// we can precompute signatures for constant JSON values as soon
		// as we have a hash function, instead of recalculating them each
		// time they occur in the JSON data
		trueSig  []byte
		falseSig []byte
		nullSig  []byte
	}

	// Token indicates the type of the value at a given JSON path. It is always
	// one of: Null, Number, Boolean, String, Array, or Object.
	Token int
)

// String returns a string in the set {Null, Number, Boolean, String,
// Array, Object}, or "<unknown>" if the value of Token t is invalid.
func (t Token) String() string {
	switch t {
	case Null:
		return "Null"
	case Number:
		return "Number"
	case Boolean:
		return "Boolean"
	case String:
		return "String"
	case Array:
		return "Array"
	case Object:
		return "Object"
	default:
		return "<unknown>"
	}
}

// Recurse performs a depth-first search over a JSON tree and invokes the methods
// of the provided Visitor for each value in the tree.
func Recurse(r io.Reader, vis Visitor) error {

	h := sha1.New()

	rec := &recurser{
		dec:     json.NewDecoder(r),
		visitor: vis,
		hasher:  h,
		sigpool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, h.Size())
			},
		},
	}

	rec.precompute()

	rec.dec.UseNumber()

	_, err := rec.recurse()
	return err
}

// recurse recurses through the JSON from r.dec
func (r *recurser) recurse() ([]byte, error) {
	t, err := r.dec.Token()
	if err != nil {
		return nil, err
	}
	return r.recurseToken(".", t)
}

// precompute generates hashes for true, false, and null.
func (r *recurser) precompute() {

	r.hasher.Reset()
	r.hasher.Write([]byte("true"))
	r.trueSig = r.hasher.Sum(nil)

	r.hasher.Reset()
	r.hasher.Write([]byte("false"))
	r.falseSig = r.hasher.Sum(nil)

	r.hasher.Reset()
	r.hasher.Write([]byte("null"))
	r.nullSig = r.hasher.Sum(nil)

}

// recurseToken generates the hash of any JSON token.
func (r *recurser) recurseToken(path string, t json.Token) (sig []byte, err error) {

	switch tt := t.(type) {

	case json.Delim: // for the four JSON delimiters [ ] { }
		if tt == '[' {
			r.visitor.Enter(path, Array)
			sig, err = r.recurseArray(path)
			r.visitor.Exit(path, Array, sig)
		} else if tt == '{' {
			r.visitor.Enter(path, Object)
			sig, err = r.recurseObject(path)
			r.visitor.Exit(path, Object, sig)
		}

	case bool: // for JSON booleans
		r.visitor.Enter(path, Boolean)
		if tt {
			sig = r.trueSig[:]
		} else {
			sig = r.falseSig[:]
		}
		r.visitor.Exit(path, Boolean, sig)

	case json.Number: // for JSON numbers
		r.visitor.Enter(path, Number)
		r.hasher.Reset()
		r.hasher.Write([]byte(tt))
		sig = r.hasher.Sum(nil)
		r.visitor.Exit(path, Number, sig)

	case string: // for JSON string literals
		r.visitor.Enter(path, String)
		r.hasher.Reset()
		r.hasher.Write([]byte(`"` + tt + `"`))
		sig = r.hasher.Sum(nil)
		r.visitor.Exit(path, String, sig)

	case nil: // for JSON null
		r.visitor.Enter(path, Null)
		sig = r.nullSig[:]
		r.visitor.Exit(path, Null, sig)

	}

	return

}

// recurseArray generates the hash of an array.
func (r *recurser) recurseArray(path string) ([]byte, error) {

	hh := sha1.New()
	var idx int64

	for r.dec.More() {

		t, err := r.dec.Token()

		h, err := r.recurseToken(path+"["+strconv.FormatInt(idx, 10)+"]", t)
		if err != nil {
			return nil, err
		}

		hh.Write(h[:])

		idx++

	}

	r.dec.Token() // consume final ']'

	return hh.Sum(nil), nil

}

// recurseObject generates the hash of an object.
func (r *recurser) recurseObject(path string) ([]byte, error) {

	// obtain a buffer to hold the object signature
	sig := r.sigpool.Get().([]byte)

	// reset the signature
	for i := range sig {
		sig[i] = 0
	}

	for r.dec.More() {

		// read the key from the object
		t, err := r.dec.Token()
		if err != nil {
			return nil, err
		}

		key := t.(string) // we know it is valid since r.dec.Token didn't error

		// figure out the subpath for this key
		var subpath string
		if strings.HasSuffix(path, ".") {
			subpath = path + key
		} else {
			subpath = path + "." + key
		}

		// recursively read the key's value
		t, err = r.dec.Token()
		if err != nil {
			return nil, err
		}

		valueSignature, err := r.recurseToken(subpath, t)
		if err != nil {
			return nil, err
		}

		// generate a signature for this KV pair
		r.hasher.Reset()
		r.hasher.Write([]byte(key))
		r.hasher.Write(valueSignature)

		// xor this KV hash into our final KV hash
		for i, v := range r.hasher.Sum(nil) {
			sig[i] = sig[i] ^ v
		}

	}

	// consume the final '}'
	r.dec.Token()

	return sig, nil

}
