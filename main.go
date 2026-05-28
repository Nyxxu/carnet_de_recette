package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Nyxxu/carnet-de-recette/internal/handlers"
	"github.com/Nyxxu/carnet-de-recette/internal/recettes"
)

// version est injectée à la compilation via -ldflags="-X main.version=<date>".
// Vaut "dev" lors des runs locaux (go run .).
var version = "dev"

//go:embed recettes/*.yaml
var fichiersRecettes embed.FS

//go:embed images
var fichiersImages embed.FS

//go:embed templates/*.html
var fichiersTemplates embed.FS

func main() {
	catalogue, err := recettes.Charger(fichiersRecettes, "recettes")
	if err != nil {
		log.Fatalf("chargement des recettes : %v", err)
	}
	log.Printf("catalogue chargé : %d recettes", catalogue.Nombre())

	templates, err := chargerTemplates(fichiersTemplates)
	if err != nil {
		log.Fatalf("chargement des templates : %v", err)
	}

	imagesFS, err := fs.Sub(fichiersImages, "images")
	if err != nil {
		log.Fatalf("sous-système images : %v", err)
	}

	serveur := handlers.Nouveau(catalogue, templates, imagesFS)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	srv := &http.Server{
		Addr:              addr,
		Handler:           serveur.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("écoute sur http://localhost%s", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("serveur HTTP : %v", err)
	}
}

// chargerTemplates parse les templates de pages. Chaque page est parsée avec
// le layout commun pour former un set de templates indépendant.
func chargerTemplates(systeme fs.FS) (map[string]*template.Template, error) {
	funcs := template.FuncMap{
		"add":        func(a, b int) int { return a + b },
		"normaliser": recettes.Normaliser,
		// version() dans les templates retourne la version courante du binaire.
		"version": func() string { return version },
		// thumb("lasagnes.jpg") → "lasagnes-thumb.jpg". Utilisé sur la page
		// d'accueil pour charger la version 600x600 au lieu du hero 1200x675.
		"thumb": func(filename string) string {
			if filename == "" {
				return ""
			}
			ext := filepath.Ext(filename)
			return strings.TrimSuffix(filename, ext) + "-thumb" + ext
		},
	}

	pages := map[string]string{
		"index":   "templates/index.html",
		"recette": "templates/recette.html",
	}

	out := make(map[string]*template.Template, len(pages))
	for nom, chemin := range pages {
		t, err := template.New(nom).Funcs(funcs).ParseFS(systeme, "templates/layout.html", chemin)
		if err != nil {
			return nil, err
		}
		out[nom] = t
	}
	return out, nil
}
