package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Ator struct {
	Id     string   `json:"id"`
	Name   string   `json:"name"`
	Movies []string `json:"movies"`
}

func main() {
	file, err := os.Open("actors.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		actorID := scanner.Text()
		actorID = actorID[1 : len(actorID)-1]

		go doGet(actorID)
		time.Sleep(1 * time.Second)
	}
}

func doGet(actorID string) {
	url := fmt.Sprintf("http://150.165.15.91:8001/actors/%s", actorID)

	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("Erro ao fazer a requisição: %s\n", err)
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Erro ao ler o corpo da resposta: %s\n", err)
		return
	}

	var ator Ator
	err = json.Unmarshal([]byte(body), &ator)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}

	fmt.Println(ator.Movies)
}
