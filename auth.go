package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopkg.in/urfave/cli.v2"
)

const (
	iijAPIURL = "https://api.iijmio.jp/mobile/d/v1/authorization/"
	authURL   = "http://localhost:8080/auth"
)

type validResponse struct {
	AccessToken string `form:"access_token" binding:"required"`
	TokenType   string `form:"token_type" binding:"required"`
	ExpiresIn   int    `form:"expires_in" binding:"required,min=1"`
	State       string `form:"state" binding:"required"`
}

type errorResponse struct {
	Error       string `form:"error" binding:"required"`
	Description string `form:"error_description" binding:"required"`
	State       string `form:"state" binding:"required"`
}

func auth(cc *cli.Context) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	sessionCfg, err := sessionConfig(cc.String("session-config"))
	if err != nil {
		return err
	}
	store := cookie.NewStore(sessionCfg.HashKey, sessionCfg.BlockKey)
	tmpl := template.Must(template.New("").ParseFiles(
		"tmpl/index.tmpl",
		"tmpl/auth.tmpl",
		"tmpl/error.tmpl",
	))
	r.SetHTMLTemplate(tmpl)
	r.Use(sessions.Sessions(cc.App.Name, store))
	r.GET("/", index(cc))
	r.GET("/auth", authGET(cc))
	r.POST("/auth", authPOST(cc))
	log.Println("Server initialization finished.  " +
		"Access http://localhost:8080 from your borwser")
	r.Run(":8080")
	return nil
}

func index(cc *cli.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := sessions.Default(c)
		state := uuid.Must(uuid.NewUUID()).String()
		s.Set("state", state)
		var cfg *Config
		cfg, err := loadConfig(cc.String("config"))
		if err == nil {
			if err = cfg.Validate(); err != nil {
				s.AddFlash(fmt.Errorf("トークンが異常です: %v", err))
			}
		}
		if err = s.Save(); err != nil {
			error500(c, err)
			return
		}
		u, err := iijURL(&state)
		if err != nil {
			error500(c, err)
			return
		}
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"IIJURL":  u.String(),
			"Flashes": s.Flashes(),
		})
	}
}

func authGET(cc *cli.Context) gin.HandlerFunc {
	return func(c *gin.Context) { c.HTML(http.StatusOK, "auth.tmpl", nil) }
}

func authPOST(cc *cli.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := sessions.Default(c)
		state, ok := s.Get("state").(string)
		if !ok {
			error400(c, errors.New("state not found"))
			return
		}
		var er errorResponse
		if err := c.ShouldBind(&er); err == nil {
			if er.State != state {
				error400(c, errors.New("invalid state"))
				return
			}
			error400(c,
				fmt.Errorf("error: %s, description: %s", er.Error, er.Description))
			return
		}
		var f validResponse
		if err := c.ShouldBind(&f); err != nil {
			error400(c, err)
			return
		}
		if f.State != state {
			error400(c, errors.New("invalid state"))
			return
		}
		err := saveConfig(cc.String("config"), f.AccessToken, f.ExpiresIn)
		if err != nil {
			error500(c, err)
			return
		}
		s.AddFlash("Config file successfully created!")
		if err := s.Save(); err != nil {
			error500(c, err)
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}
}

func iijURL(state *string) (*url.URL, error) {
	u, err := url.Parse(iijAPIURL)
	if err != nil {
		return nil, err
	}
	v := u.Query()
	v.Set("response_type", "token")
	v.Set("client_id", "")
	v.Set("redirect_uri", authURL)
	v.Set("state", *state)
	u.RawQuery = v.Encode()
	return u, nil
}

func error500(c *gin.Context, err error) {
	c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
		"Error": err.Error(),
	})
}

func error400(c *gin.Context, err error) {
	c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
		"Error": err.Error(),
	})
}