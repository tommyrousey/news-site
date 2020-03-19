package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// create a html template output
// the must makes sure execution does not continue with a bad template
var tpl = template.Must(template.ParseFiles("index.html"))

// api key for news api
var apiKey *string

// struct to convert json response from news api
type Results struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}
type Source struct {
	ID   interface{} `json:"id"`
	Name string      `json:"name"`
}
type Article struct {
	Source      Source    `json:"source"`
	Author      string    `json:"author"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	URLToImage  string    `json:"urlToImage"`
	PublishedAt time.Time `json:"publishedAt"`
	Content     string    `json:"content"`
}

// struct for search
type Search struct {
	SearchKey  string
	NextPage   int
	TotalPages int
	Results    Results
}

// formats the date of article to be more readable
func (a *Article) FormatPublishedDate() string {
	year, month, day := a.PublishedAt.Date()
	return fmt.Sprintf("%v %d %d", month, day, year)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// generates html safe from code injection
	tpl.Execute(w, nil)
}

// handles search requests for news articles
func searchHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	// error out if url is bad
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	params := u.Query()
	// get user query
	searchKey := params.Get("q")
	page := params.Get("page")
	if page == "" {
		page = "1"
	}

	// create new search struct
	search := &Search{}
	// set the search key to q parameter in the http request
	search.SearchKey = searchKey

	// convert page into integer and assign it to next page parameter
	next, err := strconv.Atoi(page)
	if err != nil {
		http.Error(w, "Unexpected server error",
			http.StatusInternalServerError)
		return
	}

	search.NextPage = next
	// make a default page size (0 - 100)
	pageSize := 20

	// construct endpoint by making a GET request
	endpoint := fmt.Sprintf(
		"https://newsapi.org/v2/everything?q=%s&pageSize=%d&page=%d&apiKey=%s&sortBy=publishedAt&language=en",
		url.QueryEscape(search.SearchKey), pageSize, search.NextPage, *apiKey)

	resp, err := http.Get(endpoint)

	// ensure that the response code is 200
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// parse the json response into struct
	err = json.NewDecoder(resp.Body).Decode(&search.Results)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	search.TotalPages = int(math.Ceil(float64(search.Results.TotalResults / pageSize)))
	err = tpl.Execute(w, search)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func main() {

	// get api key from flags, no default value for this
	apiKey = flag.String("apikey", "", "Newsapi.org access key")
	flag.Parse()

	if *apiKey == "" {
		log.Fatal("apiKey must be set")
	}

	// get a port
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// make a server mux
	mux := http.NewServeMux()

	// creates a file server obj by passig the
	// directory where all static files are placed
	fs := http.FileServer(http.Dir("assets"))
	// tell mux to use this file server obj for all paths beginning with /assets/
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// tell mux to handle things in the search directory
	mux.HandleFunc("/search", searchHandler)

	// give the mux the indexHandler function
	mux.HandleFunc("/", indexHandler)
	// listen on the given port with the mux
	http.ListenAndServe(":"+port, mux)

}
