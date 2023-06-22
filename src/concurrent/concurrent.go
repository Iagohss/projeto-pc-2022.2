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
	ACTORS_PREFIX         = "actors/"
	MOVIES_PREFIX         = "movies/"
	BASE_URL              = "http://150.165.15.91:8001/"
	ACTORS_DATA_PATH      = "../data/actors.txt"
	MAX_ACOTRS_GOROUTINES = 200
	MAX_ACTORS_TO_READ    = 10000
)

var (
	GOROUTINES_LOCK   sync.Mutex
	ACTORS_GOROUTINES = 0
)

type Ator struct {
	Id            string   `json:"id"`
	Name          string   `json:"name"`
	Movies        []string `json:"movies"`
	AverageRating float32
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

	cont := 0
	scanner := bufio.NewScanner(file)
	var wgAVGs sync.WaitGroup
	for scanner.Scan() {
		actorID := scanner.Text()
		actorID = actorID[1 : len(actorID)-1]

		for {
			if canStartActorGoroutine() {
				wgAVGs.Add(1)
				go handleActor(&wgAVGs, actorID, results)
				break
			}
		}

		cont++
		if cont == MAX_ACTORS_TO_READ {
			break
		}
	}

	wgAVGs.Wait()
	close(results)
	<-done

	elapsed := time.Since(start)
	fmt.Println("\n--------------------------------------------------")
	fmt.Printf("Total execution time in millis: %d", elapsed.Milliseconds())
}

func ranking(results chan Ator, done chan<- int) {
	lenRanking := 10
	ranking := make([]Ator, lenRanking)
	for i := 0; i < lenRanking; i++ {
		ranking[i] = Ator{}
	}

	cont := 0
	for ator := range results {
		cont++
		fmt.Printf("atores processados: %d / %d\n", cont, MAX_ACTORS_TO_READ)
		i := sort.Search(len(ranking), func(i int) bool { return ranking[i].AverageRating <= ator.AverageRating })
		if i < len(ranking) {
			ranking = append(ranking[:i], append([]Ator{ator}, ranking[i:]...)...)
			ranking = ranking[:len(ranking)-1]
		}
	}

	for i, ator := range ranking {
		fmt.Printf("Top %d: {id: %s, name: %s, movies: %v, rating: %.2f}\n",
			i+1,
			ator.Id,
			ator.Name,
			ator.Movies,
			ator.AverageRating)
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
	getActorAVGRating(&ator)
	results <- ator
}

func getActorAVGRating(ator *Ator) {
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

	ator.AverageRating = sum / float32(len(ator.Movies))
	releaseActorGoroutine()
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
	var response *http.Response
	var err error
	for {
		response, err = http.Get(url)
		if err == nil {
			break
		}
	}

	var body []byte
	for {
		body, err = io.ReadAll(response.Body)
		if err == nil {
			break
		}
	}
	defer response.Body.Close()

	return body
}

func canStartActorGoroutine() bool {
	GOROUTINES_LOCK.Lock()
	defer GOROUTINES_LOCK.Unlock()
	return ACTORS_GOROUTINES < MAX_ACOTRS_GOROUTINES
}

func releaseActorGoroutine() {
	GOROUTINES_LOCK.Lock()
	defer GOROUTINES_LOCK.Unlock()
	ACTORS_GOROUTINES--
}
