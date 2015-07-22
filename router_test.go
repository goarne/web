package web

import (
	//	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestAddSingleRoute(t *testing.T) {
	router := NewWebRouter()
	router.AddRoute(NewRoute().Path("/"))

	if router.routeCount() != 1 {
		t.Errorf("Could not add route.")
	}
}

func TestThreeRoutes(t *testing.T) {
	router := NewWebRouter()

	router.AddRoute(NewRoute().Path("/route"))
	router.AddRoute(NewRoute().Path("/route/articles/").Method(HttpOption))
	router.AddRoute(NewRoute().Path("/route/articles/1").Method(HttpGet))

	if router.routeCount() != 3 {
		t.Errorf("Could not add three routes.")
	}
}

func TestFinder(t *testing.T) {

	router := NewWebRouter()
	router.AddRoute(NewRoute().Path("/test"))
	found := router.findRoute(createGetRequest("/test"))

	if found == nil {
		t.Errorf("Could not find route.")
	}
}

func TestFindRouteWithMethod(t *testing.T) {
	router := NewWebRouter()

	router.AddRoute(NewRoute().Path("/route").Method(HttpGet))
	router.AddRoute(NewRoute().Path("/route/articles/").Method(HttpGet))
	router.AddRoute(NewRoute().Path("/route/articles/1").Method(HttpGet))

	found := router.findRoute(createGetRequest("/route/articles/"))

	if found == nil {
		t.Errorf("Could not find route with method.")
	}
}

func TestFindRouteWithId(t *testing.T) {
	router := NewWebRouter()

	router.AddRoute(NewRoute().Path("/route/too/{id}").Method(HttpGet))
	router.AddRoute(NewRoute().Path("/route/articles/{name}").Method(HttpGet))

	found := router.findRoute(createGetRequest("/route/articles/joe"))

	if found == nil || found.urlVariables["name"] != "joe" {
		t.Errorf("Could not find route with id.")
	}
}

func TestFindRouteWithIdVariable(t *testing.T) {
	router := NewWebRouter()

	urlWithIdVariable := "/route/articles/{id:\\d*}/detaljer/"
	router.AddRoute(NewRoute().Path("/route/articles/{name}").Method(HttpGet))
	router.AddRoute(NewRoute().Path(urlWithIdVariable).Method(HttpGet))
	found := router.findRoute(createGetRequest("/route/articles/21/detaljer/"))

	if found == nil || found.urlVariables["id"] != "21" {
		t.Errorf("Could not find correct route with id variable.")
	}
}

func TestFindRouteWithPost(t *testing.T) {
	router := NewWebRouter()

	router.AddRoute(NewRoute().Path("/route/articles/").Method(HttpPost))

	found := router.findRoute(createPostRequest("/route/articles/"))

	if found == nil {
		t.Errorf("Could not find route with POST method.")
	}
}

func TestCannotFindPostRoute(t *testing.T) {
	router := NewWebRouter()

	router.AddRoute(NewRoute().Path("/route/articles/").Method(HttpGet))
	router.AddRoute(NewRoute().Path("/route/articles/test").Method(HttpGet))

	found := router.findRoute(createPostRequest("/route/article"))

	if found != nil {
		t.Errorf("Found route with POST method.")
	}
}

func TestFindRouteWithContentTypeJSON(t *testing.T) {
	router := NewWebRouter()

	router.AddRoute(NewRoute().Path("/route/articles/").Method(HttpGet).Header(HttpAccept, "application/json"))
	router.AddRoute(NewRoute().Path("/route/articles/edit/").Method(HttpGet).Header(HttpAccept, "application/html"))

	req := createGetRequest("/route/articles/")
	req.Header.Add(HttpAccept.String(), "application/json")

	found := router.findRoute(req)

	if found == nil {
		t.Errorf("Could not find route with correct content type.")
	}
}

func TestCannotFindRouteWithCorrectContentTypeHTML(t *testing.T) {
	router := NewWebRouter()

	router.AddRoute(NewRoute().Path("/route/articles/").Method(HttpGet).Header(HttpAccept, "application/json"))
	router.AddRoute(NewRoute().Path("/route/articles/list").Method(HttpGet).Header(HttpAccept, "application/json"))

	req := createGetRequest("/route/articles/")
	req.Header.Add(HttpAccept.String(), "application/html")

	found := router.findRoute(createGetRequest("/route/article"))

	if found != nil {
		t.Errorf("Could not find route with correct contenttype.")
	}
}

func TestFindRouteWithPathPrefix(t *testing.T) {
	router := NewWebRouter()
	router.AddRoute(NewRoute().Path("/route/article/").PathPrefix("/css/").PathPrefix("/images/").Method(HttpGet))

	found := router.findRoute(createGetRequest("/route/article/css/mystylesheet.css"))

	if found == nil {
		t.Errorf("Could not find route with correct pathprefix.")
	}
}

func TestHTTPHandlerServiceWithId(t *testing.T) {
	id := "21"

	//The func validates the unit test.
	checkIdFunc := func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		id := params.Get("id")

		if id != id {
			t.Errorf("Could not find route with ID:" + id)
		}
	}

	resp := httptest.NewRecorder()
	req := createGetRequest("/route/articles/" + id)

	router := NewWebRouter()
	router.AddRoute(NewRoute().Path("/route/articles/{id:\\d*}").HandlerFunc(checkIdFunc))

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Handlerequest returned %v", resp.Code)
	}
}

//Helper methods
func createGetRequest(path string) *http.Request {
	req, _ := http.NewRequest(HttpGet.String(), path, nil)
	return req
}

func createPostRequest(path string) *http.Request {
	req := &http.Request{}
	req.URL, _ = url.Parse(path)
	req.Method = HttpPost.String()
	return req
}
