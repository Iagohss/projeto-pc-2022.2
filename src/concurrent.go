package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

const (
	ACTORS_PREFIX    = "actors/"
	MOVIES_PREFIX    = "movies/"
	BASE_URL         = "http://150.165.15.91:8001/"
	ACTORS_DATA_PATH = "./data/actors.txt"
	NUMBER_OF_ACTORS = 100
)

type Ator struct {
	Id        string   `json:"id"`
	Name      string   `json:"name"`
	Movies    []string `json:"movies"`
	AVGrating float32
}

type Movie struct {
	Id            string  `json:"id"`
	AverageRating float32 `json:"averagerating"`
}

func main() {
	start := time.Now()

	file, err := os.Open(ACTORS_DATA_PATH)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	results := make(chan Ator, 100)
	done := make(chan int)
	go ranking(results, done)

	scanner := bufio.NewScanner(file)
	var wgAVGs sync.WaitGroup
	cont := 0
	for scanner.Scan() {
		actorID := scanner.Text()
		actorID = actorID[1 : len(actorID)-1]

		wgAVGs.Add(1)
		go handleActor(&wgAVGs, actorID, results)

		//limitando o número de atores analisados, caso contrário o socket crasha
		cont++
		if cont == NUMBER_OF_ACTORS {
			break
		}
	}

	wgAVGs.Wait()
	close(results)
	<-done

	elapsed := time.Since(start)
	fmt.Println("\n--------------------------------------------------")
	fmt.Printf("Total execution time in millis: %d", elapsed/1000000)
}

func ranking(results chan Ator, done chan<- int) {
	lenRanking := 10
	ranking := make([]Ator, lenRanking)
	for i := 0; i < lenRanking; i++ {
		ranking[i] = Ator{}
	}

	for ator := range results {
		i := sort.Search(len(ranking), func(i int) bool { return ranking[i].AVGrating <= ator.AVGrating })
		if i <= len(ranking) {
			ranking = append(ranking[:i], append([]Ator{ator}, ranking[i:]...)...)
			ranking = ranking[:len(ranking)-1]
		}
	}

	for i, ator := range ranking {
		fmt.Printf("Top %d: {id: %s, name: %s, movies: %v, rating: %.2f}\n", i+1, ator.Id, ator.Name, ator.Movies, ator.AVGrating)
	}

	done <- 0
}

func handleActor(wgAVGs *sync.WaitGroup, actorID string, results chan<- Ator) {
	defer wgAVGs.Done()
	url := fmt.Sprintf("%s%s%s", BASE_URL, ACTORS_PREFIX, actorID)
	response := doGet(url)

	var ator Ator
	err := json.Unmarshal([]byte(response), &ator)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}
	getActorAVGrating(ator, results)
}

func getActorAVGrating(ator Ator, results chan<- Ator) {
	var wgMovies sync.WaitGroup
	ratings := make(chan float32, len(ator.Movies))
	for _, movie := range ator.Movies {
		wgMovies.Add(1)
		go getMovieRating(&wgMovies, movie, ratings)
	}

	wgMovies.Wait()
	close(ratings)

	var sum float32
	for rating := range ratings {
		sum += rating
	}

	ator.AVGrating = sum / float32(len(ator.Movies))
	results <- ator
}

func getMovieRating(wgMovies *sync.WaitGroup, movieID string, ratings chan<- float32) {
	defer wgMovies.Done()
	url := fmt.Sprintf("%s%s%s", BASE_URL, MOVIES_PREFIX, movieID)
	response := doGet(url)

	var movie Movie
	err := json.Unmarshal(response, &movie)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}

	ratings <- movie.AverageRating
}

func doGet(url string) []byte {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal("Erro ao fazer a requisição:", err.Error())
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Erro ao ler o corpo da resposta:", err.Error())
	}

	return body
}
