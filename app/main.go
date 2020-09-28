package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/takaaki-s/ne-sdk-go/nextengine"
)

var (
	redirectURL  = "https://localhost:8080/callback"
	clientID     = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
)

type tokenRepository struct {
	c *gin.Context
}

func (t *tokenRepository) Token(ctx context.Context) (nextengine.Token, error) {
	sess := sessions.Default(t.c)
	token := nextengine.Token{}
	if tj := sess.Get("token"); tj != nil {
		err := json.Unmarshal(tj.([]byte), &token)
		return token, err
	}
	return token, nil
}

func (t *tokenRepository) Save(ctx context.Context, token nextengine.Token) error {
	if json, err := json.Marshal(token); err == nil {
		sess := sessions.Default(t.c)
		sess.Set("token", json)
		sess.Save()
	}
	return nil
}

func getNextEngineClient(c *gin.Context) *nextengine.Config {
	tr := &tokenRepository{
		c: c,
	}

	ncli := nextengine.NewClient(clientID, clientSecret, redirectURL, &http.Client{}, tr)
	return ncli
}

func main() {
	r := gin.Default()

	store := memstore.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{})
	})

	r.GET("/callback", callback)

	private := r.Group("/private")
	private.Use(authenticator)
	{
		private.GET("/company", company)
		private.GET("/user", loginuser)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServeTLS("./srv.cert", "./srv.key"); err != nil {
			if err != http.ErrServerClosed {
				log.Fatalf("error: %s\n", err)
			} else {
				log.Printf("listen: %s\n", err)
			}
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

func company(c *gin.Context) {
	ncli := getNextEngineClient(c)
	ctx := context.Background()
	res, err := ncli.APIExecuteNoRequiredLogin(ctx, "/api_app/company", nil)
	if err != nil {
		log.Printf("error: %#v", err)
		return
	}

	c.HTML(http.StatusOK, "company.tmpl", gin.H{"hoge": res.Data})
}

func loginuser(c *gin.Context) {
	ncli := getNextEngineClient(c)
	ctx := context.Background()
	res, err := ncli.APIExecute(ctx, "/api_v1_login_user/info", nil)
	if err != nil {
		log.Printf("error: %#v", err)
		return
	}

	c.HTML(http.StatusOK, "user.tmpl", gin.H{"user": res.Data[0]})
}

func callback(c *gin.Context) {
	ncli := getNextEngineClient(c)
	_, err := ncli.Authorize(c, c.Query("uid"), c.Query("state"))
	if err != nil {
		log.Printf("%v\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/")
		c.Abort()
		return
	}

	previousURI, _ := url.Parse("/")
	if c.Query("previous_uri") != "" {
		previousURI, _ = url.Parse(c.Query("previous_uri"))
	}

	c.Redirect(http.StatusTemporaryRedirect, previousURI.RequestURI())
}

func authenticator(c *gin.Context) {
	sess := sessions.Default(c)
	u := sess.Get("token")
	if u == nil {
		ncli := getNextEngineClient(c)
		extraParams := url.Values{}
		extraParams.Add("previous_uri", c.Request.RequestURI)
		uri := ncli.SignInURI(extraParams)
		c.Redirect(http.StatusTemporaryRedirect, uri)
		c.Abort()
		return
	}
	c.Next()
}
