package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	// Abre o arquivo
	file, err := os.Open("actors.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Cria um scanner para ler o arquivo linha por linha
	scanner := bufio.NewScanner(file)

	// Itera sobre as linhas do arquivo
	for scanner.Scan() {
		actorID := scanner.Text()
		actorID = actorID[1 : len(actorID)-1]

		url := fmt.Sprintf("http://150.165.15.91:8001/actors/%s", actorID)

		// Faz a requisição GET
		response, err := http.Get(url)
		if err != nil {
			fmt.Printf("Erro ao fazer a requisição: %s\n", err)
			return
		}
		defer response.Body.Close()

		// Lê o corpo da resposta
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("Erro ao ler o corpo da resposta: %s\n", err)
			return
		}

		// Exibe a resposta
		fmt.Println(string(body))
	}
}
