package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	twitchApi "github.com/Onestay/go-new-twitch"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/twitch"
)

var client *twitchApi.Client
var followedCallback string
var events []FollowEvent
var url string
var wsConnection *websocket.Conn

func main() {
	url = "whispering-meadow-13437.herokuapp.com"
	client = twitchApi.NewClient("93uv9e2fs5bp8j65wyzdbqu1h7ulth")

	gothic.Store = sessions.NewCookieStore([]byte("secret"))
	followedCallback = "https://" + url + "/followed"

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	twitchProvider := twitch.New("93uv9e2fs5bp8j65wyzdbqu1h7ulth", "3uf68z2zwti1tt72ica3vk90behzw2", "https://"+url+"/callback")
	goth.UseProviders(twitchProvider)

	r.GET("/", renderIndex)
	r.POST("/login", login)
	r.GET("/callback", callback)
	r.POST("/followed", followedCallbackFn)
	r.GET("/followed", confirmFollowedListener)
	r.GET("/websocket", wshandler)

	r.Run() // listen and serve on 0.0.0.0:8080
}

func renderIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", []string{})
}

func login(c *gin.Context) {
	q := c.Request.URL.Query()
	q.Add("provider", "twitch")
	c.Request.URL.RawQuery = q.Encode()
	streamer := c.PostForm("streamer")
	session, _ := gothic.Store.Get(c.Request, "current_session")
	session.Values["streamer"] = streamer
	session.Save(c.Request, c.Writer)

	gothic.BeginAuthHandler(c.Writer, c.Request)
}

func callback(c *gin.Context) {
	q := c.Request.URL.Query()
	q.Add("provider", "twitch")
	c.Request.URL.RawQuery = q.Encode()
	_, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	session, _ := gothic.Store.Get(c.Request, "current_session")
	streamer := session.Values["streamer"]

	streamerData, err := client.GetUsersByLogin(streamer.(string))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	id := streamerData[0].ID
	subscriptionUrl := fmt.Sprintf("https://api.twitch.tv/helix/users/follows?first=1&to_id=%v", id)

	subscribe(c, followedCallback, subscriptionUrl)

	subscriptionUrl = fmt.Sprintf("https://api.twitch.tv/helix/users/follows?first=1&from_id=%v", id)

	subscribe(c, followedCallback, subscriptionUrl)

	c.HTML(http.StatusOK, "streamer.tmpl", gin.H{"streamer": streamer, "url": url})
}

func subscribe(c *gin.Context, callback string, subscriptionUrl string) {
	subscriptionRequest := SubscriptionRequest{callback, "subscribe", subscriptionUrl, 864000}

	requestJson, err := json.Marshal(subscriptionRequest)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	req, err := http.NewRequest("POST", "https://api.twitch.tv/helix/webhooks/hub", bytes.NewBuffer(requestJson))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Client-ID", "93uv9e2fs5bp8j65wyzdbqu1h7ulth")

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

}

func confirmFollowedListener(c *gin.Context) {
	challenge := c.Query("hub.challenge")
	c.String(http.StatusOK, challenge)
}

func followedCallbackFn(c *gin.Context) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Println("request Body:", string(body))
	var followedEvents FollowedEvents
	json.Unmarshal(body, &followedEvents)

	newEvents := followedEvents.Data
	for _, event := range newEvents {
		wsConnection.WriteMessage(websocket.TextMessage, []byte(event.FromName+" followed "+event.ToName))
	}
	fmt.Println(events)
}

func wshandler(c *gin.Context) {
	var err error
	wsConnection, err = wsupgrader.Upgrade(c.Writer, c.Request, nil)
	fmt.Println("Creating connection")
	if err != nil {
		panic(err)
	}
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type FollowedEvents struct {
	Data []FollowEvent `json:"data"`
}

type FollowEvent struct {
	FromId     string `json:"from_id"`
	FromName   string `json:"from_name"`
	ToId       string `json:"to_id"`
	ToName     string `json:"to_name"`
	FollowedAt string `json:"followed_at"`
}
type SubscriptionRequest struct {
	Callback     string `json:"hub.callback"`
	Mode         string `json:"hub.mode"`
	Topic        string `json:"hub.topic"`
	LeaseSeconds int    `json:"hub.lease_seconds"`
}
