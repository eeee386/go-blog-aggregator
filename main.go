package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"go-blog-aggregator/internal/database"
	"go-blog-aggregator/utils"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	utils.RespondWithJSON(w, 200, struct {
		Status string `json:"status"`
	}{Status: "OK"})
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	utils.RespondWithError(w, 500, "Internal Server Error")
}

type apiConfig struct {
	DB  *database.Queries
	ctx context.Context
}

type CreateUserRequest struct {
	Name string `json:"name"`
}


type CreateFeedRequest struct {
	Name string `json:"name"`
	Url string `json:"url"`
}


func (a *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userReq := CreateUserRequest{}
	err := decoder.Decode(&userReq)
	if err != nil {
		utils.RespondWithError(w, 500, "Something went wrong")
		return
	}
	dbObj := database.CreateUserParams{}
	dbObj.ID = uuid.New()
	timeStamp := time.Now()
	dbObj.CreatedAt = timeStamp
	dbObj.UpdatedAt = timeStamp
	dbObj.Name = userReq.Name

	user, err := a.DB.CreateUser(a.ctx, dbObj)
	if err != nil {
		utils.RespondWithError(w, 400, "Bad Request")
		return
	}
	utils.RespondWithJSON(w, 201, user)
}

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (a *apiConfig) authenticate(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bearerToken := r.Header.Get("Authorization")
		apikey := strings.Split(bearerToken, " ")[1]
		user, err := a.DB.GetUserByApiKey(a.ctx, apikey)
		if err != nil {
			utils.RespondWithError(w, 401, "Unauthorized")
		} else {
			handler(w, r, user)
		}
	}
}

func GetUser(w http.ResponseWriter, r *http.Request, user database.User) {
	utils.RespondWithJSON(w, 200, user)
}

func (a *apiConfig) CreateFeed(w http.ResponseWriter, r *http.Request, user database.User) {

	decoder := json.NewDecoder(r.Body)
	feedReq := CreateFeedRequest{}
	err := decoder.Decode(&feedReq)
	if err != nil {
		utils.RespondWithError(w, 400, "Bad Request")
		return
	}
	dbObj := database.CreateFeedParams{}
	dbObj.ID = uuid.New()
	timeStamp := time.Now()
	dbObj.CreatedAt = timeStamp
	dbObj.UpdatedAt = timeStamp
	dbObj.Name = feedReq.Name
	dbObj.UserID = user.ID
	dbObj.Url = feedReq.Url
  feed, err := a.DB.CreateFeed(a.ctx, dbObj)
	if err != nil {
		utils.RespondWithError(w, 500, "Internal Server Error")
		return
	}
	utils.RespondWithJSON(w, 201, feed)
}

func (a *apiConfig) GetAllFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := a.DB.GetAllFeeds(a.ctx);
	if err != nil {
		utils.RespondWithError(w, 500, "Internal Server Error")
	} else {
	  utils.RespondWithJSON(w, 200, feeds)
	}
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")

	dbURL := os.Getenv("CONN")
	db, derr := sql.Open("postgres", dbURL)
	if derr != nil {
		fmt.Println(derr.Error())
		return
	}

	dbQueries := database.New(db)
	apiCfg := apiConfig{}
	apiCfg.DB = dbQueries
	apiCfg.ctx = context.Background()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET, POST, OPTIONS, PUT, DELETE"},
		AllowedHeaders: []string{"*"},
	}))

	v1Router := chi.NewRouter()
	r.Mount("/v1", v1Router)

	v1Router.Get("/readiness", readinessHandler)
	v1Router.Get("/err", errorHandler)

	v1Router.Post("/users", apiCfg.createUser)
	v1Router.Get("/users", apiCfg.authenticate(GetUser))

	v1Router.Post("/feeds", apiCfg.authenticate(apiCfg.CreateFeed))
	v1Router.Get("/feeds", apiCfg.GetAllFeeds)

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), r)
	fmt.Println(err)
}
