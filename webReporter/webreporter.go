package webreporter

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/nikolainp/TLogViewer/storage"
)

type WebReporter struct {
	storage   *storage.Storage
	srv       http.Server
	templates *template.Template

	title     string
	filter    *dataFilter
	navigator navigation

	port int

	cancelChan chan bool
}

//go:embed static
var staticContent embed.FS

//go:embed templates
var templateContent embed.FS

func New(storage *storage.Storage, isCancelChan chan bool) *WebReporter {
	var err error

	obj := new(WebReporter)

	obj.port = 8090

	obj.storage = storage
	obj.srv = http.Server{
		Handler: logging(log.New(os.Stdout, "http: ", log.LstdFlags))(obj.getHandlers()),
		Addr:    fmt.Sprintf(":%d", obj.port),
	}
	obj.templates, err = template.ParseFS(templateContent, "templates/*.gohtml")
	checkErr(err)

	details := obj.getRootDetails()
	obj.title = details.Title
	obj.filter = getDataFilter(obj.templates.Lookup("dataFilter.gohtml"))
	obj.filter.setTime(details.FirstEventTime, details.LastEventTime)

	obj.cancelChan = isCancelChan

	return obj
}

func (obj *WebReporter) Start() error {
	log.Printf("start web-server, port: %d\n", obj.port)

	done := make(chan bool)

	go func() {
		<-obj.cancelChan

		log.Println("web-server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		obj.srv.SetKeepAlivesEnabled(false)
		if err := obj.srv.Shutdown(ctx); err != nil {
			log.Fatalf("could not gracefully shutdown the web-server: %v\n", err)
		}
		close(done)
	}()

	err := obj.srv.ListenAndServe()
	if err != http.ErrServerClosed {
		return err
	}

	<-done
	log.Println("server stopped")
	return nil
}

///////////////////////////////////////////////////////////////////////////////

func (obj *WebReporter) getHandlers() *http.ServeMux {

	sm := http.NewServeMux()

	sm.HandleFunc("/", obj.rootPage)
	sm.HandleFunc("/processes", obj.processes)
	sm.HandleFunc("/performance", obj.performance)
	sm.HandleFunc("/performance/{id}", obj.performance)

	sm.HandleFunc("/datafilter", obj.filter.setContext)

	sm.Handle("/static/", http.FileServer(http.FS(staticContent)))
	sm.HandleFunc("/headers", obj.headers)

	//logger.Printf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	return sm
}

func (obj *WebReporter) headers(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////

var checkErr = func(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				logger.Println(r.Method, r.URL.Path)
			}()
			next.ServeHTTP(w, r)
		})
	}
}
