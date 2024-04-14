package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

func (a *App) Lookup311(w http.ResponseWriter, r *http.Request) {

	type Page struct {
	}
	body := Page{}
	r.ParseForm()
	var err error
	t := newTemplate(a.templateFS, "311.html")
	err = t.ExecuteTemplate(w, "311.html", body)
	if err != nil {
		log.Error(err)
		a.WebInternalError500(w, "")
	}
}

type LookupRequest struct {
	SRNumber string `json:"sr_number"`
}

type SRResponse struct {
	SRNumber       string `json:"SRNumber"`
	SRID           string `json:"SRID"`
	Link           string `json:"Link"`
	Agency         string `json:"Agency"`
	Problem        string `json:"Problem"`
	ProblemDetails string `json:"ProblemDetails"`
	Address        string `json:"Address"`
	Description    string `json:"Description"`

	Status                      string `json:"Status"`
	ResolutionActionUpdatedDate string `json:"ResolutionActionUpdatedDate"`
	DateTimeSubmitted           string `json:"DateTimeSubmitted"`
	ResolutionAction            string `json:"ResolutionAction"`

	NotesToCustomer string `json:"NotesToCustomer"`
}

func (a *App) Lookup311Post(w http.ResponseWriter, r *http.Request) {
	var lookupRequest LookupRequest
	err := json.NewDecoder(r.Body).Decode(&lookupRequest)
	if err != nil {
		log.Error(err)
		a.WebInternalError500(w, "")
		return
	}

	req, err := http.NewRequest("GET", "https://portal.311.nyc.gov/api-get-sr-or-correspondence-by-number/?number="+lookupRequest.SRNumber, nil)
	if err != nil {
		log.Error(err)
		a.WebInternalError500(w, "")
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(err)
		a.WebInternalError500(w, "")
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		a.WebInternalError500(w, "")
		return
	}
	body = bytes.Replace(body, []byte(`\r`), []byte(`\r\n`), -1)
	type expectedResponse struct {
		SRID string `json:"srid"`
	}
	var expected expectedResponse
	err = json.Unmarshal(body, &expected)
	if err != nil {
		log.Errorf("error %s unmarshaling %q", err, body)
		a.WebInternalError500(w, "")
		return
	}

	log.WithField("srid", expected.SRID).Info("detected srid")
	srDetail, err := FetchSRDetailWeb(expected.SRID)
	if err != nil {
		log.Error(err)
		a.WebInternalError500(w, "")
		return
	}
	log.Printf("%#v", srDetail)

	srAPI, err := a.nycAPI.GetServiceRequest(r.Context(), lookupRequest.SRNumber)
	if err != nil {
		log.Error(err)
		a.WebInternalError500(w, "")
		return
	}

	respBody, _ := json.Marshal(SRResponse{
		SRNumber:       srAPI.SRNumber,
		SRID:           expected.SRID,
		Link:           srWebLink(expected.SRID),
		Agency:         srAPI.Agency,
		Problem:        srAPI.Problem,
		ProblemDetails: srAPI.ProblemDetails,
		Address:        srAPI.Address.FullAddress,

		Status:                      srAPI.Status,
		ResolutionActionUpdatedDate: srAPI.ResolutionActionUpdatedDate,
		DateTimeSubmitted:           srAPI.DateTimeSubmitted,
		ResolutionAction:            srAPI.ResolutionAction,

		Description:     srDetail.Description,
		NotesToCustomer: srDetail.NotesToCustomer,
	})
	log.Printf("response %s", respBody)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(respBody))
}

type SRDetailWeb struct {
	SRNumber        string `html_value:"n311_name"`
	Description     string `html_content:"n311_description"`
	NotesToCustomer string `html_content:"n311_notestocustomertext"`
	// Problem       string `html_value:"n311_problemid_label"`
	// ProblemDetail string `html_value:"n311_problemdetailid_label"`
	// Address       string `html_value:"n311_addressid_label"`
	Submitted string `html_value:"n311_datetimesubmitted"`
	Closed    string `html_value:"n311_datetimeclosed"`
}

func srWebLink(srid string) string {
	return "https://portal.311.nyc.gov/sr-details/?id=" + url.QueryEscape(srid)
}

// FetchSRDetailWeb fetches the "SR Lookup" online
// because it is the only location that contains the submission description
// and any response comments
func FetchSRDetailWeb(srid string) (*SRDetailWeb, error) {
	req, _ := http.NewRequest("GET", srWebLink(srid), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	srDetail, err := ParseSRDetailWeb(resp.Body)
	return &srDetail, err
}

func ParseSRDetailWeb(body io.Reader) (SRDetailWeb, error) {
	var srDetail SRDetailWeb
	err := TokenParser(body, &srDetail)
	return srDetail, err
}

// TokenParser parses HTML tokens into a struct based on struct tags identifying IDs
func TokenParser(r io.Reader, s interface{}) error {
	z := html.NewTokenizer(r)
	value := reflect.ValueOf(s).Elem()

	var currentField reflect.Value
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				return nil
			}
			return z.Err()
		case html.StartTagToken, html.SelfClosingTagToken:
			token := z.Token()
			tokenID := tokenAttr(token, "id")
			tokenValue := strings.TrimSpace(tokenAttr(token, "value"))
			if tokenID == "" {
				continue
			}
			// log.Printf("id %q token value %q %#v", tokenID, tokenValue, token)
			for i := 0; i < value.NumField(); i++ {
				field := value.Type().Field(i)
				idTag := field.Tag.Get("html_content")
				if idTag == tokenID {
					// log.WithField("tokenID", tokenID).Infof("found content field %s %s ", idTag, field.Name)
					currentField = value.Field(i)
					break
				}
			}
			if tokenValue != "" {
				for i := 0; i < value.NumField(); i++ {
					field := value.Type().Field(i)
					idTag := field.Tag.Get("html_value")
					if idTag == tokenID {
						// log.Infof("found value field %s %s value %q", idTag, field.Name, tokenValue)
						value.Field(i).SetString(tokenValue)
						break
					}
				}
			}
		case html.TextToken:
			if currentField.IsValid() && currentField.CanSet() {
				currentField.SetString(strings.TrimSpace(string(z.Text())))
				currentField = reflect.Value{}
			}
		}
	}
}

func tokenAttr(t html.Token, key string) string {
	for _, a := range t.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}
