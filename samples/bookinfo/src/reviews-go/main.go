package main

import (
	"encoding/json"
	"fmt"
	"log"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// docker build -t istio/examples-bookinfo-reviews-v4:1.17.0 -f Dockerfile .

/*
docker build -t istio/examples-bookinfo-reviews-v4:1.17.0 -f Dockerfile .  --build-arg service_version=v3 \
	   --build-arg enable_ratings=true --build-arg star_color=red 
*/
func main() {
	InitConst()
	r := mux.NewRouter()
	r.HandleFunc("/health", Health)
	r.HandleFunc("/reviews/{productId}", ReviewProduct).Methods("GET")

	http.ListenAndServe(":9080", r)
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "{\"status\": \"Reviews is healthy\"}")
}

func ReviewProduct(w http.ResponseWriter, r *http.Request) {
	values := mux.Vars(r)

	starsReviewer1 := -1
	starsReviewer2 := -1
	productId := values["productId"]
	if ratings_enabled {
		result := getRatings(productId, r)
		if result != nil {
			if result.Rating != nil {
				starsReviewer1 = result.Rating.Reviewer1
				starsReviewer2 = result.Rating.Reviewer2
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, getJsonResponse(productId, starsReviewer1, starsReviewer2))
}

type RationResponse struct {
	Id     string  `json:"id"`
	Rating *Rating `json:"ratings"`
}

type Rating struct {
	Reviewer1 int
	Reviewer2 int
}

func getRatings(productId string, r *http.Request) *RationResponse {
	response := &RationResponse{}
	path := fmt.Sprintf("%v/%v", ratings_service, productId)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		log.Println(err.Error())
		return response
	}
	for _, h := range headers_to_propagate {
		v := r.Header.Get(h)
		req.Header.Set(h, v)
	}

	timeOut := 10000
	if strings.EqualFold(star_color, "black") {
		timeOut = 25000
	}

	client := &http.Client{}
	client.Timeout = time.Duration(timeOut) * time.Millisecond

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return response
	}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, response)
	return response
}

func getJsonResponse(productId string, s1, s2 int) string {
	starsReviewer1 := strconv.Itoa(s1)
	starsReviewer2 := strconv.Itoa(s2)
	result := "{"
	result += "\"id\": \"" + productId + "\","
	result += "\"podname\": \"" + pod_hostname + "\","
	result += "\"clustername\": \"" + clustername + "\","
	result += "\"reviews\": ["

	// reviewer 1:
	result += "{"
	result += "  \"reviewer\": \"Reviewer1\","
	result += "  \"text\": \"An extremely entertaining play by Shakespeare. The slapstick humour is refreshing!\""
	if ratings_enabled {
		if s1 != -1 {
			result += ", \"rating\": {\"stars\": " + starsReviewer1 + ", \"color\": \"" + star_color + "\"}"
		} else {
			result += ", \"rating\": {\"error\": \"Ratings service is currently unavailable\"}"
		}
	}
	result += "},"

	// reviewer 2:
	result += "{"
	result += "  \"reviewer\": \"Reviewer2\","
	result += "  \"text\": \"Absolutely fun and entertaining. The play lacks thematic depth when compared to other plays by Shakespeare.\""
	if ratings_enabled {
		if s2 != -1 {
			result += ", \"rating\": {\"stars\": " + starsReviewer2 + ", \"color\": \"" + star_color + "\"}"
		} else {
			result += ", \"rating\": {\"error\": \"Ratings service is currently unavailable\"}"
		}
	}
	result += "}"

	result += "]"
	result += "}"

	return result
}
