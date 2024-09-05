package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Estrutura para armazenar o resultado das APIs
type Address struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	ApiSource   string
}

// Função para fazer a requisição à API BrasilAPI
func fetchFromBrasilAPI(cep string, ch chan<- Address, ctx context.Context) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	req, _ := http.NewRequest("GET", url, nil)
	req = req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var address Address
	if err := json.NewDecoder(resp.Body).Decode(&address); err != nil {
		return
	}
	address.ApiSource = "BrasilAPI"
	select {
	case ch <- address:
	case <-ctx.Done():
		return
	}
}

// Função para fazer a requisição à API ViaCEP
func fetchFromViaCEP(cep string, ch chan<- Address, ctx context.Context) {
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)
	req, _ := http.NewRequest("GET", url, nil)
	req = req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var address Address
	if err := json.NewDecoder(resp.Body).Decode(&address); err != nil {
		return
	}
	address.ApiSource = "ViaCEP"
	select {
	case ch <- address:
	case <-ctx.Done():
		return
	}
}

func main() {
	cep := "01153000"
	ch := make(chan Address)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Iniciando as requisições concorrentes
	go fetchFromBrasilAPI(cep, ch, ctx)
	go fetchFromViaCEP(cep, ch, ctx)

	// Selecionando a resposta mais rápida ou timeout
	select {
	case address := <-ch:
		fmt.Printf("Resultado obtido da API %s: %+v\n", address.ApiSource, address)
	case <-ctx.Done():
		fmt.Println("Erro: Timeout - Nenhuma resposta foi obtida em 1 segundo.")
	}
}
