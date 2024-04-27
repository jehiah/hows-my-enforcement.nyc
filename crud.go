package main

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/jehiah/hows-my-enforcement.nyc/internal/account"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
)

type Report struct {
	ID                   ReportID
	FiscalYear           string
	Borough              string
	Agency               string
	Streets              []string
	HouseStart, HouseEnd int
	Intersections        []string
	UID                  account.UID
	Created              time.Time
	PreviewImage         string
	PreviewLastUpdated   time.Time
}
type ReportID string

func (id ReportID) String() string {
	return string(id)
}

func IsValidReportID(r ReportID) bool {
	return len(string(r)) > 1
}

func (r Report) Link() string {
	return "/" + string(r.ID)
}

func (r Report) IsPreviewStale() bool {
	switch {
	case r.PreviewImage == "":
		return true
	case r.PreviewLastUpdated.IsZero():
		return true
	case r.PreviewLastUpdated.Unix() == 0:
		return true
	case time.Since(r.PreviewLastUpdated) > time.Hour*24*30:
		return true
	default:
		return false
	}
}

func (r Report) PublicPreview() string {
	if r.PreviewImage == "" {
		return ""
	}
	return "https://storage.googleapis.com/hows-my-enforcement-nyc/" + r.PreviewImage
}

func (r Report) Description() string {
	if len(r.Streets) == 0 {
		return ""
	}
	return fmt.Sprintf("Violations on %d-%d %s", r.HouseStart, r.HouseEnd, r.Streets[0])
}

func (a *App) GetReport(ctx context.Context, ID ReportID) (*Report, error) {
	if !IsValidReportID(ID) {
		return nil, nil
	}
	dsnap, err := a.firestore.Collection("reports").Doc(string(ID)).Get(ctx)
	if err != nil {
		if IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if !dsnap.Exists() {
		return nil, nil
	}
	var p Report
	err = dsnap.DataTo(&p)
	if len(p.Intersections) == 0 {
		p.Intersections = make([]string, 0)
	}
	return &p, err
}

func (a *App) GetReports(ctx context.Context, UID account.UID) ([]Report, error) {
	query := a.firestore.Collection("reports").Where("UID", "==", string(UID)).OrderBy("Created", firestore.Asc).Limit(100)
	// ref := a.firestore.Collection(fmt.Sprintf("users/%s/profiles", UID))
	iter := query.Documents(ctx)
	defer iter.Stop()
	var out []Report
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var r Report
		err = doc.DataTo(&r)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func (a *App) CreateReport(ctx context.Context, r Report) error {
	r.Created = time.Now().UTC()
	log.Printf("creating report %#v", r)
	_, err := a.firestore.Collection("reports").Doc(string(r.ID)).Create(ctx, r)
	return err
}
