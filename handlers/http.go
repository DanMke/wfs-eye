package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	cors "github.com/itsjamie/gin-cors"
)

type HTTPServer struct {
	server *http.Server
	router *gin.Engine
}

type Options struct {
	WFSURL string
}

func NewHTTPServer(opt Options) *HTTPServer {
	router := gin.Default()

	router.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET",
		RequestHeaders:  "Origin, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          24 * 3600 * time.Second,
		Credentials:     false,
		ValidateHeaders: false,
	}))

	h := &HTTPServer{server: &http.Server{
		Addr:    ":4000",
		Handler: router,
	}, router: router}

	h.setupWFSHandlers(opt)
	h.setupAPIHandlers(opt)

	return h
}

//Start the main HTTP Server entry
func (s *HTTPServer) Start() error {
	logrus.Infof("Starting HTTP Server on port 3000")
	return s.server.ListenAndServe()
}
