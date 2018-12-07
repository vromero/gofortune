package strfile

import (
	"math/rand"
	"time"

	"github.com/vromero/gofortune/pkg"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func Shuffle(input []pkg.DataPos) {
	for i := range input {
		j := rand.Intn(i + 1)
		input[i], input[j] = input[j], input[i]
	}
}
