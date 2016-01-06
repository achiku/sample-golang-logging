package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	logrus "github.com/Sirupsen/logrus"
	"github.com/doloopwhile/logrusltsv"
	"github.com/go-zoo/bone"
	"github.com/rs/xhandler"
	"golang.org/x/net/context"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func normalLoggingMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("access [%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}
	return http.HandlerFunc(fn)
}

func log15LoggingMiddleware(next http.Handler) http.Handler {
	srvlog := log15.New("module", "app/server")
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		srvlog.Info("access", "method", r.Method, "path", r.URL.String(), "req_time", t2.Sub(t1))
	}
	return http.HandlerFunc(fn)
}

func logrusLoggingMiddleware(next http.Handler) http.Handler {
	logrus.SetFormatter(&logrusltsv.Formatter{})
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		logrus.WithFields(logrus.Fields{
			"method":   r.Method,
			"path":     r.URL.String(),
			"req_time": t2.Sub(t1),
		}).Info("access")
	}
	return http.HandlerFunc(fn)
}

func recoverMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("panic: %+v", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func account(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	accountId := bone.GetValue(req, "id")
	fmt.Fprintf(rw, "accountId: %s", accountId)
}

func note(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	noteId := bone.GetValue(req, "id")
	fmt.Fprintf(rw, "noteId: %s", noteId)
}

func simple(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "Hello, world!!")
}

func main() {
	c := xhandler.Chain{}
	c.Use(recoverMiddleware)
	c.Use(normalLoggingMiddleware)
	c.Use(log15LoggingMiddleware)
	c.Use(logrusLoggingMiddleware)

	simpleHandler := xhandler.HandlerFuncC(simple)
	accountHandler := xhandler.HandlerFuncC(account)
	noteHandler := xhandler.HandlerFuncC(note)

	mux := bone.New()
	mux.Get("/account/:id", c.Handler(accountHandler))
	mux.Get("/note/:id", c.Handler(noteHandler))
	mux.Get("/simple", c.Handler(simpleHandler))
	http.ListenAndServe(":8080", mux)
}
