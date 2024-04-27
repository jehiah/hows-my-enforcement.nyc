package main

import (
	"io"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/jehiah/hows-my-enforcement.nyc/internal/apiresponse"
	log "github.com/sirupsen/logrus"
)

// SaveReportPreview saves a preview image to google storage
// PUT /{report}
func (a *App) SaveReportPreview(w http.ResponseWriter, r *http.Request) {
	id := ReportID(r.PathValue("report"))
	if !IsValidReportID(id) {
		apiresponse.NotFound404(w)
		return
	}
	ctx := r.Context()
	report, err := a.GetReport(ctx, id)
	if err != nil {
		log.WithField("reportID", id).Errorf("%#v", err)
		apiresponse.InternalError500(w)
		return
	}
	if report == nil {
		apiresponse.NotFound404(w)
		return
	}
	// don't allow overwrites unless it's stale
	if !a.devMode && !report.IsPreviewStale() {
		apiresponse.OK200(w, "OK")
		return
	}

	// TODO: verify token
	obj := a.storage.Object(id.String() + ".png")
	wc := obj.NewWriter(ctx)
	wc.ContentType = "image/png"                              // TODO get from request?
	_, err = io.Copy(wc, io.LimitReader(r.Body, 5*1024*1024)) // 5M
	if err != nil {
		log.WithField("reportID", id).Errorf("%#v", err)
		apiresponse.InternalError500(w)
		return
	}
	err = wc.Close()
	if err != nil {
		log.WithField("reportID", id).Errorf("%#v", err)
		apiresponse.InternalError500(w)
		return
	}

	_, err = a.firestore.Collection("reports").Doc(string(id)).Update(ctx,
		[]firestore.Update{
			{Path: "PreviewImage", Value: string(id) + ".png"},
			{Path: "PreviewLastUpdated", Value: time.Now()},
		})
	if err != nil {
		log.WithField("reportID", id).Errorf("%#v", err)
		apiresponse.InternalError500(w)
		return
	}

	apiresponse.OK200(w, "OK")
}
