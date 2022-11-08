package game

import (
	"encoding/json"
	"log"
	"sort"
	"strings"
	"we-will-call-you/communication"
	"we-will-call-you/repository/in_memory"
)

var C *Controller

type Controller struct {
	Games              map[string]*Game
	Repositories       map[string]Repository
	DomainBroadcast    chan communication.Message
	messageFunctionMap map[int]func(args ...interface{}) []byte
}

func (c *Controller) ProcessCommand(message communication.Message) {
	g := c.Games[message.GetRoomId()]

	if g == nil {
		return
	}

	strMessage := string(message.GetData())
	if strMessage == "/start" && message.GetConnection().IsAdmin {
		err := g.StartGame()

		if err != nil {
			log.Println(err)
			return
		}

		byt := c.messageFunctionMap[actionGameStarted](g.InterviewerPlayer.Name, g.IntervieweePlayer.Name)

		m := communication.NewMessage(nil, byt, message.GetRoomId())
		c.sendMessage(m)
	}

	if strMessage == "/dealcards" && message.GetConnection().IsAdmin {
		err := g.PickRandomWordsAndJob()

		if err != nil {
			return
		}

		jobByt := c.messageFunctionMap[actionJobPrepared](g.GetJob())

		m := communication.NewMessage(nil, jobByt, message.GetRoomId())
		c.sendMessage(m)

		interviewerCardsByt := c.messageFunctionMap[actionCardsPrepared](g.GetInterviewerCards())
		intervieweeCardsByt := c.messageFunctionMap[actionCardsPrepared](g.GetIntervieweeCards())

		interviewerMessage := communication.NewMessage(g.InterviewerPlayer, interviewerCardsByt, m.GetRoomId())
		intervieweeMessage := communication.NewMessage(g.IntervieweePlayer, intervieweeCardsByt, m.GetRoomId())
		c.sendMessage(interviewerMessage)
		c.sendMessage(intervieweeMessage)
	}

	if strings.Contains(strMessage, "/card:") && g != nil && g.IsConnectionBelongsToActivePlayer(message.GetConnection().Id) {
		splittedWord := strings.Split(strMessage, "/card:")

		word := ""

		if len(splittedWord) == 2 {
			word = splittedWord[1]
		}

		if ok, playerType := g.PlayWord(message.GetConnection().Id, word); ok {
			cardPlayedByte := c.messageFunctionMap[actionCardPlayed](word, playerType)

			m := communication.NewMessage(nil, cardPlayedByte, message.GetRoomId())
			c.sendMessage(m)
		}
	}

	if strMessage == "/nextInterviewee" {
		ok := g.NextInterviewee()

		if ok {
			byt := c.messageFunctionMap[actionGameStarted](g.InterviewerPlayer.Name, g.IntervieweePlayer.Name)

			m := communication.NewMessage(nil, byt, message.GetRoomId())
			c.sendMessage(m)
		}
	}
}

func (c *Controller) AddConnection(roomId string, connection *communication.Connection) {
	g := c.Games[roomId]

	if g == nil {
		var repository Repository
		if val, ok := c.Repositories[connection.Language]; ok {
			repository = val
		} else {
			repository = c.Repositories[in_memory.EN]
		}

		g = NewGame(roomId, repository)
		c.Games[roomId] = g
	}

	g.mu.Lock()
	isAdmin := g.addPlayer(connection)

	if isAdmin {
		byt := c.messageFunctionMap[actionNotifyAdmin]()

		m := communication.NewMessage(connection, byt, roomId)
		c.sendMessage(m)
	}

	c.updatePlayerList(roomId)
	g.mu.Unlock()
}

func (c *Controller) sendMessage(m communication.Message) {
	go func() {
		c.DomainBroadcast <- m
	}()
}

func (c *Controller) RemoveConnection(roomId string, connection *communication.Connection) {
	g := c.Games[roomId]

	if g == nil {
		return
	}

	g.mu.Lock()
	g.removePlayer(connection)

	playerList := g.getPlayerList()

	if len(playerList) == 0 {
		delete(c.Games, roomId)
		return
	} else if connection.IsAdmin == true {
		for key, val := range playerList {
			if val == true {
				key.IsAdmin = true
				byt := c.messageFunctionMap[actionNotifyAdmin]()

				m := communication.NewMessage(key, byt, roomId)
				c.sendMessage(m)
				break
			}
		}
	}

	c.updatePlayerList(roomId)
	g.mu.Unlock()
}

func (c *Controller) Initialize(domainBroadcast chan communication.Message) {
	c.DomainBroadcast = domainBroadcast
}

func (c *Controller) updatePlayerList(roomId string) {
	g := c.Games[roomId]

	if g == nil {
		return
	}

	connections := g.getPlayerList()
	playerList := make([]string, 0)
	for k, v := range connections {
		if v == true {
			playerList = append(playerList, k.Name)
		}
	}
	sort.Strings(playerList)

	type playerListMessage struct {
		Action     int      `json:"action"`
		PlayerList []string `json:"playerList"`
	}

	playerListMsg := playerListMessage{
		Action:     actionPlayerList,
		PlayerList: playerList,
	}

	byt, err := json.Marshal(playerListMsg)

	if err == nil {
		m := communication.NewMessage(nil, byt, roomId)
		c.sendMessage(m)
	}
}

func (c *Controller) DeleteGame(roomId string) {
	delete(c.Games, roomId)
}

func init() {
	repositories := map[string]Repository{
		in_memory.EN: in_memory.NewInMemoryRepository(in_memory.EN),
		in_memory.TR: in_memory.NewInMemoryRepository(in_memory.TR),
	}

	C = &Controller{
		Games:              make(map[string]*Game),
		Repositories:       repositories,
		DomainBroadcast:    nil,
		messageFunctionMap: getMessageMap(),
	}
}
