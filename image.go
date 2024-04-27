package main

import (
	"fmt"
	"io"
	"net/http"

	"cloud.google.com/go/firestore"
	log "github.com/sirupsen/logrus"
)

func (a *App) SaveReportPreview(w http.ResponseWriter, r *http.Request) {
	id := ReportID(r.PathValue("report"))
	if !IsValidReportID(id) {
		http.NotFound(w, r)
		return
	}
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
	if !a.devMode && report.PreviewImage != "" {
		fmt.Fprintf(w, "NA")
		return
	}

	// TODO: verify token
	obj := a.storage.Object(id.String() + ".png")
	wc := obj.NewWriter(ctx)
	wc.ContentType = "image/png"                              // TODO get from request?
	_, err = io.Copy(wc, io.LimitReader(r.Body, 5*1024*1024)) // 5M
	if err != nil {
		log.WithField("reportID", id).Errorf("%#v", err)
		http.Error(w, "unexpected error", 500)
		return
	}
	err = wc.Close()
	if err != nil {
		log.WithField("reportID", id).Errorf("%#v", err)
		http.Error(w, "unexpected error", 500)
		return
	}

	_, err = a.firestore.Collection("reports").Doc(string(id)).Update(ctx,
		[]firestore.Update{{Path: "PreviewImage", Value: string(id) + ".png"}})
	if err != nil {
		log.WithField("reportID", id).Errorf("%#v", err)
		a.WebInternalError500(w, "")
		return
	}

	fmt.Fprintf(w, "OK")
}
