package recettes

import (
	"fmt"
	"io/fs"
	"math/rand/v2"
	"path"
	"sort"
	"strings"

	"golang.org/x/text/unicode/norm"
	"gopkg.in/yaml.v3"
)

// Catalogue contient l'ensemble des recettes chargées en mémoire au démarrage.
// Il est immuable après chargement, donc safe pour un accès concurrent en lecture.
type Catalogue struct {
	parSlug    map[string]Recette
	ordreSlugs []string
}

// Charger lit tous les fichiers YAML présents dans le sous-dossier indiqué du
// système de fichiers donné (typiquement un embed.FS), les parse et retourne
// le catalogue indexé.
//
// Une erreur est retournée si un fichier est invalide (parse YAML, champ
// obligatoire manquant, slug ou id en collision, catégorie inconnue).
func Charger(systeme fs.FS, racine string) (*Catalogue, error) {
	entrees, err := fs.ReadDir(systeme, racine)
	if err != nil {
		return nil, fmt.Errorf("lecture du dossier %q : %w", racine, err)
	}

	cat := &Catalogue{
		parSlug: make(map[string]Recette),
	}

	for _, entree := range entrees {
		if entree.IsDir() {
			continue
		}
		nom := entree.Name()
		if !strings.HasSuffix(nom, ".yaml") && !strings.HasSuffix(nom, ".yml") {
			continue
		}

		chemin := path.Join(racine, nom)
		donnees, err := fs.ReadFile(systeme, chemin)
		if err != nil {
			return nil, fmt.Errorf("lecture de %q : %w", chemin, err)
		}

		var r Recette
		if err := yaml.Unmarshal(donnees, &r); err != nil {
			return nil, fmt.Errorf("parsing de %q : %w", chemin, err)
		}

		if err := valider(r, chemin); err != nil {
			return nil, err
		}

		if existant, ok := cat.parSlug[r.Slug]; ok {
			return nil, fmt.Errorf("slug en doublon %q : %q et %q", r.Slug, existant.Titre, r.Titre)
		}

		cat.parSlug[r.Slug] = r
		cat.ordreSlugs = append(cat.ordreSlugs, r.Slug)
	}

	// Tri global par titre normalisé pour avoir un ordre stable et lisible.
	sort.SliceStable(cat.ordreSlugs, func(i, j int) bool {
		ri := cat.parSlug[cat.ordreSlugs[i]]
		rj := cat.parSlug[cat.ordreSlugs[j]]
		return Normaliser(ri.Titre) < Normaliser(rj.Titre)
	})

	return cat, nil
}

// Toutes retourne toutes les recettes dans l'ordre alphabétique de leur titre.
func (c *Catalogue) Toutes() []Recette {
	out := make([]Recette, 0, len(c.ordreSlugs))
	for _, s := range c.ordreSlugs {
		out = append(out, c.parSlug[s])
	}
	return out
}

// ParSlug retourne la recette correspondant au slug donné. Le booléen vaut
// false si aucune recette ne correspond.
func (c *Catalogue) ParSlug(slug string) (Recette, bool) {
	r, ok := c.parSlug[slug]
	return r, ok
}

// GroupeCategorie représente une catégorie et ses recettes pour l'affichage.
type GroupeCategorie struct {
	Categorie Categorie
	Libelle   string
	Recettes  []Recette
}

// GroupeesParCategorie retourne les recettes regroupées par catégorie, dans
// l'ordre défini par OrdreCategories. Les catégories sans recette ne sont pas
// retournées.
func (c *Catalogue) GroupeesParCategorie() []GroupeCategorie {
	parCat := make(map[Categorie][]Recette)
	for _, s := range c.ordreSlugs {
		r := c.parSlug[s]
		parCat[r.Categorie] = append(parCat[r.Categorie], r)
	}

	var groupes []GroupeCategorie
	for _, cat := range OrdreCategories {
		if recs, ok := parCat[cat]; ok {
			groupes = append(groupes, GroupeCategorie{
				Categorie: cat,
				Libelle:   cat.Libelle(),
				Recettes:  recs,
			})
		}
	}
	return groupes
}

// Nombre retourne le nombre de recettes dans le catalogue.
func (c *Catalogue) Nombre() int {
	return len(c.ordreSlugs)
}

// Aleatoire retourne une recette tirée au sort, éventuellement restreinte à
// une catégorie. Passer une catégorie vide pour piocher dans tout le catalogue.
// Le booléen vaut false si aucune recette ne correspond.
func (c *Catalogue) Aleatoire(cat Categorie) (Recette, bool) {
	var candidats []string
	for _, slug := range c.ordreSlugs {
		if cat == "" || c.parSlug[slug].Categorie == cat {
			candidats = append(candidats, slug)
		}
	}
	if len(candidats) == 0 {
		return Recette{}, false
	}
	return c.parSlug[candidats[rand.IntN(len(candidats))]], true
}

func valider(r Recette, chemin string) error {
	if r.Slug == "" {
		return fmt.Errorf("%s : champ slug manquant", chemin)
	}
	if r.Titre == "" {
		return fmt.Errorf("%s : champ titre manquant", chemin)
	}
	if !r.Categorie.Valide() {
		return fmt.Errorf("%s : catégorie %q invalide (attendu : entree, plat, dessert, glace)", chemin, r.Categorie)
	}
	if len(r.Sections) == 0 {
		return fmt.Errorf("%s : au moins une section est requise", chemin)
	}
	return nil
}

// Normaliser retourne une version comparable d'une chaîne pour le tri et la
// recherche : minuscules + sans accents. Utilisée à la fois côté Go (recherche
// serveur, tri du catalogue) et exposée aux templates pour permettre un filtre
// client cohérent.
func Normaliser(s string) string {
	decompose := norm.NFD.String(s)
	var b strings.Builder
	b.Grow(len(decompose))
	for _, r := range decompose {
		// Ignore les marques diacritiques (catégorie Unicode "Mn" approximée
		// par la plage des combining marks).
		if r >= 0x0300 && r <= 0x036F {
			continue
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}
