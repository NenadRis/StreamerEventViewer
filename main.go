package main

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/twitch"
)

func main() {
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
	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	res, err := json.Marshal(user)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	jsonString := string(res)

	session, _ := gothic.Store.Get(c.Request, "current_session")
	streamer := session.Values["streamer"]

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(jsonString+" "+streamer.(string)))
}
