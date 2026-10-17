// fakeboop simula o serviço parceiro para desenvolvimento local.
// Mesma foto (bytes idênticos) => mesma identification; primeira vez retorna created=true.
// POST /identify só reconhece fotos já cadastradas (desconhecida => null).
package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"sync"
)

func main() {
	token := os.Getenv("BOOP_TOKEN")
	addr := os.Getenv("FAKEBOOP_ADDR")
	if addr == "" {
		addr = ":8000"
	}

	var mu sync.Mutex
	known := map[[32]byte]int64{}

	http.HandleFunc("POST /persons", func(w http.ResponseWriter, r *http.Request) {
		if token != "" && r.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		f, _, err := r.FormFile("image")
		if err != nil {
			http.Error(w, `{"error":"image required"}`, http.StatusBadRequest)
			return
		}
		data, _ := io.ReadAll(f)
		sum := sha256.Sum256(data)

		mu.Lock()
		id, ok := known[sum]
		if !ok {
			id = rand.Int64N(900_000_000) + 100_000_000
			known[sum] = id
		}
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": id, "identification": id, "precision": 1.0,
			"history": []int64{}, "created": !ok,
		})
	})

	authorized := func(w http.ResponseWriter, r *http.Request) bool {
		if token != "" && r.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, `{"detail":"Token inválido"}`, http.StatusUnauthorized)
			return false
		}
		return true
	}

	// POST /identify: com imagem devolve a pessoa conhecida ou null (não cadastra).
	// Sem imagem responde 422 (usado pelo portal para validar o token).
	http.HandleFunc("POST /identify", func(w http.ResponseWriter, r *http.Request) {
		if !authorized(w, r) {
			return
		}
		f, _, err := r.FormFile("image")
		if err != nil {
			http.Error(w, `{"detail":"image required"}`, http.StatusUnprocessableEntity)
			return
		}
		data, _ := io.ReadAll(f)
		sum := sha256.Sum256(data)

		mu.Lock()
		id, ok := known[sum]
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		if !ok {
			w.Write([]byte("null"))
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"identification": id, "precision": 1.0, "history": []int64{},
		})
	})

	http.HandleFunc("DELETE /persons/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !authorized(w, r) {
			return
		}
		mu.Lock()
		for k, v := range known {
			if r.PathValue("id") == fmt.Sprint(v) {
				delete(known, k)
			}
		}
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"deleted":true}`))
	})

	slog.Info("fakeboop listening", "addr", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}
