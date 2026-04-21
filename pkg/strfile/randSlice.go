package strfile

import (
	"math/rand"

	"github.com/vromero/gofortune/pkg"
)

func Shuffle(input []pkg.DataPos) {
	for i := range input {
		j := rand.Intn(i + 1)
		input[i], input[j] = input[j], input[i]
	}
}
