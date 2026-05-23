package recettes

// Categorie représente le type de plat. Les valeurs valides sont définies par
// les constantes ci-dessous.
type Categorie string

const (
	CategorieEntree  Categorie = "entree"
	CategoriePlat    Categorie = "plat"
	CategorieDessert Categorie = "dessert"
	CategorieGlace   Categorie = "glace"
)

// LibelleCategorie retourne le libellé affiché en français pour une catégorie.
func (c Categorie) Libelle() string {
	switch c {
	case CategorieEntree:
		return "Entrées"
	case CategoriePlat:
		return "Plats"
	case CategorieDessert:
		return "Desserts"
	case CategorieGlace:
		return "Glaces"
	default:
		return string(c)
	}
}

// Valide indique si la catégorie correspond à une des valeurs connues.
func (c Categorie) Valide() bool {
	switch c {
	case CategorieEntree, CategoriePlat, CategorieDessert, CategorieGlace:
		return true
	}
	return false
}

// OrdreCategories définit l'ordre d'affichage des catégories sur la page d'accueil.
var OrdreCategories = []Categorie{
	CategorieEntree,
	CategoriePlat,
	CategorieDessert,
	CategorieGlace,
}

// Ingredient représente un ingrédient d'une section de recette.
//
// Quantite vaut 0 quand la quantité n'est pas spécifiée (ex : "sel selon le goût").
// Details est un texte libre informatif (ex : "brique de 200 ml").
type Ingredient struct {
	Nom      string  `yaml:"nom"`
	Quantite float64 `yaml:"quantite"`
	Unite    string  `yaml:"unite"`
	Details  string  `yaml:"details"`
}

// Section regroupe un ensemble d'ingrédients et d'étapes pour une partie de la
// recette (ex : "Sauce bolognaise", "Béchamel"). Pour les recettes simples, on
// utilise une unique section sans titre.
type Section struct {
	Titre       string       `yaml:"titre"`
	Ingredients []Ingredient `yaml:"ingredients"`
	Etapes      []string     `yaml:"etapes"`
}

// TagCoupDeCoeur est le tag qui marque les recettes favorites — affichées avec
// un badge visuel particulier sur les cards et la page de détail.
const TagCoupDeCoeur = "coup-de-coeur"

// Recette représente une recette complète, telle que parsée depuis un fichier
// YAML dans le dossier recettes/.
type Recette struct {
	Titre               string    `yaml:"titre"`
	Slug                string    `yaml:"slug"`
	Categorie           Categorie `yaml:"categorie"`
	Portions            int       `yaml:"portions"`
	TempsPreparationMin int       `yaml:"temps_preparation_min"`
	TempsCuissonMin     int       `yaml:"temps_cuisson_min"`
	Contient            []string  `yaml:"contient"`
	Tags                []string  `yaml:"tags"`
	Image               string    `yaml:"image"`
	Sections            []Section `yaml:"sections"`
	Notes               string    `yaml:"notes"`
}

// TempsTotalMin retourne la somme des temps de préparation et de cuisson.
func (r Recette) TempsTotalMin() int {
	return r.TempsPreparationMin + r.TempsCuissonMin
}

// EstCoupDeCoeur indique si la recette porte le tag coup-de-coeur.
func (r Recette) EstCoupDeCoeur() bool {
	for _, t := range r.Tags {
		if t == TagCoupDeCoeur {
			return true
		}
	}
	return false
}
