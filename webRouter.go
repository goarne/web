//WebRouter and Route for routing http requests .
package web

import (
	"net/http"
	"regexp"
	"strings"
)

//Regexp constants used to build the URL path pattern for matching incomming http requests url.
const REGEX_START_OF_STRING string = "\\A"
const REGEX_END_OF_STRING string = "\\z"
const REGEX_URL_VARIABLE string = "{[^}]*?}"
const REGEX_URL_VARIABLE_DEF_PATTERN string = "\\w*"

//The WebRouter serves as a middleware router for http requests.
//It routes webrequests to the correct Handler based on path, method and headers.
//
//	Sample:
//		router := web.NewWebRouter()
//
//		r := web.NewRoute()
//		r.Path("/sample/")
//		r.Method("GET")
//		r.Header("Accept", "html")
//		r.Handler(&SampleHtmlHandler{})
//
//	 	router.AddRoute(r)
type WebRouter struct {
	routes []*Route
}

//Creates a new instance of a webrouter.
func NewWebRouter() *WebRouter {
	return &WebRouter{[]*Route{}}
}

//The method called by the HTTP server when receiving new requests.
//It implements the http.Handler ServeHTTP function.
func (wr *WebRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p := req.URL.Path

	r := wr.findRoute(req)

	if r == nil {
		http.Error(w, "Could not find route for: "+p, 404)
		return
	}

	r.handler.ServeHTTP(w, req)
}

//Adds a new route to the router.
func (wr *WebRouter) AddRoute(r *Route) *WebRouter {
	wr.routes = append(wr.routes, r)

	return wr
}

//Finds a route which matches the incomming http request.
func (wr *WebRouter) findRoute(req *http.Request) *Route {

	for _, v := range wr.routes {

		if v.match(req) {
			return v
		}
	}

	return nil
}

//Return the number of routes in the webrouter.
func (wr *WebRouter) routeCount() int {
	return len(wr.routes)
}

//Route is an instance of an route endpoint for a http request.
//It implements logic to match a http requst on path, method and headers, and
//holds a reference to the http.Handler which implements the business logic.
//If incomming request matches the route, the route passes the request to its
//referenced http.handler.

//Sample
//	r := web.NewRoute()
//	r.Path("/sample/{id}")
//	r.Method("GET")
//	r.Header("Accept", "html")
//	r.Handler(&SampleHtmlHandler{})

type Route struct {
	//Supported HTTP methods
	methods map[string]bool

	//Url to resource request should be mapped to.
	urlMatcher *regexp.Regexp

	//Variables allowed in request, with corresponding regexp pattern.
	urlVariables map[string]string

	//Root path to resource. Used to build pathPrefix.
	rootPath string

	//Url to resource request should be mapped to.
	pathPrefixMatchers []*regexp.Regexp

	//Headers acceptet by resource.
	headers map[string]*regexp.Regexp

	//Handler instance which conforms to http.Handler interface
	handler http.Handler
}

//Creates a new instance of a Route. It must be further built with matcher
//rules and a http.Handler for serving the response.
func NewRoute() *Route {
	r := &Route{}
	r.methods = make(map[string]bool)
	r.headers = make(map[string]*regexp.Regexp)
	r.urlVariables = make(map[string]string)

	return r
}

//Method adds a path to the route which is used by the router to match
//the incomming http request. The match is done with regular expressions.
//
//Sample path 1:	/example
//The path given requires an exact match in order to be successfull and
//will only match a requested path with /path
//
//Samplepath 2: 	/example/{id}
//The path can build in variables aswell with the convention {variable}.
//The route will match the http request path http://example.url/1abc
//It will not match a request with pat http://example.url/1abc/
//
//Samplepath 3:		/example/{id:\\d*}/
//The path variables can be defined with the regex rule that requires the
//id to be a digit in the format {variablename:regexpattern}
//The route will match the http request path http://example.url/1
//It will not match a request with pat http://example.url/1/
func (r *Route) Path(path string) *Route {
	r.rootPath = path

	r.buildPathPattern()

	return r
}

//Method adds a pathprefix support to the routes path.
//This adds match attributes for the route.
//
//Samplepath:	/example/
//The path will match the path /example/xxxx
func (r *Route) PathPrefix(path string) *Route {
	pathPrefix := strings.Replace(r.rootPath+path, "//", "/", -1)
	r.pathPrefixMatchers = append(r.pathPrefixMatchers, regexp.MustCompile(pathPrefix))
	return r
}

//Builder method for adding supported methods.
//This adds match attributes for the route.
//It support all http verbs created.
//
//Samplemethod: "POST"
func (r *Route) Method(m string) *Route {

	_, exists := r.methods[strings.ToUpper(m)]

	if !exists {
		r.methods[strings.ToUpper(m)] = true
	}

	return r
}

//Builder method for adding required headers.
//This adds match attributes for the route.
//
//Samleheader:  key "Accept", value "application/json")
//The header will match the request with  a similar http header,
//if its not present the request will not match.
func (r *Route) Header(k, v string) *Route {
	_, exists := r.headers[k]

	if !exists {
		r.headers[k] = regexp.MustCompile(v)
	}

	return r
}

//Method for setting handler.
func (r *Route) Handler(h http.Handler) *Route {
	r.handler = h

	return r
}

//Method for setting a handler function.
func (r *Route) HandlerFunc(f func(http.ResponseWriter, *http.Request)) *Route {
	r.handler = http.HandlerFunc(f)

	return r
}

//The main match method for the route.
//It checks wether a request matches the routes definition.
//It checks the abolute path, then substringpath, method and then headers.
//The path, method and header are matched as &&.
//If returns true if route matches, otherwise false.
func (r *Route) match(req *http.Request) bool {

	rootPath := req.URL.Path

	if !r.urlMatcher.Match([]byte(rootPath)) {
		if !r.matchPathPrefix(req) {
			return false
		}
	}

	if (len(r.methods) > 0) && (!r.methods[req.Method]) {
		return false
	}

	for k, v := range r.headers {
		reqHeader := req.Header.Get(k)
		if !v.Match([]byte(reqHeader)) {
			return false
		}
	}

	r.parseVariables(req)

	return true
}

//Matches the substring of a path.
func (r *Route) matchPathPrefix(req *http.Request) bool {

	pathPrefix := req.URL.Path

	for _, m := range r.pathPrefixMatchers {
		if m.Match([]byte(pathPrefix)) {
			return true
		}
	}

	return false
}

//The method parses the variables from the http request
//which maches the named regexp pattern.
func (r *Route) parseVariables(req *http.Request) {

	match := r.urlMatcher.FindStringSubmatch(req.URL.Path)
	if match == nil {
		return
	}
	params := req.URL.Query()

	for i, name := range r.urlMatcher.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}

		r.urlVariables[name] = match[i]
		params.Add(name, match[i])
	}

	req.URL.RawQuery = params.Encode()
}

//The function builds the regex pattern based on the pathed received.
//First rule is that the path is post and prefixed with regex directives
//that the path must match from start to end.
//The second rule is to handle variables in the path.
//By default variables are handled as strings, but its possible to replace
//the default regex by defining it in the format {variablename:regexpattern}
//The incomming path can be a simple path or contain variables with regex patterns.
//All patterns result in a pattern which must have an absolute match.
//
//The path definition "http://example.url/{id:\\d*}/"
//will look like "\\Ahttp://example.url/(?P<id>\\d*)/\\z"

func (r *Route) buildPathPattern() {
	path := r.rootPath
	varsRegexp := regexp.MustCompile(REGEX_URL_VARIABLE)
	vars := varsRegexp.FindAllString(path, -1)

	for _, v := range vars {
		variable := v[1 : len(v)-1]

		variableContent := strings.Split(variable, ":")
		namedPattern := "(?P<" + variableContent[0] + ">"
		variablePattern := REGEX_URL_VARIABLE_DEF_PATTERN

		if len(variableContent) == 2 {
			variablePattern = variableContent[1]
		}
		namedPattern = namedPattern + variablePattern + ")"

		path = strings.Replace(path, v, namedPattern, 1)
	}

	pattern := REGEX_START_OF_STRING + path + REGEX_END_OF_STRING

	r.urlMatcher = regexp.MustCompile(pattern)
}
