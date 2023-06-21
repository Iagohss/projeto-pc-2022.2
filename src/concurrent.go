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
	file, err := os.Open("actors.txt")
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
		if cont == 100 {
			break
		}
	}

	wgAVGs.Wait()
	close(results)
	<-done

	elapsed := time.Since(start)
	fmt.Printf("Total execution time in millis: %d", elapsed)
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
	url := fmt.Sprintf("http://150.165.15.91:8001/actors/%s", actorID)
	response := doGet(url)
	body := readBody(response)

	var ator Ator
	err := json.Unmarshal([]byte(body), &ator)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}

	numMovies := len(ator.Movies)
	var wgMovies sync.WaitGroup
	ratings := make(chan float32, numMovies)
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

	ator.AVGrating = sum / float32(numMovies)
	results <- ator
}

func getMovieRating(wg *sync.WaitGroup, movieID string, ratings chan<- float32) {
	defer wg.Done()

	url := fmt.Sprintf("http://150.165.15.91:8001/movies/%s", movieID)
	response := doGet(url)
	body := readBody(response)

	var movie Movie
	err := json.Unmarshal(body, &movie)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}

	ratings <- movie.AverageRating
}

func doGet(url string) *http.Response {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal("Erro ao fazer a requisição:", err.Error())
	}
	return response
}

func readBody(response *http.Response) []byte {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Erro ao ler o corpo da resposta:", err.Error())
	}

	return body
}
