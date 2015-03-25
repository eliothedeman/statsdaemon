package metric

import (
	"log"
	"testing"
)

func TestChanToSlice(t *testing.T) {
	c := make(chan float64, 100)
	for i := 0; i < cap(c); i++ {
		c <- float64(i)
	}

	arr := chanToSlice(c)
	log.Println(arr)
}
