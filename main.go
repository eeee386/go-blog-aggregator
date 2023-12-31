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
	"sync"
	"time"
	"strconv"

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
	Url  string `json:"url"`
}

type CreateFeedResponse struct {
	Feed       database.Feed       `json:"feed"`
	FeedFollow database.Feedfollow `json:"feed_follow"`
}

type CreateFeedFollowRequest struct {
	FeedId string `json:"feed_id"`
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
	dbObjf := database.CreateFeedFollowParams{}
	dbObjf.ID = uuid.New()
	dbObjf.CreatedAt = timeStamp
	dbObjf.UpdatedAt = timeStamp
	dbObjf.FeedID = feed.ID
	dbObjf.UserID = user.ID
	feedFollow, err := a.DB.CreateFeedFollow(a.ctx, dbObjf)
	if err != nil {
		utils.RespondWithError(w, 500, "internal Server Error")
		return
	}
	createFeedResponse := CreateFeedResponse{}
	createFeedResponse.Feed = feed
	createFeedResponse.FeedFollow = feedFollow
	utils.RespondWithJSON(w, 201, createFeedResponse)
}

func (a *apiConfig) GetAllFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := a.DB.GetAllFeeds(a.ctx)
	if err != nil {
		utils.RespondWithError(w, 500, "Internal Server Error")
	} else {
		utils.RespondWithJSON(w, 200, feeds)
	}
}

func (a *apiConfig) CreateFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
	decoder := json.NewDecoder(r.Body)
	feedFollowReq := CreateFeedFollowRequest{}
	err := decoder.Decode(&feedFollowReq)
	if err != nil {
		utils.RespondWithError(w, 400, "Bad Request")
		return
	}
	feedId, err := uuid.Parse(feedFollowReq.FeedId)
	if err != nil {
		utils.RespondWithError(w, 400, "Invalid UUID string")
		return
	}
	feed, err := a.DB.GetFeedById(a.ctx, feedId)
	if err != nil {
		utils.RespondWithError(w, 500, "Internal Server Error")
		return
	}
	dbObj := database.CreateFeedFollowParams{}
	dbObj.ID = uuid.New()
	timeStamp := time.Now()
	dbObj.CreatedAt = timeStamp
	dbObj.UpdatedAt = timeStamp
	dbObj.FeedID = feed.ID
	dbObj.UserID = user.ID
	feedFollow, err := a.DB.CreateFeedFollow(a.ctx, dbObj)
	if err != nil {
		utils.RespondWithError(w, 500, "internal Server Error")
		return
	}
	utils.RespondWithJSON(w, 201, feedFollow)
}

func (a *apiConfig) DeleteFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
	id := chi.URLParam(r, "id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		utils.RespondWithError(w, 400, "Invalid UUID")
		return
	}
	derr := a.DB.DeleteFeedFollowById(a.ctx, uuid)
	if derr != nil {
		utils.RespondWithError(w, 400, "Bad Request")
		return
	}
	utils.RespondWithJSON(w, 200, "OK")
}


func (a *apiConfig) GetAllFeedFollowsByUser(w http.ResponseWriter, r *http.Request, user database.User) {
	feedFollows, err := a.DB.GetAllFeedFollowsByUserID(a.ctx, user.ID)
	if err != nil {
		utils.RespondWithError(w, 500, "Internal Server Error")
		return
	}
	utils.RespondWithJSON(w, 200, feedFollows)
}

func (a *apiConfig) GetAllPostsByUserId(w http.ResponseWriter, r *http.Request, user database.User) {
	limitStr := chi.URLParam(r, "limit")
	dbObj := database.GetPostsByUserParams{}
	limit, cerr := strconv.Atoi(limitStr)
	if cerr != nil {
		utils.RespondWithError(w, 500, "Internal Server Error")
	}
	dbObj.Limit = int32(limit)
	dbObj.UserID = user.ID
	postsByUser, err := a.DB.GetPostsByUser(a.ctx, dbObj)
	if err != nil {
		utils.RespondWithError(w, 500, "Internal Server Error")
		return 
	}
	utils.RespondWithJSON(w, 200, postsByUser)
}

func (a *apiConfig) Worker(n int32) {
	fmt.Println("Called")
	for tick := range time.Tick(6 * time.Second) {
		fmt.Println("Tick")
	  var wg sync.WaitGroup
		fmt.Println(tick);
		dbObj := database.GetNextFeedsToFetchParams{}
		dbObj.Limit = n
		dbObj.Offset = 0
		feeds, err := a.DB.GetNextFeedsToFetch(a.ctx, dbObj);
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			return 
		}
		for _, feed := range(feeds) {
			wg.Add(1)
			go func(url string, id uuid.UUID) {
				defer wg.Done()
				go ProcessRSSURL(url, id, a)
				dbObj := database.MarkFeedFetchedParams{}
				dbObj.ID = id
				timeObj := sql.NullTime{}
				timeObj.Time = time.Now()
				dbObj.LastFetchedAt = timeObj
				fmt.Println("mark fetched called")
				a.DB.MarkFeedFetched(a.ctx, dbObj)
			}(feed.Url, feed.ID)
		}
		wg.Wait()
	}
}

func ProcessRSSURL(url string, feedId uuid.UUID, a *apiConfig) {
	rss := utils.RSS{}
	cerr := utils.RSSUrlToStruct(url, &rss);
	if cerr != nil {
		fmt.Printf("ERROR: %s", cerr.Error())
		return
	}
	
	fmt.Printf("Channel title: %v\n", rss.Channel.Title)
	fmt.Printf("Channel link: %v\n", rss.Channel.Link)

	for i, item := range rss.Channel.Items {
		fmt.Printf("%v. item title: %v\n", i, item.Title)
    publishDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
		  fmt.Printf("ERROR: %s", err.Error())
			return
		}
		dbObj := database.CreatePostParams{}
		dbObj.ID = uuid.New()
		dbObj.FeedID = feedId
		dbObj.Title = item.Title
		dbObj.PublishedAt = publishDate
		dbObj.Url = item.Link
		dbObj.Description = item.Desc
		timeStamp := time.Now()
		dbObj.CreatedAt = timeStamp
		dbObj.UpdatedAt = timeStamp
		// inserts in PostgreSQL cannot be parallelized, only with multiple connections
		_, derr := a.DB.CreatePost(a.ctx, dbObj)
		if derr != nil {
			fmt.Printf("ERROR: %s\n", derr);
			return
		}
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

	// user endpoints
	v1Router.Post("/users", apiCfg.createUser)
	v1Router.Get("/users", apiCfg.authenticate(GetUser))

	// feed endpoints
	v1Router.Post("/feeds", apiCfg.authenticate(apiCfg.CreateFeed))
	v1Router.Get("/feeds", apiCfg.GetAllFeeds)

	// feed follow endpoints
	v1Router.Post("/feed_follows", apiCfg.authenticate(apiCfg.CreateFeedFollow))
	v1Router.Delete("/feed_follows/{id}", apiCfg.authenticate(apiCfg.DeleteFeedFollow))
	v1Router.Get("/feed_follows", apiCfg.authenticate(apiCfg.GetAllFeedFollowsByUser))

	v1Router.Get("/posts", apiCfg.authenticate(apiCfg.GetAllPostsByUserId))

	go apiCfg.Worker(60);

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), r)
	fmt.Println(err)
}
