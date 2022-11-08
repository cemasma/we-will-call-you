package game

type Repository interface {
	GetRandomWords(n int) ([]string, error)
	GetRandomJob() string
}
