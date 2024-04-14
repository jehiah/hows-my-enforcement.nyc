package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"firebase.google.com/go/v4/auth"
	"github.com/gorilla/handlers"
	"github.com/jehiah/hows-my-enforcement.nyc/internal/account"
	"github.com/jehiah/hows-my-enforcement.nyc/internal/dthash"
	"github.com/jehiah/hows-my-enforcement.nyc/internal/nycapi"
	log "github.com/sirupsen/logrus"
)

//go:embed templates/*
var content embed.FS

//go:embed www/*
var static embed.FS

type App struct {
	devMode   bool
	firestore *firestore.Client
	firebase  *auth.Client
	storage   *storage.BucketHandle
	nycAPI    nycapi.Client

	staticHandler http.Handler
	templateFS    fs.FS
}

func newTemplate(fs fs.FS, n string) *template.Template {
	funcMap := template.FuncMap{
		// "Time":    humanize.Time,
	}
	t := template.New("empty").Funcs(funcMap)
	if n == "error.html" {
		return template.Must(t.ParseFS(fs, filepath.Join("templates", n)))
	}
	return template.Must(t.ParseFS(fs, filepath.Join("templates", n))) // , "templates/base.html"
}

// RobotsTXT renders /robots.txt
func (a *App) RobotsTXT(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain")
	a.addExpireHeaders(w, time.Hour*24*7)
	io.WriteString(w, `# robots welcome
# https://github.com/jehiah/hows-my-enforcement.nyc

User-agent: *
Disallow: /summons
`)

}

func (a *App) addExpireHeaders(w http.ResponseWriter, duration time.Duration) {
	if a.devMode {
		return
	}
	w.Header().Add("Cache-Control", fmt.Sprintf("public; max-age=%d", int(duration.Seconds())))
	w.Header().Add("Expires", time.Now().Add(duration).Format(http.TimeFormat))
}

func (a *App) WebInternalError500(w http.ResponseWriter, msg string) {
	if msg == "" {
		msg = "Server Error"
	}
	a.WebError(w, 500, msg)
}
func (a *App) WebPermissionError403(w http.ResponseWriter, msg string) {
	if msg == "" {
		msg = "Permission Denied"
	}
	a.WebError(w, 403, msg)
}
func (a *App) WebBadRequestError400(w http.ResponseWriter, msg string) {
	if msg == "" {
		msg = "Bad Request"
	}
	a.WebError(w, 400, msg)
}

func (a *App) WebError(w http.ResponseWriter, code int, msg string) {
	type Page struct {
		Title   string
		Code    int
		Message string
	}
	t := newTemplate(a.templateFS, "error.html")
	err := t.ExecuteTemplate(w, "error.html", Page{
		Title:   fmt.Sprintf("HTTP Error %d", code),
		Code:    code,
		Message: msg,
	})
	if err != nil {
		log.Errorf("%s", err)
	}
}

func (a *App) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var uid account.UID

	type Page struct {
		Reports []Report
	}
	body := Page{}

	var err error
	body.Reports, err = a.GetReports(ctx, uid)
	if err != nil {
		log.Print(err)
		// a.WebInternalError500(w, "")
		// return
	}

	t := newTemplate(a.templateFS, "index.html")
	err = t.ExecuteTemplate(w, "index.html", body)
	if err != nil {
		log.Print(err)
		a.WebInternalError500(w, "")
	}
}

func (a *App) IndexPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := a.User(r)
	if uid == "" {
		http.Redirect(w, r, "/", 302)
		return
	}
	defer r.Body.Close()
	report := Report{
		ID:      ReportID(time.Now().Format("2003-01-02-15-16")),
		UID:     uid,
		Created: time.Now().UTC(),
	}
	err := json.NewDecoder(r.Body).Decode(&report)
	if err != nil {
		log.Errorf("error decoding", err)
		a.WebError(w, 400, "invalid body")
		return
	}
	report.ID = ReportID(dthash.DateHash(time.Now()).String()[:3] + Randomness())

	// if !account.IsValidProfileID(profile.ID) {
	// 	log.WithField("uid", uid).Infof("profile ID %q is invalid", profile.ID)
	// 	http.Error(w, fmt.Sprintf("profile ID %q is invalid", profile.ID), 422)
	// 	return
	// }

	// if profile.Name == "" {
	// 	profile.Name = string(profile.ID)
	// }

	err = a.CreateReport(ctx, report)
	if err != nil {
		// duplicate?
		log.WithField("uid", uid).Warningf("%#v %s", err, err)
		http.Error(w, fmt.Sprintf("report %q is already taken", report.ID), 409)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct{ URL string }{URL: report.Link()})
	// http.Redirect(w, r, report.Link(), 302)
}

func (a *App) Report(w http.ResponseWriter, r *http.Request, id ReportID) {
	t := newTemplate(a.templateFS, "report.html")
	ctx := r.Context()
	report, err := a.GetReport(ctx, id)
	if err != nil {
		log.WithField("reportID", id).Errorf("%#v", err)
		a.WebInternalError500(w, "")
		return
	}
	if report == nil {
		http.Error(w, "Not Found", 404)
		return
	}

	type Page struct {
		Report      Report
		SavePreview bool
	}
	body := Page{
		Report:      *report,
		SavePreview: a.devMode || report.PreviewImage == "",
	}

	err = t.ExecuteTemplate(w, "report.html", body)
	if err != nil {
		log.WithField("report", id).Error(err)
		a.WebInternalError500(w, "")
	}
}

func (a *App) Precinct(w http.ResponseWriter, r *http.Request) {

	type Page struct {
		Report      Report
		SavePreview bool
	}
	body := Page{}

	t := newTemplate(a.templateFS, "precinct.html")
	err := t.ExecuteTemplate(w, "precinct.html", body)
	if err != nil {
		log.Print(err)
		a.WebInternalError500(w, "")
	}
}

func (a *App) Summons(w http.ResponseWriter, r *http.Request) {

	type Page struct {
		SummonsNumber int64
	}
	body := Page{}
	r.ParseForm()
	var err error
	body.SummonsNumber, err = strconv.ParseInt(r.Form.Get("number"), 10, 64)
	// nypd summonses are ~~ 1493972194
	// nypd traffic summonses are ~~ 9031199333
	if err != nil || body.SummonsNumber < 1000000000 || body.SummonsNumber > 10000000000 {
		a.WebBadRequestError400(w, "")
		return
	}

	t := newTemplate(a.templateFS, "summons.html")
	err = t.ExecuteTemplate(w, "summons.html", body)
	if err != nil {
		log.Print(err)
		a.WebInternalError500(w, "")
	}
}

func (app *App) User(*http.Request) account.UID {
	return account.UID("test")
}

func (app App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		switch r.URL.Path {
		case "/":
			app.Index(w, r)
			return
		case "/precinct":
			app.Precinct(w, r)
			return
		case "/summons":
			app.Summons(w, r)
			return
		case "/311":
			app.Lookup311(w, r)
			return
		case "/robots.txt":
			app.RobotsTXT(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/static/") {
			app.staticHandler.ServeHTTP(w, r)
			return
		}
		if p := ReportID(strings.TrimPrefix(r.URL.Path, "/")); IsValidReportID(p) {
			app.Report(w, r, p)
			return
		}
	case "POST":
		switch r.URL.Path {
		case "/":
			app.IndexPost(w, r)
			return
		case "/data/311":
			app.Lookup311Post(w, r)
			return
		}
	case "PUT":
		if p := ReportID(strings.TrimPrefix(r.URL.Path, "/")); IsValidReportID(p) {
			app.SaveReportPreview(w, r, p)
			return
		}
	default:
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
		return
	}
	http.NotFound(w, r)
}

// tsFmt is used to match logrus timestamp format
// w/ our stdlib log fmt (Ldate | Ltime)
const tsFmt = "2006/01/02 15:04:05"

func main() {
	logRequests := flag.Bool("log-requests", false, "log requests")
	devMode := flag.Bool("dev-mode", false, "development mode")
	flag.Parse()
	log.SetReportCaller(true)
	if *devMode {
		log.SetFormatter(&log.TextFormatter{TimestampFormat: tsFmt, FullTimestamp: true})
	} else {
		log.SetFormatter(&fluentdFormatter{})
	}

	log.Print("starting server...")
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Panic(err)
	}
	app := &App{
		devMode:    *devMode,
		firestore:  createClient(ctx),
		templateFS: content,
		storage:    client.Bucket("hows-my-enforcement-nyc"),
		nycAPI: nycapi.Client{
			SubscriptionKey: os.Getenv("OCP_APM_KEY"),
		},
	}
	if *devMode {
		app.templateFS = os.DirFS(".")
		app.staticHandler = http.StripPrefix("/static/", http.FileServer(http.Dir("www")))
	} else {
		d, err := fs.Sub(static, "www")
		if err != nil {
			log.Panic(err)
		}
		app.staticHandler = http.StripPrefix("/static/", http.FileServer(http.FS(d)))
	}

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	var h http.Handler = app
	if *logRequests {
		h = handlers.LoggingHandler(os.Stdout, h)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, h); err != nil {
		log.Fatal(err)
	}
}
