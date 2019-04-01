package main

import (
	"net/http"

	twitchApi "github.com/Onestay/go-new-twitch"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/twitch"
)

var client *twitchApi.Client

func main() {
	client = twitchApi.NewClient("93uv9e2fs5bp8j65wyzdbqu1h7ulth")

	gothic.Store = sessions.NewCookieStore([]byte("secret"))

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	twitchProvider := twitch.New("93uv9e2fs5bp8j65wyzdbqu1h7ulth", "3uf68z2zwti1tt72ica3vk90behzw2", "http://localhost:8080/callback")
	goth.UseProviders(twitchProvider)

	r.GET("/", renderIndex)
	r.POST("/login", login)
	r.GET("/callback", callback)

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
	/*res, err := json.Marshal(user)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	jsonString := string(res)*/

	session, _ := gothic.Store.Get(c.Request, "current_session")
	streamer := session.Values["streamer"]

	c.HTML(http.StatusOK, "streamer.tmpl", gin.H{"streamer": streamer})
}
