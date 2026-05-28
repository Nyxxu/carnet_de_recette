package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Nyxxu/carnet-de-recette/internal/loto"
)

// maxTiragesParRequete borne le nombre de tirages renvoyés en une requête.
// Volontairement bas (POC) : évite qu'un client mal intentionné force des
// milliers d'appels à crypto/rand par requête.
const maxTiragesParRequete = 5

// handleLoto : GET /loto[?n=N]
//   - n absent  → 1 tirage
//   - n ∈ [1,5] → N tirages
//   - sinon     → 400 Bad Request
//
// Réponse JSON :
//
//	{"tirages":[{"numeros":[3,17,22,38,45],"chance":7}, ...]}
func (s *Serveur) handleLoto(w http.ResponseWriter, r *http.Request) {
	n := 1
	if q := r.URL.Query().Get("n"); q != "" {
		v, err := strconv.Atoi(q)
		if err != nil || v < 1 || v > maxTiragesParRequete {
			http.Error(w,
				"paramètre n invalide : entier attendu entre 1 et 5",
				http.StatusBadRequest)
			return
		}
		n = v
	}

	tirages := make([]loto.Tirage, 0, n)
	for i := 0; i < n; i++ {
		t, err := loto.Genere()
		if err != nil {
			http.Error(w, "erreur de tirage : "+err.Error(), http.StatusInternalServerError)
			return
		}
		tirages = append(tirages, t)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"tirages": tirages,
	})
}
