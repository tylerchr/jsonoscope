package jsonoscope

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestEqual(t *testing.T) {

	cases := []struct {
		First, Second []byte
		Equal         bool
	}{
		// formatting does not matter
		{
			First:  []byte(`[1,2,3]`),
			Second: []byte(`[ 1, 2, 3 ]`),
			Equal:  true,
		},

		// order matters in arrays
		{
			First:  []byte(`[1, 2, 3]`),
			Second: []byte(`[1, 3, 2]`),
			Equal:  false,
		},

		// order does not matter in objects
		{
			First:  []byte(`{ "Planet": "Earth", "Index": 3 }`),
			Second: []byte(`{ "Index": 3, "Planet": "Earth" }`),
			Equal:  true,
		},
	}

	for i, c := range cases {

		eq, err := Equal(bytes.NewReader(c.First), bytes.NewReader(c.Second))
		if err != nil {
			panic(err)
		}

		if eq != c.Equal {
			t.Errorf("[case %d] Unexpected equality: expected %t but got %t\n", i, c.Equal, eq)
		}

	}

}

func BenchmarkEqual(b *testing.B) {

	for i := 0; i < b.N; i++ {
		eq, _ := Equal(bytes.NewReader(SampleJSON), bytes.NewReader(SampleJSON))
		if !eq {
			b.Fatalf("not equal")
		}
	}

}

func BenchmarkDeepEqual(b *testing.B) {

	for i := 0; i < b.N; i++ {

		var json1, json2 map[string]interface{}

		_ = json.NewDecoder(bytes.NewReader(SampleJSON)).Decode(&json1)
		_ = json.NewDecoder(bytes.NewReader(SampleJSON)).Decode(&json2)

		eq := reflect.DeepEqual(json1, json2)
		if !eq {
			b.Fatalf("not equal")
		}
	}
}
