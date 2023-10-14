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
	DB *database.Queries
	ctx context.Context
}

type CreateUserRequest struct {
	Name string `json:"name"`
}

type CreateUserResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string `json:"name"`
}

func (a apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
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
		utils.RespondWithError(w, 500, "Internal Server Error")
		return
	}
	userRes := CreateUserResponse{}
	userRes.ID = user.ID
	userRes.Name = user.Name
	userRes.CreatedAt = user.CreatedAt
	userRes.UpdatedAt = user.UpdatedAt
	utils.RespondWithJSON(w, 200, userRes)
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

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), r)
	fmt.Println(err)
}
