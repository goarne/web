package web

type HttpHeader string

func (h HttpHeader) String() string {
	return string(h)
}

const (
	HttpAccept      HttpHeader = "Accept"
	HttpContentType HttpHeader = "Content-Type"
)

type HttpMethod string

func (m HttpMethod) String() string {
	return string(m)
}

const (
	HttpPost    HttpMethod = "POST"
	HttpGet     HttpMethod = "GET"
	HttpPut     HttpMethod = "PUT"
	HttpDelete  HttpMethod = "DELETE"
	HttpOptions HttpMethod = "OPTIONS"
	HttpTrace   HttpMethod = "TRACE"
	HttpConnect HttpMethod = "CONNECT"
)
