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
	"time"
)

const (
	ACTORS_PREFIX      = "actors/"
	MOVIES_PREFIX      = "movies/"
	BASE_URL           = "http://150.165.15.91:8001/"
	ACTORS_DATA_PATH   = "../data/actors.txt"
	MAX_ACTORS_TO_READ = 10000
)

var ranking = make([]Ator, 10)

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

	for i := 0; i < len(ranking); i++ {
		ranking[i] = Ator{}
	}

	cont := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		actorID := scanner.Text()
		actorID = actorID[1 : len(actorID)-1]

		handleActor(actorID)

		cont++
		fmt.Printf("atores processados: %d / %d\n", cont, MAX_ACTORS_TO_READ)
		if cont == MAX_ACTORS_TO_READ {
			break
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

	elapsed := time.Since(start)
	fmt.Println("\n--------------------------------------------------")
	fmt.Printf("Total execution time in millis: %d", elapsed.Milliseconds())
}

func insereRanking(ator Ator) {
	i := sort.Search(len(ranking), func(i int) bool { return ranking[i].AverageRating <= ator.AverageRating })
	if i < len(ranking) {
		ranking = append(ranking[:i], append([]Ator{ator}, ranking[i:]...)...)
		ranking = ranking[:len(ranking)-1]
	}
}

func handleActor(actorID string) {
	url := fmt.Sprintf("%s%s%s", BASE_URL, ACTORS_PREFIX, actorID)
	response := doGet(url)

	var ator Ator
	err := json.Unmarshal([]byte(response), &ator)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}
	getActorAVGRating(&ator)
	insereRanking(ator)
}

func getActorAVGRating(ator *Ator) {
	var sum float32
	for _, movie := range ator.Movies {
		sum += getMovieRating(movie)
	}

	ator.AverageRating = sum / float32(len(ator.Movies))
}

func getMovieRating(movieID string) float32 {
	url := fmt.Sprintf("%s%s%s", BASE_URL, MOVIES_PREFIX, movieID)
	response := doGet(url)

	var movie Movie
	err := json.Unmarshal(response, &movie)
	if err != nil {
		log.Fatal("Erro ao decodificar JSON:", err)
	}

	return movie.AverageRating
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
