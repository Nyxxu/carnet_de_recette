# syntax=docker/dockerfile:1.7

# =============================================================================
# Stage 1 — Builder
# Compile un binaire Go statique avec toutes les ressources (recettes, images,
# templates) embarquées via //go:embed.
# =============================================================================
FROM golang:1.26-alpine AS builder

WORKDIR /src

# Cache des dépendances : on copie d'abord les fichiers de modules pour que
# l'étape `go mod download` soit cachée tant que go.mod/go.sum ne changent pas.
COPY go.mod go.sum ./
RUN go mod download

# Copie du reste des sources.
COPY . .

# Build statique :
# - CGO_ENABLED=0 : pas de dépendances libc → tourne sur scratch / distroless static.
# - GOOS=linux : cible explicite (Scaleway tourne sur Linux).
# - -ldflags="-s -w" : strip des symboles, binaire plus léger.
# - -trimpath : retire les chemins absolus pour des builds reproductibles.
RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/carnet-recette .


# =============================================================================
# Stage 2 — Runtime
# Image distroless static : ~2 MB, pas de shell, pas d'utilitaires. Surface
# d'attaque minimale, idéale pour un binaire Go statique.
# =============================================================================
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=builder /out/carnet-recette /carnet-recette

# Scaleway Serverless Containers injecte la variable PORT à l'exécution.
# main.go lit os.Getenv("PORT") et défaut à 8080.
EXPOSE 8080

ENTRYPOINT ["/carnet-recette"]
