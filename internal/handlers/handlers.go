package handlers

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/Nyxxu/carnet-de-recette/internal/recettes"
)

// Serveur regroupe les dépendances HTTP : catalogue de recettes et templates
// rendus côté serveur.
type Serveur struct {
	catalogue *recettes.Catalogue
	templates map[string]*template.Template
	imagesFS  fs.FS
}

// Nouveau construit un Serveur prêt à être branché sur un http.Mux.
//
// templates doit contenir au moins les clés "index" et "recette". Chaque
// template doit définir un block "layout" qui sera exécuté pour rendre la page.
// imagesFS doit pointer directement sur le contenu du dossier images/.
func Nouveau(catalogue *recettes.Catalogue, templates map[string]*template.Template, imagesFS fs.FS) *Serveur {
	return &Serveur{
		catalogue: catalogue,
		templates: templates,
		imagesFS:  imagesFS,
	}
}

// Routes retourne le handler racine avec toutes les routes enregistrées.
func (s *Serveur) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /recette/{slug}", s.handleRecette)
	mux.HandleFunc("GET /aleatoire", s.handleAleatoire)
	mux.HandleFunc("GET /loto", s.handleLoto)
	mux.Handle("GET /images/", http.StripPrefix("/images/", http.FileServer(http.FS(s.imagesFS))))
	return mux
}

type vueIndex struct {
	Titre               string
	TotalCatalogue      int
	Critères            recettes.Criteres
	Resultats           []recettes.Recette
	Groupes             []recettes.GroupeCategorie
	NombreResultats     int
	FiltreActif         bool
	CategoriesDisponibles []categorieOption
	TagsDisponibles     []string
	TagsSelectionnesMap map[string]bool
}

type categorieOption struct {
	Valeur  recettes.Categorie
	Libelle string
	Active  bool
}

func (s *Serveur) handleIndex(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	crit := recettes.Criteres{
		Texte:     q.Get("q"),
		Categorie: recettes.Categorie(q.Get("categorie")),
		Tags:      q["tag"],
	}
	// Une catégorie inconnue est traitée comme "aucun filtre".
	if !crit.Categorie.Valide() {
		crit.Categorie = ""
	}

	resultats := s.catalogue.Rechercher(crit)
	groupes := recettes.GrouperParCategorie(resultats)

	categories := []categorieOption{
		{Valeur: "", Libelle: "Tout", Active: crit.Categorie == ""},
	}
	for _, c := range recettes.OrdreCategories {
		categories = append(categories, categorieOption{
			Valeur:  c,
			Libelle: c.Libelle(),
			Active:  crit.Categorie == c,
		})
	}

	tagsSelectionnes := make(map[string]bool, len(crit.Tags))
	for _, t := range crit.Tags {
		tagsSelectionnes[t] = true
	}

	donnees := vueIndex{
		Titre:                 "Carnet de recettes",
		TotalCatalogue:        s.catalogue.Nombre(),
		Critères:              crit,
		Resultats:             resultats,
		Groupes:               groupes,
		NombreResultats:       len(resultats),
		FiltreActif:           !crit.Vide(),
		CategoriesDisponibles: categories,
		TagsDisponibles:       s.catalogue.TousLesTags(),
		TagsSelectionnesMap:   tagsSelectionnes,
	}
	s.rendre(w, "index", donnees)
}

type vueRecette struct {
	Titre   string
	Recette recettes.Recette
}

func (s *Serveur) handleAleatoire(w http.ResponseWriter, r *http.Request) {
	// Le tirage au sort est limité aux plats : c'est ce qu'on cherche quand on
	// manque d'idée pour le repas, pas une glace ou une entrée.
	rec, ok := s.catalogue.Aleatoire(recettes.CategoriePlat)
	if !ok {
		http.Error(w, "aucun plat dans le catalogue", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, "/recette/"+rec.Slug, http.StatusSeeOther)
}

func (s *Serveur) handleRecette(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	rec, ok := s.catalogue.ParSlug(slug)
	if !ok {
		http.NotFound(w, r)
		return
	}
	s.rendre(w, "recette", vueRecette{
		Titre:   rec.Titre,
		Recette: rec,
	})
}

func (s *Serveur) rendre(w http.ResponseWriter, nom string, donnees any) {
	tmpl, ok := s.templates[nom]
	if !ok {
		http.Error(w, fmt.Sprintf("template inconnu : %q", nom), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", donnees); err != nil {
		http.Error(w, "Erreur de rendu : "+err.Error(), http.StatusInternalServerError)
	}
}
