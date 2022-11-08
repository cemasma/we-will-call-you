package game

import (
	"errors"
	"math/rand"
	"sync"
	"time"
	"we-will-call-you/communication"
)

const CardCountPerRound = 12
const CardCountPerPlayer = 6

const (
	PlayerTypeInterviewer = 1
	PlayerTypeInterviewee = 2
)

type Game struct {
	Id                 string
	InterviewerPlayer  *communication.Connection
	IntervieweePlayer  *communication.Connection
	interviewerCards   []string
	intervieweeCards   []string
	job                string
	turn               int
	repository         Repository
	playedInterviewees map[string]bool
	roomConnections    map[*communication.Connection]bool
	mu                 sync.Mutex
}

func NewGame(roomId string, repository Repository) (game *Game) {

	g := new(Game)

	g.Id = roomId
	g.turn = CardCountPerRound
	g.repository = repository
	g.roomConnections = make(map[*communication.Connection]bool)
	g.playedInterviewees = make(map[string]bool)

	return g
}

func (g *Game) StartGame() (err error) {
	if len(g.roomConnections) < 2 {
		return errors.New("game is not reached to minimum required player capacity")
	}

	g.InterviewerPlayer, g.IntervieweePlayer = getRandomPlayers(g.roomConnections)
	g.turn = CardCountPerRound
	g.playedInterviewees = make(map[string]bool)

	return nil
}

func (g *Game) addPlayer(c *communication.Connection) bool {
	if len(g.roomConnections) == 0 {
		c.IsAdmin = true
	}

	g.roomConnections[c] = true

	return c.IsAdmin
}

func (g *Game) removePlayer(c *communication.Connection) {
	delete(g.roomConnections, c)
}

func (g *Game) getPlayerList() map[*communication.Connection]bool {
	return g.roomConnections
}

func (g *Game) PickRandomWordsAndJob() error {
	if !g.IsGameStarted() {
		return errors.New("game should be started to pick random words and job")
	}

	if !g.IsAlreadyStarted() {
		g.pickRandomJob()
	}

	err := g.pickRandomWords()

	return err
}

func (g *Game) pickRandomWords() error {
	words, err := g.repository.GetRandomWords(CardCountPerRound)

	if err != nil {
		return err
	}

	g.interviewerCards, g.intervieweeCards = words[:CardCountPerPlayer], words[CardCountPerPlayer:]

	return nil
}

func (g *Game) pickRandomJob() {
	job := g.repository.GetRandomJob()

	g.job = job
}

func getRandomPlayers(roomConnections map[*communication.Connection]bool) (interviewerPlayer, intervieweePlayer *communication.Connection) {
	playerCount := len(roomConnections)

	rand.Seed(time.Now().UnixNano())
	interviewerIndex := rand.Intn(playerCount)
	intervieweeIndex := rand.Intn(playerCount)

	if intervieweeIndex == interviewerIndex && intervieweeIndex < playerCount-1 {
		intervieweeIndex += 1
	} else if intervieweeIndex == interviewerIndex {
		intervieweeIndex -= 1
	}

	index := 0
	for key := range roomConnections {
		if index == interviewerIndex {
			interviewerPlayer = key
		}

		if index == intervieweeIndex {
			intervieweePlayer = key
		}

		if interviewerPlayer != nil && intervieweePlayer != nil {
			break
		}

		index++
	}

	return
}

func (g *Game) GetJob() string {
	return g.job
}

func (g *Game) GetInterviewerCards() []string {
	return g.interviewerCards
}

func (g *Game) GetIntervieweeCards() []string {
	return g.intervieweeCards
}

func (g *Game) IsConnectionBelongsToActivePlayer(playerId string) bool {
	return playerId == g.InterviewerPlayer.Id || playerId == g.IntervieweePlayer.Id
}

func (g *Game) isWordBelongsToPlayer(playerId string, word string) int {
	searchFunc := func(value string, arr []string) int {
		for i, v := range arr {
			if v == word {
				return i
			}
		}

		return -1
	}

	if playerId == g.InterviewerPlayer.Id {
		return searchFunc(word, g.interviewerCards)
	}

	if playerId == g.IntervieweePlayer.Id {
		return searchFunc(word, g.intervieweeCards)
	}

	return -1
}

func (g *Game) PlayWord(playerId string, word string) (bool, int) {
	indexOfWord := g.isWordBelongsToPlayer(playerId, word)

	if g.turn == 0 {
		return false, 0
	}

	if indexOfWord == -1 {
		return false, 0
	}

	removeFunc := func(arr []string, index int) []string {
		arr[index] = arr[len(arr)-1]

		return arr[:len(arr)-1]
	}

	playerType := 0
	if playerId == g.InterviewerPlayer.Id {
		playerType = PlayerTypeInterviewer
		g.interviewerCards = removeFunc(g.interviewerCards, indexOfWord)
	} else if playerId == g.IntervieweePlayer.Id {
		playerType = PlayerTypeInterviewee
		g.intervieweeCards = removeFunc(g.intervieweeCards, indexOfWord)
	}

	g.turn -= 1

	return true, playerType
}

func (g *Game) NextInterviewee() bool {
	if g == nil || g.InterviewerPlayer == nil {
		return false
	}

	g.playedInterviewees[g.IntervieweePlayer.Id] = true

	remainPlayers := make([]*communication.Connection, 0)
	for player := range g.roomConnections {
		if !g.playedInterviewees[player.Id] && player.Id != g.InterviewerPlayer.Id {
			remainPlayers = append(remainPlayers, player)
		}
	}

	if len(remainPlayers) == 0 {
		return false
	}

	rand.Seed(time.Now().UnixNano())

	nextInterviewee := remainPlayers[rand.Intn(len(remainPlayers))]

	g.IntervieweePlayer = nextInterviewee
	g.turn = 12

	return true
}

func (g *Game) IsGameStarted() bool {
	return g != nil && g.InterviewerPlayer != nil && g.IntervieweePlayer != nil
}

func (g *Game) IsAlreadyStarted() bool {
	return len(g.playedInterviewees) > 0
}
