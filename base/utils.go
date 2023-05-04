package base

import (
	"fmt"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"math"
	"math/rand"
	"reflect"
	"time"
	"unicode"
)

func ShuffleArray(arr interface{}) {
	rand.Seed(time.Now().UnixNano())
	rv := reflect.ValueOf(arr)
	swap := reflect.Swapper(arr)
	length := rv.Len()
	for i := length - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		swap(i, j)
	}
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func Float64ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

// Removing accents from strings
func ExampleRemove() {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ := transform.String(t, "résumé")
	fmt.Println(s)
	// Output:
	// resume
}
