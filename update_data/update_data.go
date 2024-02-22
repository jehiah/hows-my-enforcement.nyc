package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/storage"
	"github.com/jehiah/hows-my-enforcement.nyc/internal/dthash"
	"google.golang.org/api/iterator"
)

type Row struct {
	County           string
	Street           string
	Intersections    []string
	Agencies         []string
	NumberViolations int
}

func Headers(a any) []string {
	var fields []string
	t := reflect.TypeOf(a)
	for i := 0; i < t.NumField(); i++ {
		fields = append(fields, t.Field(i).Name)
	}
	return fields
}

func (r Row) CSV() []string {
	return []string{
		r.Street,
		strings.Join(r.Intersections, "|"),
		strings.Join(r.Agencies, "|"),
		fmt.Sprintf("%d", r.NumberViolations),
	}
}

func main() {
	// Create a BigQuery client
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "hows-my-enforcement-nyc")
	if err != nil {
		log.Fatal(err)
	}

	storageClient, err := storage.NewClient(ctx)
	bkt := storageClient.Bucket("hows-my-enforcement-nyc")
	err = Query(ctx, client, bkt)
	if err != nil {
		log.Fatal(err)
	}
}

func Query(ctx context.Context, bq *bigquery.Client, gs *storage.BucketHandle) error {
	query := `
	SELECT
CASE
  WHEN violation_county in ("MN", "NY") THEN "ny"
  WHEN violation_county in ("BX", "Bronx") THEN "bx"
  WHEN violation_county in ("BK", "K", "Kings") THEN "bk"
  WHEN violation_county in ("Q", "QN", "Qns") THEN "qn"
  WHEN violation_county in ("R", "Rich", "ST") THEN "si"
  ELSE "Unknown"
END as County,
  street_name as Street,
  ARRAY_AGG(distinct regexp_replace(LOWER(intersecting_street), "((^[^/]+$)|(.+ (.+/of .*)))", "\\2\\4") IGNORE NULLS ) as Intersections,
  ARRAY_AGG(distinct issuing_agency) as Agencies,
  count(*) as NumberViolations
FROM ` + "`jehiah-p1.opendata.parking_violations_fy`" + `
WHERE 
issuing_agency in ('T','P','S')
and issue_date is not null
and issue_date between DATE_ADD(CURRENT_DATE, INTERVAL -13 MONTH) and CURRENT_DATE
and street_name is not null
and violation_county is not null
and violation_county in ("MN", "NY", "BX", "Bronx", "BK", "K", "Kings", "Q", "QN", "Qns", "R", "Rich", "ST")
group by 1, 2
-- having number_violations > 1
order by 1, lower(street_name)
`

	q := bq.Query(query)

	// Wait for the query job to finish
	it, err := q.Read(ctx)
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now()
	// now.Format("20060102")
	// suffix := uuid.Must(uuid.NewRandom()).String()[:5]
	files := make(map[string]*csv.Writer)
	fmt.Printf("const DatasetID = %q\n", dthash.DateHash(now))

	for _, county := range []string{"bk", "bx", "ny", "qn", "si"} {
		fn := fmt.Sprintf("fy_2024_%s_%s_intersecting_streets.tsv", dthash.DateHash(now), county)
		// log.Printf("filename %s", fn)
		obj := gs.Object(fmt.Sprintf("data/%s", fn))
		log.Printf("https://storage.googleapis.com/%s/%s", obj.BucketName(), obj.ObjectName())
		w := obj.NewWriter(ctx)
		// w.ACL = append(w.ACL, storage.ACLRule{
		// 	Role:   storage.RoleReader,
		// 	Entity: storage.AllUsers,
		// })
		defer func() {
			err = w.Close()
			if err != nil {
				log.Printf("error %s", err)
			}
		}()

		csvWriter := csv.NewWriter(w)
		csvWriter.Comma = '\t'
		defer csvWriter.Flush()

		csvWriter.Write(Headers(Row{})[1:])
		files[county] = csvWriter
	}

	var rowCount = 0
	for {
		var row Row
		err := it.Next(&row)
		if err == iterator.Done { // from "google.golang.org/api/iterator"
			break
		}
		rowCount++
		if err != nil {
			log.Fatal(err)
		}

		err = files[row.County].Write(row.CSV())
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("rowCount %d", rowCount)
	return nil
}
