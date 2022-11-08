package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"we-will-call-you/communication"
	"we-will-call-you/room"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	go room.H.Run()

	r := gin.Default()
	r.Static("/cards", "./templates/cards")
	r.Static("/scripts", "./templates/scripts")
	r.LoadHTMLGlob("templates/views/*")

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"communication": "pong",
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.POST("/room", func(c *gin.Context) {
		roomId := room.NewRequestId().String()

		c.JSON(http.StatusOK, gin.H{
			"roomId": roomId,
		})
	})

	r.GET("/room/:roomId", func(c *gin.Context) {
		c.HTML(http.StatusOK, "room.html", gin.H{
			"roomId":   c.Param("roomId"),
			"name":     c.Query("name"),
			"language": c.Query("language"),
		})
	})

	r.GET("/ws/:roomId", func(c *gin.Context) {
		roomId := c.Param("roomId")
		name := c.Query("name")
		language := c.Query("language")

		serveWs(c.Writer, c.Request, roomId, name, language)
	})
	r.Run()
}

func serveWs(w http.ResponseWriter, r *http.Request, roomId, name, language string) {
	ws, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	c := &communication.Connection{Id: room.NewRequestId().String(), Name: name, Send: make(chan []byte, 256), Ws: ws, Language: language}
	s := room.Subscription{Conn: c, Room: roomId}
	room.H.Register <- s
	go s.WritePump()
	go s.ReadPump()
}
