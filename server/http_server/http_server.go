package http_server

import (
	"fmt"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"net/http"
	"pmail/config"
	"pmail/controllers"
	"pmail/controllers/email"
	"pmail/session"
	"time"
)

// 这个服务是为了拦截http请求转发到https
var httpServer *http.Server

func HttpStop() {
	if httpServer != nil {
		httpServer.Close()
	}
}

func HttpStart() {
	mux := http.NewServeMux()

	HttpPort := 80
	if config.Instance.HttpPort > 0 {
		HttpPort = config.Instance.HttpPort
	}

	if config.Instance.HttpsEnabled != 2 {
		mux.HandleFunc("/", controllers.Interceptor)
		httpServer = &http.Server{
			Addr: fmt.Sprintf(":%d", HttpPort),
			Handler: cors.New(cors.Options{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders: []string{"account", "password"},
			}).Handler(mux),
			ReadTimeout:  time.Second * 90,
			WriteTimeout: time.Second * 90,
		}
	} else {
		fe, err := fs.Sub(local, "dist")
		if err != nil {
			panic(err)
		}
		mux.Handle("/", http.FileServer(http.FS(fe)))
		// 挑战请求类似这样 /.well-known/acme-challenge/QPyMAyaWw9s5JvV1oruyqWHG7OqkHMJEHPoUz2046KM
		mux.HandleFunc("/.well-known/", controllers.AcmeChallenge)
		mux.HandleFunc("/api/ping", contextIterceptor(controllers.Ping))
		mux.HandleFunc("/api/login", contextIterceptor(controllers.Login))
		mux.HandleFunc("/api/group", contextIterceptor(controllers.GetUserGroup))
		mux.HandleFunc("/api/group/list", contextIterceptor(controllers.GetUserGroupList))
		mux.HandleFunc("/api/group/add", contextIterceptor(controllers.AddGroup))
		mux.HandleFunc("/api/group/del", contextIterceptor(controllers.DelGroup))
		mux.HandleFunc("/api/email/list", contextIterceptor(email.EmailList))
		mux.HandleFunc("/api/email/del", contextIterceptor(email.EmailDelete))
		mux.HandleFunc("/api/email/read", contextIterceptor(email.MarkRead))
		mux.HandleFunc("/api/email/detail", contextIterceptor(email.EmailDetail))
		mux.HandleFunc("/api/email/move", contextIterceptor(email.Move))
		mux.HandleFunc("/api/email/send", contextIterceptor(email.Send))
		mux.HandleFunc("/api/settings/modify_password", contextIterceptor(controllers.ModifyPassword))
		mux.HandleFunc("/api/rule/get", contextIterceptor(controllers.GetRule))
		mux.HandleFunc("/api/rule/add", contextIterceptor(controllers.UpsertRule))
		mux.HandleFunc("/api/rule/update", contextIterceptor(controllers.UpsertRule))
		mux.HandleFunc("/api/rule/del", contextIterceptor(controllers.DelRule))
		mux.HandleFunc("/attachments/", contextIterceptor(controllers.GetAttachments))
		mux.HandleFunc("/attachments/download/", contextIterceptor(controllers.Download))
		log.Infof("HttpServer Start On Port :%d", HttpPort)
		httpServer = &http.Server{
			Addr: fmt.Sprintf(":%d", HttpPort),
			Handler: session.Instance.LoadAndSave(cors.New(cors.Options{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders: []string{"account", "password"},
			}).Handler(mux)),
			ReadTimeout:  time.Second * 90,
			WriteTimeout: time.Second * 90,
		}
	}

	err := httpServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
