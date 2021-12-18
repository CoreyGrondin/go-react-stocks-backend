package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	_ "github.com/lib/pq"
)

func main() {

	//Create the default mux
	mux := http.NewServeMux()

	//Handling the /v1/teachers. The handler is a function here
	mux.HandleFunc("/", Serve)

	//Create the server.
	s := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	err := s.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}

type cat struct {
	Fact   string `json:"fact"`
	Length int    `json:"length"`
}

func index(w http.ResponseWriter, r *http.Request) {

	resp, err := http.Get("https://catfact.ninja/fact")
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	var cat1 cat

	if err := json.NewDecoder(resp.Body).Decode(&cat1); err != nil {
		log.Fatalln(err)
	}

	if err := json.NewEncoder(w).Encode(cat1); err != nil {
		log.Fatalln(err)
	}
}

func indexPost(w http.ResponseWriter, r *http.Request) {

	resp, err := http.Get("https://api.coindesk.com/v1/bpi/currentprice.json")
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	var cat1 cat

	if err := json.NewDecoder(resp.Body).Decode(&cat1); err != nil {
		log.Fatalln(err)
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Fatalln(err)
	}

}


func teacherHandler(res http.ResponseWriter, req *http.Request) {
	data := []byte("V1 of teacher's called")
	res.WriteHeader(200)
	_, err := res.Write(data)
	if err != nil {
		log.Fatalln(err)
	}

}

func home(res http.ResponseWriter, req *http.Request) {
	data := []byte("home page :)")
	res.WriteHeader(200)
	_, err := res.Write(data)
	if err != nil {
		log.Fatalln(err)
	}
}

var routes = []route{
	newRoute("GET", "/", home),
	newRoute("GET", "/contact", /*contact*/ index),
	newRoute("GET", "/api/widgets", /*apiGetWidgets*/ teacherHandler),
	newRoute("POST", "/api/widgets", indexPost),
	newRoute("POST", "/api/insert", dbinsert),
	newRoute("GET", "/api/insert", dbget),
	// newRoute("POST", "/api/widgets/([^/]+)", apiUpdateWidget),
	// newRoute("POST", "/api/widgets/([^/]+)/parts", apiCreateWidgetPart),
	// newRoute("POST", "/api/widgets/([^/]+)/parts/([0-9]+)/update", apiUpdateWidgetPart),
	// newRoute("POST", "/api/widgets/([^/]+)/parts/([0-9]+)/delete", apiDeleteWidgetPart),
	// newRoute("GET", "/([^/]+)", widget),
	// newRoute("GET", "/([^/]+)/admin", widgetAdmin),
	// newRoute("POST", "/([^/]+)/image", widgetImage),
}

func newRoute(method, pattern string, handler http.HandlerFunc) route {
	return route{method, regexp.MustCompile("^" + pattern + "$"), handler}
}

type route struct {
	method  string
	regex   *regexp.Regexp
	handler http.HandlerFunc
}

type ctxKey struct{}

func Serve(w http.ResponseWriter, r *http.Request) {
	var allow []string
	for _, route := range routes {
		matches := route.regex.FindStringSubmatch(r.URL.Path)
		if len(matches) > 0 {
			if r.Method != route.method {
				allow = append(allow, route.method)
				continue
			}
			ctx := context.WithValue(r.Context(), ctxKey{}, matches[1:])
			route.handler(w, r.WithContext(ctx))
			return
		}
	}
	if len(allow) > 0 {
		w.Header().Set("Allow", strings.Join(allow, ", "))
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.NotFound(w, r)
}

type configuration struct {
    Host string
    Port int
    User string
	Password string
    Dbname string
}


func dbinsert(w http.ResponseWriter, r *http.Request) {

	resp, err := http.Get("https://catfact.ninja/fact")
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	var cat1 cat

	if err := json.NewDecoder(resp.Body).Decode(&cat1); err != nil {
		log.Fatalln(err)
	}

	var config configuration

	var filename = "./config/config.development.json"

	//filename is the path to the json config file
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {  panic(err)}


	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.Host, config.Port, config.User, config.Password, config.Dbname)

    db, _ := sql.Open("postgres", psqlconn)

    defer db.Close()

    // dynamic
    insertDynStmt := `insert into "catfacts"("fact", "length") values($1, $2)`
    _, _ = db.Exec(insertDynStmt, cat1.Fact, cat1.Length)

	if err := json.NewEncoder(w).Encode(cat1); err != nil {
		log.Fatalln(err)
	}

}

func dbget(w http.ResponseWriter, r *http.Request) {

	resp, err := http.Get("https://catfact.ninja/fact")
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	var cat1 cat

	if err := json.NewDecoder(resp.Body).Decode(&cat1); err != nil {
		log.Fatalln(err)
	}

	var config configuration

	var filename = "./config/config.development.json"

	//filename is the path to the json config file
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {  panic(err)}


	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.Host, config.Port, config.User, config.Password, config.Dbname)

    db, _ := sql.Open("postgres", psqlconn)

    defer db.Close()

    // dynamic
	rows, _ := db.Query(`SELECT "fact", "length" FROM "catfacts"`)

	//fmt.Println(rows)

	defer rows.Close()
    for rows.Next() {
		var fact string
		var length int

		err = rows.Scan(&fact, &length)
		CheckError(err)

		fmt.Println(fact, length)
	}

	if err := json.NewEncoder(w).Encode(cat1); err != nil {
		log.Fatalln(err)
	}

}

func CheckError(err error) {
    if err != nil {
        panic(err)
    }
}
