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

		go handleActor(actorID)
		time.Sleep(1 * time.Second) //Lembrar de retirar
	}
}

func handleActor(actorID string) {
	url := fmt.Sprintf("http://150.165.15.91:8001/actors/%s", actorID)
	response := doGet(url)
	body := readBody(response)

	var ator Ator
	err := json.Unmarshal([]byte(body), &ator)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}

	var wg sync.WaitGroup
	moviesMap := make(map[string]float32)
	for _, movie := range ator.Movies {
		wg.Add(1)
		go getMovieRating(&wg, movie, moviesMap)
	}

	wg.Wait()
	var sum float32
	for _, rating := range moviesMap {
		sum += rating
	}
	averageRating := sum / float32(len(moviesMap))
	fmt.Printf("Average of %s is: %.2f\n", actorID, averageRating)
}

func getMovieRating(wg *sync.WaitGroup, movieID string, moviesMap map[string]float32) {
	defer wg.Done()

	url := fmt.Sprintf("http://150.165.15.91:8001/movies/%s", movieID)
	response := doGet(url)
	body := readBody(response)

	var movie Movie
	err := json.Unmarshal([]byte(body), &movie)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}

	moviesMap[movieID] = movie.AverageRating
}

func doGet(url string) *http.Response {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal("Erro ao fazer a requisição:", err.Error())
	}

	return response
}

func readBody(response *http.Response) []byte {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Erro ao ler o corpo da resposta:", err.Error())
	}
	return body
}
