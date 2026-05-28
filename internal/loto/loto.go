// Package loto produit des tirages aléatoires du Loto français.
//
// Règles : 5 numéros distincts dans [1, 49] + 1 numéro chance dans [1, 10].
// L'aléa vient de crypto/rand pour garantir une vraie entropie (pas un PRNG
// seedé déterministe). C'est un POC indépendant du carnet de recettes.
package loto

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"sort"
)

const (
	NumerosMax   = 49 // numéros tirés dans 1..49
	NumerosTires = 5
	ChanceMax    = 10 // chance dans 1..10
)

type Tirage struct {
	Numeros []int `json:"numeros"` // 5 distincts, triés croissant
	Chance  int   `json:"chance"`  // 1..10
}

func Genere() (Tirage, error) {
	nums, err := tirerSansRemise(NumerosMax, NumerosTires)
	if err != nil {
		return Tirage{}, fmt.Errorf("tirage numéros : %w", err)
	}
	sort.Ints(nums)

	chance, err := intN(ChanceMax)
	if err != nil {
		return Tirage{}, fmt.Errorf("tirage chance : %w", err)
	}
	return Tirage{Numeros: nums, Chance: chance + 1}, nil
}

func tirerSansRemise(n, k int) ([]int, error) {
	if k > n {
		return nil, fmt.Errorf("k=%d > n=%d", k, n)
	}
	pool := make([]int, n)
	for i := range pool {
		pool[i] = i + 1
	}
	for i := 0; i < k; i++ {
		j, err := intN(n - i)
		if err != nil {
			return nil, err
		}
		pool[i], pool[i+j] = pool[i+j], pool[i]
	}
	return pool[:k], nil
}

func intN(n int) (int, error) {
	if n <= 0 {
		return 0, fmt.Errorf("intN : n doit être > 0, reçu %d", n)
	}
	v, err := crand.Int(crand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0, err
	}
	return int(v.Int64()), nil
}
