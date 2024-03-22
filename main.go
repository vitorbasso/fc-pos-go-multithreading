package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	cep = *flag.String("cep", "01153000", "CEP to search")
)

func init() {
	flag.Parse()
	if cep == "" {
		log.Panic("CEP is required")
	}
}

func main() {
	viaCepChan := make(chan viacepResponse)
	brasilApiChan := make(chan brasilapiResponse)

	go getViaCep(viaCepChan, cep)
	go getBrasilApi(brasilApiChan, cep)

	select {
	case viaCep := <-viaCepChan:
		print(viaCep, "viaCep")
	case brasilApi := <-brasilApiChan:
		print(brasilApi, "brasilApi")
	case <-time.After(1 * time.Second):
		log.Println("Timeout")
	}
}

func print(response any, from string) {
	res, err := json.Marshal(response)
	if err != nil {
		log.Panic(fmt.Errorf("print error: %w", err).Error())
	}
	log.Printf("Response through %s:\n %s", from, string(res))
}

type viacepResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

func getViaCep(ch chan<- viacepResponse, cep string) {
	client := &http.Client{Timeout: 1001 * time.Millisecond}
	res, err := client.Get("https://viacep.com.br/ws/" + cep + "/json")
	if err != nil {
		log.Printf("getViaCep error: %+v", err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Printf("getViaCep error: %+v", fmt.Errorf("status code is not 200"))
		return
	}
	var address viacepResponse
	if err := json.NewDecoder(res.Body).Decode(&address); err != nil {
		log.Printf("getViaCep error: %+v", err)
		return
	}
	ch <- address
}

type brasilapiResponse struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

func getBrasilApi(ch chan<- brasilapiResponse, cep string) {
	client := &http.Client{Timeout: 1001 * time.Millisecond}
	res, err := client.Get("https://brasilapi.com.br/api/cep/v1/" + cep)
	if err != nil {
		log.Printf("getBrasilApi error: %+v", err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Printf("getBrasilApi error: %+v", fmt.Errorf("status code is not 200"))
		return
	}
	var address brasilapiResponse
	if err := json.NewDecoder(res.Body).Decode(&address); err != nil {
		log.Printf("getBrasilApi error: %+v", err)
		return
	}
	ch <- address
}
