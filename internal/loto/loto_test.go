package loto

import "testing"

// 1000 itérations pour avoir une confiance raisonnable que les invariants
// tiennent sur n'importe quel tirage (5 distincts, triés, dans [1,49],
// chance dans [1,10]).
func TestGenere_RespectLesRegles(t *testing.T) {
	for i := 0; i < 1000; i++ {
		ti, err := Genere()
		if err != nil {
			t.Fatalf("itération %d : Genere : %v", i, err)
		}
		if len(ti.Numeros) != NumerosTires {
			t.Fatalf("itération %d : attendu %d numéros, reçu %d (%v)",
				i, NumerosTires, len(ti.Numeros), ti.Numeros)
		}
		for j, n := range ti.Numeros {
			if n < 1 || n > NumerosMax {
				t.Errorf("itération %d : numéro hors plage : %d (tirage %v)",
					i, n, ti.Numeros)
			}
			if j > 0 && ti.Numeros[j-1] >= n {
				t.Errorf("itération %d : non trié ou doublon : %v", i, ti.Numeros)
			}
		}
		if ti.Chance < 1 || ti.Chance > ChanceMax {
			t.Errorf("itération %d : chance hors plage : %d", i, ti.Chance)
		}
	}
}

// Vérifie tirerSansRemise sur des paramètres pathologiques.
func TestTirerSansRemise_KSupNRetourneErreur(t *testing.T) {
	if _, err := tirerSansRemise(3, 5); err == nil {
		t.Fatal("attendu erreur quand k > n, reçu nil")
	}
}
