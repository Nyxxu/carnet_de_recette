package recettes

import (
	"sort"
	"strings"
)

// Criteres regroupe les filtres applicables à une recherche.
//
// Tous les champs sont optionnels :
//   - Texte vide = pas de filtre texte.
//   - Categorie vide = toutes les catégories.
//   - Tags vide = aucun filtre de tag (sinon, la recette doit contenir TOUS les tags listés).
type Criteres struct {
	Texte     string
	Categorie Categorie
	Tags      []string
}

// Vide indique si aucun critère n'est actif (équivaut à retourner toutes les recettes).
func (c Criteres) Vide() bool {
	return strings.TrimSpace(c.Texte) == "" && c.Categorie == "" && len(c.Tags) == 0
}

// Rechercher applique les critères au catalogue et retourne les recettes
// correspondantes, dans l'ordre alphabétique du catalogue.
//
// Le filtre texte est appliqué de façon insensible à la casse et aux accents,
// sur le titre, les noms d'ingrédients et les tags.
func (c *Catalogue) Rechercher(crit Criteres) []Recette {
	texte := Normaliser(strings.TrimSpace(crit.Texte))
	tagsRecherches := make([]string, 0, len(crit.Tags))
	for _, t := range crit.Tags {
		if t = strings.TrimSpace(t); t != "" {
			tagsRecherches = append(tagsRecherches, strings.ToLower(t))
		}
	}

	var resultats []Recette
	for _, slug := range c.ordreSlugs {
		r := c.parSlug[slug]

		if crit.Categorie != "" && r.Categorie != crit.Categorie {
			continue
		}
		if !contientTousLesTags(r.Tags, tagsRecherches) {
			continue
		}
		if texte != "" && !correspondAuTexte(r, texte) {
			continue
		}

		resultats = append(resultats, r)
	}
	return resultats
}

// TousLesTags retourne la liste triée des tags uniques présents dans le catalogue.
func (c *Catalogue) TousLesTags() []string {
	vus := make(map[string]struct{})
	for _, r := range c.parSlug {
		for _, t := range r.Tags {
			vus[t] = struct{}{}
		}
	}
	out := make([]string, 0, len(vus))
	for t := range vus {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool {
		return Normaliser(out[i]) < Normaliser(out[j])
	})
	return out
}

// Groupe les recettes par catégorie selon OrdreCategories. Les catégories sans
// recette ne sont pas retournées.
func GrouperParCategorie(recs []Recette) []GroupeCategorie {
	parCat := make(map[Categorie][]Recette)
	for _, r := range recs {
		parCat[r.Categorie] = append(parCat[r.Categorie], r)
	}
	var groupes []GroupeCategorie
	for _, cat := range OrdreCategories {
		if rs, ok := parCat[cat]; ok {
			groupes = append(groupes, GroupeCategorie{
				Categorie: cat,
				Libelle:   cat.Libelle(),
				Recettes:  rs,
			})
		}
	}
	return groupes
}

func correspondAuTexte(r Recette, texteNormalise string) bool {
	if strings.Contains(Normaliser(r.Titre), texteNormalise) {
		return true
	}
	for _, t := range r.Tags {
		if strings.Contains(Normaliser(t), texteNormalise) {
			return true
		}
	}
	for _, s := range r.Sections {
		for _, ing := range s.Ingredients {
			if strings.Contains(Normaliser(ing.Nom), texteNormalise) {
				return true
			}
		}
	}
	return false
}

func contientTousLesTags(tagsRecette, tagsRecherches []string) bool {
	if len(tagsRecherches) == 0 {
		return true
	}
	présents := make(map[string]struct{}, len(tagsRecette))
	for _, t := range tagsRecette {
		présents[strings.ToLower(t)] = struct{}{}
	}
	for _, t := range tagsRecherches {
		if _, ok := présents[t]; !ok {
			return false
		}
	}
	return true
}
