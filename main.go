package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	h "microservice-shortener/api"
	mr "microservice-shortener/repository/mongo"
	"microservice-shortener/shortener"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	//"github.com/joho/godotenv"
)

func main() {
	//err := godotenv.Load()
	/*
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	*/

	err := godotenv.Load()
	_, localEnvSettled := os.LookupEnv("MONGO_URL")
	if err != nil && localEnvSetted == false {
		log.Fatal("Error loading .env file")
	}
	repo := chooseRepo()
	service := shortener.NewRedirectService(repo)
	handler := h.NewHandler(service)
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/{code}", handler.Get)
	r.Get("/", handler.Post)

	errs := make(chan error, 2)
	go func() {
		fmt.Println("Listening on port :8000")
		errs <- http.ListenAndServe(httpPort(), r)

	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	fmt.Printf("Terminated %s", <-errs)

}

func httpPort() string {
	port := "8000"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	return fmt.Sprintf(":%s", port)
}
func chooseRepo() shortener.RedirectRepository {
	if os.Getenv("URL_DB") == "mongo" {
		mongoURL := os.Getenv("MONGO_URL")
		mongodb := os.Getenv("MONGO_DB")
		mongoTimeout, _ := strconv.Atoi(os.Getenv("MONGO_TIMEOUT"))
		repo, err := mr.NewMongoRepository(mongoURL, mongodb, mongoTimeout)

		if err != nil {
			log.Fatal(err)
		}
		return repo
	}
	return nil
}
