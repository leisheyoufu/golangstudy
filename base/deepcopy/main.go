package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"
	"time"

	"strconv"
)

// Test ...
type Test struct {
	Prop1 int
	Prop2 string
}

func (t *Test) DeepCopyInto(out *Test) {
	*out = *t
	out.Prop1 = t.Prop1
	out.Prop2 = t.Prop2
}

// Clone deep-copies a to b
func Clone(a, b interface{}) {

	buff := new(bytes.Buffer)
	enc := gob.NewEncoder(buff)
	dec := gob.NewDecoder(buff)
	enc.Encode(a)
	dec.Decode(b)
}

// DeepCopy deepcopies a to b using json marshaling
func DeepCopy(a, b interface{}) {
	byt, _ := json.Marshal(a)
	json.Unmarshal(byt, b)
}

func main() {
	i := 0
	tClone := time.Duration(0)
	tCopy := time.Duration(0)
	tCopyStruct := time.Duration(0)
	end := 3000
	for {
		if i == end {
			break
		}

		r := Test{Prop1: i, Prop2: strconv.Itoa(i)}
		var rNew Test
		t0 := time.Now()
		Clone(r, &rNew)
		t2 := time.Now().Sub(t0)
		tClone += t2

		r2 := Test{Prop1: i, Prop2: strconv.Itoa(i)}
		var rNew2 Test
		t0 = time.Now()
		DeepCopy(&r2, &rNew2)
		t2 = time.Now().Sub(t0)
		tCopy += t2

		t0 = time.Now()
		r3 := Test{Prop1: i, Prop2: strconv.Itoa(i)}
		var rNew3 Test
		r3.DeepCopyInto(&rNew3)
		i++
		t3 := time.Now().Sub(t0)
		tCopyStruct += t3
	}
	log.Printf("Total items %+v, Clone avg. %+v, DeepCopy avg. %+v, DeepCopyInto avg. %+v Total Difference %+v\n", i, tClone/3000, tCopy/3000, tCopyStruct/3000, (tClone - tCopy))
}
