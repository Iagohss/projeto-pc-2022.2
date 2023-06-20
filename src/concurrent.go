package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Ator struct {
	Id     string   `json:"id"`
	Name   string   `json:"name"`
	Movies []string `json:"movies"`
}

type Movie struct {
	Id              string   `json:"id"`
	Title           string   `json:"title"`
	AverageRating   float32  `json:"averagerating"`
	NumberOfVotes   int      `json:"numberOfVotes"`
	StartYear       int      `json:"startYear"`
	LenghtInMinutes int      `json:"lenghtInMinutes"`
	Genres          []string `json:"genres"`
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

		go doGetAtor(actorID)
		time.Sleep(1 * time.Second) //Lembrar de retirar
	}
}

func doGetAtor(actorID string) {
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

	var wg sync.WaitGroup
	moviesMap := make(map[string]float32)
	for _, movie := range ator.Movies {
		wg.Add(1)
		go doGetMovie(&wg, movie, moviesMap)
	}

	wg.Wait()
	var sum float32
	for _, rating := range moviesMap {
		fmt.Printf("%s : %.2f\n", actorID, rating)
		sum += rating
	}
	fmt.Printf("Average of %s is: %.2f\n", actorID, sum/float32(len(moviesMap)))
}

func doGetMovie(wg *sync.WaitGroup, movieID string, moviesMap map[string]float32) {
	defer wg.Done()

	url := fmt.Sprintf("http://150.165.15.91:8001/movies/%s", movieID)

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

	var movie Movie
	err = json.Unmarshal([]byte(body), &movie)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}

	moviesMap[movieID] = movie.AverageRating
}
