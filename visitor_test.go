package jsonoscope

import (
	"bytes"
	"testing"
)

// SampleJSON is an example JSON object from RFC 7159.
var SampleJSON = []byte(`[
	{
		"precision": "zip",
		"Latitude": 37.7668,
		"Longitude": -122.3959,
		"Address": "",
		"City": "SAN FRANCISCO",
		"State": "CA",
		"Zip": "94107",
		"Country": "US"
	},
	{
		"precision": "zip",
		"Latitude": 37.371991,
		"Longitude": -122.026020,
		"Address": "",
		"City": "SUNNYVALE",
		"State": "CA",
		"Zip": "94085",
		"Country": "US"
	}
]`)

func TestCountingVisitor(t *testing.T) {

	r := bytes.NewReader(SampleJSON)

	cv := new(CountingVisitor)
	Recurse(r, cv)

	if nodes := cv.Nodes(); nodes != 19 {
		t.Errorf("unexpected node count: expected 19 but got %d\n", nodes)
	}

}
