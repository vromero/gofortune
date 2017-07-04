package strfile

import (
	"math/rand"
	"time"

	"github.com/gofortune/gofortune/lib"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func Shuffle(input []lib.DataPos) {
	for i := range input {
		j := rand.Intn(i + 1)
		input[i], input[j] = input[j], input[i]
	}
}
