package services

import "math/rand"

var brazilianNames = []string{
	"Ana",
	"Beatriz",
	"Camila",
	"Daniel",
	"Eduarda",
	"Felipe",
	"Gabriela",
	"Henrique",
	"Isabela",
	"João",
}

func RandomBrazilianName() string {
	return brazilianNames[rand.Intn(len(brazilianNames))]
}
