# gopkgapi

## Record your Go package's public API

`gopkgapi` is a command line tool for generating API descriptions for a package or set of packages. The tool is inspired by Go's internal checker for ensuring the [Go 1 compatibility promise](https://golang.org/doc/go1compat) across the standard library. It spits out a similar format as the [Go API documents](https://github.com/golang/go/tree/master/api), one "feature" per line describing all exported types. 

The tool aims to allow other projects to check-in similar documents, and record exactly when package API are added, removed, or changed.

Download using `go get`.

```
go get github.com/ericchiang/gopkgapi
```

Then use the tool on any package or set of packages in your GOPATH or GOROOT.

```
$ gopkgapi net/http/httptest
pkg net/http/httptest, const DefaultRemoteAddr string
pkg net/http/httptest, func NewRecorder() *ResponseRecorder
pkg net/http/httptest, func NewRequest(string, string, io.Reader) *net/http.Request
pkg net/http/httptest, func NewServer(net/http.Handler) *Server
pkg net/http/httptest, func NewTLSServer(net/http.Handler) *Server
pkg net/http/httptest, func NewUnstartedServer(net/http.Handler) *Server
pkg net/http/httptest, method (*ResponseRecorder) Flush()
pkg net/http/httptest, method (*ResponseRecorder) Header() net/http.Header
pkg net/http/httptest, method (*ResponseRecorder) Result() *net/http.Response
pkg net/http/httptest, method (*ResponseRecorder) Write([]byte) (int, error)
pkg net/http/httptest, method (*ResponseRecorder) WriteHeader(int)
pkg net/http/httptest, method (*ResponseRecorder) WriteString(string) (int, error)
pkg net/http/httptest, method (*Server) Close()
pkg net/http/httptest, method (*Server) CloseClientConnections()
pkg net/http/httptest, method (*Server) Start()
pkg net/http/httptest, method (*Server) StartTLS()
pkg net/http/httptest, type ResponseRecorder struct
pkg net/http/httptest, type ResponseRecorder struct, Body *bytes.Buffer
pkg net/http/httptest, type ResponseRecorder struct, Code int
pkg net/http/httptest, type ResponseRecorder struct, Flushed bool
pkg net/http/httptest, type ResponseRecorder struct, HeaderMap net/http.Header
pkg net/http/httptest, type Server struct
pkg net/http/httptest, type Server struct, Config *net/http.Server
pkg net/http/httptest, type Server struct, Listener net.Listener
pkg net/http/httptest, type Server struct, TLS *crypto/tls.Config
pkg net/http/httptest, type Server struct, URL string
```

There are small differences from Go's `api/` directory format:

* Imports of types are represented by absolute path. For example, `*net/http.Request` instead of `*http.Request`.
* Constants display their type, not their value.
