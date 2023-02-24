package in_memory

import (
	"errors"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

const EN = "en"
const TR = "tr"

var Languages = map[string]bool{
	EN: true,
	TR: true,
}

type InMemoryRepository struct {
	jobs  []string
	words []string
}

func NewInMemoryRepository(language string) *InMemoryRepository {
	if Languages[language] != true {
		language = EN
	}

	inMemoryRepository := new(InMemoryRepository)

	jobsBytes, err := os.ReadFile("./repository/in_memory/" + language + "/jobs.txt")

	if err != nil {
		log.Fatal(err)
	}

	wordsBytes, err := os.ReadFile("./repository/in_memory/" + language + "/words.txt")

	inMemoryRepository.jobs = strings.Split(string(jobsBytes), "\n")
	inMemoryRepository.words = strings.Split(string(wordsBytes), "\n")

	return inMemoryRepository
}

func (inMemoryRepository InMemoryRepository) GetRandomWords(n int) ([]string, error) {
	wordCount := len(inMemoryRepository.words)

	if n > wordCount {
		return nil, errors.New("count cannot be higher than total word count")
	}

	if n < 1 {
		return nil, errors.New("count should be a positive number")
	}

	if n == wordCount {
		return inMemoryRepository.words, nil
	}

	rand.Seed(time.Now().UnixNano())

	usedIndexes := make(map[int]bool)
	words := make([]string, n)

	for i := 0; i < n; i++ {
		randomIndex := rand.Intn(wordCount)

		if ok, _ := usedIndexes[randomIndex]; ok {
			i--
			randomIndex = rand.Intn(wordCount)
		}

		usedIndexes[randomIndex] = true
		words[i] = inMemoryRepository.words[randomIndex]
	}

	return words, nil
}

func (inMemoryRepository InMemoryRepository) GetRandomJob() string {
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(inMemoryRepository.jobs))

	return inMemoryRepository.jobs[randomIndex]
}
