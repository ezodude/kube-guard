package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/ezodude/kube-guard/privelage"
)

func newK8s() (kubernetes.Interface, error) {
	home := homedir.HomeDir()
	if home == "" {
		return nil, fmt.Errorf("Could not find HOME directory")
	}

	path := filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

type searchPayload struct {
	Subjects []string `json:"subjects"`
	Format   string   `json:"format"`
}

type app struct {
	router *mux.Router
	k8s    kubernetes.Interface
}

func (a *app) initialize() {
	log.Println("App initializing")
	a.router = mux.NewRouter()

	a.router.HandleFunc("/api/v0.1/privelage/search", a.searchHandler).
		Methods("GET").
		Headers("Content-Type", "application/json")
}

func (a *app) searchHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handler searchHandler request received")

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s status [%d]: %s\n", r.RequestURI, http.StatusInternalServerError, err.Error())
		return
	}
	defer r.Body.Close()

	var payload searchPayload
	err = json.Unmarshal(data, &payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s status [%d]: %s\n", r.RequestURI, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("Handler searchHandler payload:[%#v]", payload)

	res, err := privelage.NewQuery().
		Client(a.k8s).
		Subjects(payload.Subjects).
		ResultFormat(payload.Format).
		Do()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s status [%d]: %s\n", r.RequestURI, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("%s status [%d]\n", r.RequestURI, http.StatusOK)

	switch strings.ToLower(payload.Format) {
	case "yaml", "yml":
		w.Header().Set("Content-Type", "application/x-yaml")
	default:
		w.Header().Set("Content-Type", "application/json")
	}

	w.Write(res)
}

func (a *app) run(port string) {
	addr := fmt.Sprintf(":%s", port)

	srv := &http.Server{
		Addr: addr,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      a.router,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		log.Printf("Running server on %s\n", addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)
}

func main() {
	k8s, err := newK8s()
	if err != nil {
		panic(err)
	}
	a := app{}
	a.k8s = k8s
	a.initialize()
	a.run("8080")
}
