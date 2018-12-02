package main

// код писать тут
import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"
)

type Dataset struct {
	Users []UserWithTags `xml:"row"`
}

type UserWithTags struct {
	Id     int    `xml:"id" json:"id"`
	Name   string `xml:"first_name" json:"name"`
	Age    int    `xml:"age" json:"age"`
	About  string `xml:"about" json:"about"`
	Gender string `xml:"gender" json:"gender"`
}

var dataset Dataset

func init() {
	xmlText, _ := ioutil.ReadFile("dataset.xml")
	xml.Unmarshal([]byte(xmlText), &dataset)
}

func SearchServerDummy(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("AccessToken") {
	case "TimeoutError":
		w.WriteHeader(http.StatusFound)
		time.Sleep(time.Second * 2)
	case "InvalidJson":
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `invaalid json a;lkd;lkdj`)
	case "InternalServerError":
		w.WriteHeader(http.StatusInternalServerError)
	case "BadRequest":
		w.WriteHeader(http.StatusBadRequest)
	case "ErrorBadOrderField":
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error":"ErrorBadOrderField"}`)
	case "ErrorBadOrderUnknown":
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error":"ErrorBadOrderUnknown"}`)
	case "StatusUnauthorized":
		w.WriteHeader(http.StatusUnauthorized)
	default:
		limit, _ := strconv.Atoi(r.FormValue("limit"))
		offset, _ := strconv.Atoi(r.FormValue("offset"))

		w.WriteHeader(http.StatusOK)

		if limit > 25 {
			limit = 25
		}
		startRow := offset
		if startRow > len(dataset.Users) {
			startRow = len(dataset.Users)
		}
		endRow := offset + limit
		if endRow > len(dataset.Users) {
			endRow = len(dataset.Users)
		}
		body, _ := json.Marshal(dataset.Users[startRow:endRow])
		w.Write(body)
	}

}

func TestFindUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerDummy))
	defer ts.Close()

	cases := []struct {
		client SearchClient
		req    SearchRequest
		resp   *SearchResponse
		err    bool
	}{
		{
			client: SearchClient{
				URL:         ts.URL,
				AccessToken: "ACCESS_TOKEN",
			},
			req: SearchRequest{Limit: 1, Offset: 0},
			resp: &SearchResponse{
				Users:    []User{User(dataset.Users[0])},
				NextPage: true,
			},
			err: false,
		},
		{
			client: SearchClient{
				URL:         ts.URL,
				AccessToken: "ACCESS_TOKEN",
			},
			req: SearchRequest{Offset: 30, Limit: 26},
			resp: &SearchResponse{
				Users: func(in []UserWithTags) []User {
					var out []User
					for _, v := range in {
						out = append(out, User(v))
					}
					return out
				}(dataset.Users[30:35]),
				NextPage: false,
			},
			err: false,
		},
		{
			client: SearchClient{
				URL:         ts.URL,
				AccessToken: "ACCESS_TOKEN",
			},
			req: SearchRequest{Limit: -1},
			err: true,
		},
		{
			client: SearchClient{
				URL:         ts.URL,
				AccessToken: "ACCESS_TOKEN",
			},
			req: SearchRequest{Offset: -1},
			err: true,
		},
		{
			client: SearchClient{
				URL:         ts.URL,
				AccessToken: "TimeoutError",
			},
			err: true,
		},
		{
			client: SearchClient{
				URL:         ts.URL,
				AccessToken: "InvalidJson",
			},
			req: SearchRequest{Limit: 1, Offset: 0},
			err: true,
		},
		{
			client: SearchClient{
				URL:         ts.URL,
				AccessToken: "InternalServerError",
			},
			req: SearchRequest{Limit: 1, Offset: 0},
			err: true,
		},
		{
			client: SearchClient{
				URL:         ts.URL,
				AccessToken: "BadRequest",
			},
			req: SearchRequest{Limit: 1, Offset: 0},
			err: true,
		},
		{
			client: SearchClient{
				URL:         ts.URL,
				AccessToken: "ErrorBadOrderField",
			},
			req: SearchRequest{Limit: 1, Offset: 0},
			err: true,
		},
		{
			client: SearchClient{
				URL:         ts.URL,
				AccessToken: "ErrorBadOrderUnknown",
			},
			req: SearchRequest{Limit: 1, Offset: 0},
			err: true,
		},
		{
			client: SearchClient{
				URL:         ts.URL + "sdlkfjsdlkj",
				AccessToken: "UnknownError",
			},
			req: SearchRequest{Limit: 1, Offset: 0},
			err: true,
		},
		{
			client: SearchClient{
				URL:         ts.URL,
				AccessToken: "StatusUnauthorized",
			},
			req: SearchRequest{Limit: 1, Offset: 0},
			err: true,
		},
	}

	for id, tt := range cases {
		resp, err := tt.client.FindUsers(tt.req)
		if tt.err && err == nil {
			t.Fatalf("[%d] %v \n!=\n %v", id, err, tt.err)
		}
		if !reflect.DeepEqual(resp, tt.resp) {
			t.Fatalf("[%d] %#v \n!=\n %#v", id, resp, tt.resp)
		}
	}
}
