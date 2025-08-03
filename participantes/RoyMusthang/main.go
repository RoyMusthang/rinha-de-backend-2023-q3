package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"

	"github.com/go-playground/validator/v10"
)

type PersonRequest struct {
	Apelido    string    `json:"apelido" validate:"required,max=32"`
	Nome       string    `json:"nome" validate:"required,max=100"`
	Nascimento string    `json:"nascimento" validate:"required"`
	Stack      *[]string `json:"stack"`
}

var (
	validate     = validator.New()
	storage      = make(map[string]PersonRequest)
	storageMutex sync.Mutex
	dateRegex    = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

func createPerson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var p PersonRequest
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	// Validação padrão com struct tags
	if err := validate.Struct(p); err != nil {
		http.Error(w, fmt.Sprintf("Erro de validação: %v", err), http.StatusBadRequest)
		return
	}

	// Validação do nascimento no formato AAAA-MM-DD
	if !dateRegex.MatchString(p.Nascimento) {
		http.Error(w, "Formato de nascimento inválido. Use AAAA-MM-DD", http.StatusBadRequest)
		return
	}

	// Validação dos elementos da stack (se fornecido)
	if p.Stack != nil {
		for _, item := range *p.Stack {
			if item == "" || len(item) > 32 {
				http.Error(w, "Cada item da stack deve ser não vazio e ter até 32 caracteres", http.StatusBadRequest)
				return
			}
		}
	}

	// Verifica se o apelido já foi cadastrado
	storageMutex.Lock()
	defer storageMutex.Unlock()
	if _, exists := storage[p.Apelido]; exists {
		http.Error(w, "Apelido já existente", http.StatusConflict)
		return
	}

	// Armazena a pessoa (simulação)
	storage[p.Apelido] = p

	// Retorna status 201 Created
	w.WriteHeader(http.StatusCreated)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /pessoas", createPerson)
	// TODO: GET /pessoas[:id]
	// TODO: GET /pessoas?=t[:termo da busca]
	// TODO: GET /contagem-pessoas

	err := http.ListenAndServe(":9999", mux)
	if err != nil {
		log.Fatalf("erro em iniciar o server: %v", err)
	}
}
