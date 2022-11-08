package game

import (
	"encoding/json"
	"log"
)

const (
	actionNotifyAdmin   = 0
	actionGameStarted   = 1
	actionJobPrepared   = 2
	actionCardsPrepared = 3
	actionCardPlayed    = 4
	actionPlayerList    = 5
)

type message struct {
	Action      int         `json:"action"`
	Interviewer string      `json:"interviewer,omitempty"`
	Interviewee string      `json:"interviewee,omitempty"`
	Job         string      `json:"job,omitempty"`
	Cards       interface{} `json:"cards,omitempty"`
	Word        string      `json:"word,omitempty"`
	PlayerType  int         `json:"playerType,omitempty"`
}

func getMessageMap() map[int]func(args ...interface{}) []byte {
	functionMap := map[int]func(args ...interface{}) []byte{
		actionNotifyAdmin:   notifyAdminMessage,
		actionGameStarted:   gameStartedMessage,
		actionJobPrepared:   jobPreparedMessage,
		actionCardsPrepared: cardsPreparedMessage,
		actionCardPlayed:    cardPlayedMessage,
	}

	return functionMap
}

func notifyAdminMessage(_ ...interface{}) []byte {
	m := message{Action: actionNotifyAdmin}

	return marshal(m)
}

func gameStartedMessage(args ...interface{}) []byte {
	m := message{
		Action:      actionGameStarted,
		Interviewer: args[0].(string),
		Interviewee: args[1].(string),
	}

	return marshal(m)
}

func jobPreparedMessage(args ...interface{}) []byte {
	m := message{
		Action: actionJobPrepared,
		Job:    args[0].(string),
	}

	return marshal(m)
}

func cardsPreparedMessage(args ...interface{}) []byte {
	m := message{
		Action: actionCardsPrepared,
		Cards:  args[0],
	}

	return marshal(m)
}

func cardPlayedMessage(args ...interface{}) []byte {
	m := message{
		Action:     actionCardPlayed,
		Word:       args[0].(string),
		PlayerType: args[1].(int),
	}

	return marshal(m)
}

func marshal(m message) []byte {
	byt, err := json.Marshal(m)

	if err != nil {
		log.Fatal(err)
	}

	return byt
}
