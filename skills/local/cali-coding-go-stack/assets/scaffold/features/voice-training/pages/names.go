package pages

import (
	"math/rand"
	"time"
)

type ClientProfile struct {
	Name      string
	VoiceTone string // "male" or "female"
}

var clientProfiles = []ClientProfile{
	// Female voices
	{Name: "Ana", VoiceTone: "female"},
	{Name: "Maria", VoiceTone: "female"},
	{Name: "Julia", VoiceTone: "female"},
	{Name: "Camila", VoiceTone: "female"},
	{Name: "Fernanda", VoiceTone: "female"},
	{Name: "Leticia", VoiceTone: "female"},
	{Name: "Patricia", VoiceTone: "female"},
	{Name: "Carolina", VoiceTone: "female"},
	{Name: "Beatriz", VoiceTone: "female"},
	{Name: "Amanda", VoiceTone: "female"},
	{Name: "Larissa", VoiceTone: "female"},
	{Name: "Mariana", VoiceTone: "female"},
	{Name: "Juliana", VoiceTone: "female"},
	{Name: "Isabela", VoiceTone: "female"},
	{Name: "Tatiane", VoiceTone: "female"},
	{Name: "Renata", VoiceTone: "female"},
	{Name: "Vanessa", VoiceTone: "female"},
	{Name: "Cristina", VoiceTone: "female"},
	{Name: "Daniela", VoiceTone: "female"},
	{Name: "Raquel", VoiceTone: "female"},
	{Name: "Gabriela", VoiceTone: "female"},
	{Name: "Isadora", VoiceTone: "female"},
	{Name: "Luana", VoiceTone: "female"},
	{Name: "Priscila", VoiceTone: "female"},
	{Name: "Bianca", VoiceTone: "female"},

	// Male voices
	{Name: "Lucas", VoiceTone: "male"},
	{Name: "Pedro", VoiceTone: "male"},
	{Name: "Gabriel", VoiceTone: "male"},
	{Name: "Rafael", VoiceTone: "male"},
	{Name: "Bruno", VoiceTone: "male"},
	{Name: "Rodrigo", VoiceTone: "male"},
	{Name: "Felipe", VoiceTone: "male"},
	{Name: "Gustavo", VoiceTone: "male"},
	{Name: "Diego", VoiceTone: "male"},
	{Name: "Thiago", VoiceTone: "male"},
	{Name: "Andre", VoiceTone: "male"},
	{Name: "Carlos", VoiceTone: "male"},
	{Name: "Marcelo", VoiceTone: "male"},
	{Name: "Ricardo", VoiceTone: "male"},
	{Name: "Fernando", VoiceTone: "male"},
	{Name: "Leonardo", VoiceTone: "male"},
	{Name: "Eduardo", VoiceTone: "male"},
	{Name: "Matheus", VoiceTone: "male"},
	{Name: "Bruno", VoiceTone: "male"},
	{Name: "Alex", VoiceTone: "male"},
	{Name: "Daniel", VoiceTone: "male"},
	{Name: "Henrique", VoiceTone: "male"},
	{Name: "Vinicius", VoiceTone: "male"},
	{Name: "Paulo", VoiceTone: "male"},
	{Name: "Julio", VoiceTone: "male"},
}

func RandomClientProfile() ClientProfile {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return clientProfiles[r.Intn(len(clientProfiles))]
}

func RandomBrazilianName() string {
	return RandomClientProfile().Name
}
