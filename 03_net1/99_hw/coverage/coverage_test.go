package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSearchUser(t *testing.T) {

	testCases := []struct {
		nameCase     string
		limit        int
		offset       int
		query        string
		orderField   string
		orderBy      int
		ExpectedErr  error
		expectedIDs  []int
		expectedNext bool
	}{
		{
			"Success",           // test Name
			5, 0, "", "Name", 1, // SearchRequest{limit, offset, query, orderField, orderBy}
			nil,                      // FindUsers - Err
			[]int{15, 16, 19, 22, 5}, // resp.Users - ID
			true,                     // nextPage
		},
		{
			"Query_match",        // test Name
			5, 0, "on", "Age", 1, // SearchRequest{limit, offset, query, orderField, orderBy}
			nil,                    // FindUsers - Err
			[]int{1, 15, 0, 14, 2}, // resp.Users - ID
			true,                   // nextPage
		},
		{
			"Sort_DESC",      // test Name
			6, 9, "", "", -1, // SearchRequest{limit, offset, query, orderField, orderBy}
			nil,                         // FindUsers - Err
			[]int{20, 7, 25, 34, 21, 6}, // resp.Users - ID
			true,                        // nextPage
		},
		{
			"Next_page",        // test Name
			4, 34, "", "Id", 1, // SearchRequest{limit, offset, query, orderField, orderBy}
			nil,       // FindUsers - Err
			[]int{34}, // resp.Users - ID
			false,     // nextPage
		},
		{
			"Bad_order_field",   // test Name
			5, 0, "", "test", 1, // SearchRequest{limit, offset, query, orderField, orderBy}
			fmt.Errorf("OrderField test invalid"), // FindUsers - Err
			[]int{},                               // resp.Users - ID
			true,                                  // nextPage
		},
		{
			"Limit_over_25",    // test Name
			30, 0, "", "Id", 1, // SearchRequest{limit, offset, query, orderField, orderBy}
			nil, // FindUsers - Err
			[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24}, // resp.Users - ID
			true, // nextPage
		},
		{
			"Negative_limit", // test Name
			-1, 1, "", "", 0, // SearchRequest{limit, offset, query, orderField, orderBy}
			fmt.Errorf("limit must be > 0"), // FindUsers - Err
			[]int{},                         // resp.Users - ID
			true,                            // nextPage
		},
		{
			"Negative_offset", // test Name
			1, -1, "", "", 0,  // SearchRequest{limit, offset, query, orderField, orderBy}
			fmt.Errorf("offset must be > 0"), // FindUsers - Err
			[]int{},                          // resp.Users - ID
			true,                             // nextPage
		},
		{
			"File_not_found", // test Name
			1, 1, "", "", 0,  // SearchRequest{limit, offset, query, orderField, orderBy}
			fmt.Errorf("SearchServer fatal error"), // FindUsers - Err
			[]int{},                                // resp.Users - ID
			true,                                   // nextPage
		},
	}

	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer testServer.Close()

	sc := SearchClient{
		URL:         testServer.URL,
		AccessToken: "x",
	}

	for _, tc := range testCases {
		t.Run(tc.nameCase, func(t *testing.T) {
			if tc.nameCase == "File_not_found" { // проверка на то, что файл не будет найден
				old := FilePath
				FilePath = "qwerty.xml"
				defer func() { FilePath = old }()
			}

			resp, err := sc.FindUsers(SearchRequest{
				Limit:      tc.limit,
				Offset:     tc.offset,
				Query:      tc.query,
				OrderField: tc.orderField,
				OrderBy:    tc.orderBy,
			})

			if tc.ExpectedErr != nil {
				if tc.ExpectedErr.Error() != err.Error() {
					t.Errorf("expected err: %v, got: %v", tc.ExpectedErr, err)
				}
			}
			if tc.ExpectedErr == nil && tc.ExpectedErr != err {
				t.Errorf("expected err: %v, got: %v", tc.ExpectedErr, err)
			}

			if resp == nil {
				t.Logf("resp = nil, err = %v", err)
				return
			}

			if resp.NextPage != tc.expectedNext {
				t.Errorf("expected condition NextPage: %v, got: %v", tc.expectedNext, resp.NextPage)
			}

			if len(resp.Users) != len(tc.expectedIDs) {
				t.Fatalf("expected count results: %d, got: %d", len(tc.expectedIDs), len(resp.Users))
			}

			for i, user := range resp.Users {
				if user.ID != tc.expectedIDs[i] {
					t.Errorf("expected id: %d, got: %d", tc.expectedIDs[i], user.ID)
				}
			}
		})
	}
}

func TestFindUsers_NetErrors(t *testing.T) {
	cases := []struct {
		name       string
		statusCode int
		body       string
		sleep      time.Duration
		wantErr    string
	}{
		{"unauthorized", http.StatusUnauthorized, "", 0, "bad AccessToken"},
		{"500", http.StatusInternalServerError, "", 0, "SearchServer fatal error"},
		{"400_with_bad_json", http.StatusBadRequest, "not json", 0, "cant unpack error json"},
		{"400_with_unknown_error", http.StatusBadRequest, `{"Error":"oops"}`, 0, "unknown bad request error"},
		{"200_with_bad_json", http.StatusOK, "not an array", 0, "cant unpack result json"},
		{"timeout", http.StatusOK, "{}", 2 * time.Second, "timeout for"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.sleep > 0 {
					time.Sleep(tc.sleep)
				}
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer ts.Close()
			c := &SearchClient{URL: ts.URL, AccessToken: "x"}
			_, err := c.FindUsers(SearchRequest{Limit: 5, OrderBy: 1, OrderField: "Id"})
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("want error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestSearchServer(t *testing.T) {
	testCases := []struct {
<<<<<<< HEAD
		nameCase           string
		limit              string // int
		offset             string // int
		query              string
		orderField         string
		orderBy            string // int
		ExpectedStatusCode int
		expectedIDs        []int // какие ID ожидаем в resp.Users (в каком порядке)
=======
		nameCase     string
		limit        string // int
		offset       string // int
		query        string
		orderField   string
		orderBy      string // int
		ExpectedErr  error
		expectedIDs  []int // ← какие ID ожидаем в resp.Users (в каком порядке)
		expectedNext bool  // ← каким должен быть resp.NextPage
>>>>>>> d593b196f5c232e53217b47cbe701bb4e49ac161
	}{
		{
			"Success",             // test Name
			"5", "0", "", "", "1", // SearchRequest{limit, offset, query, orderField, orderBy}
<<<<<<< HEAD
			200,                      // Status Code
=======
			nil,                      // FindUsers - Err
>>>>>>> d593b196f5c232e53217b47cbe701bb4e49ac161
			[]int{15, 16, 19, 22, 5}, // resp.Users - ID
		},
<<<<<<< HEAD
		{
			"Big_offset",           // test Name
			"1", "76", "", "", "1", // SearchRequest{limit, offset, query, orderField, orderBy}
			200, // Status Code
			[]int{},
		},
		{
			"Negative_offset",       // test Name
			"1", "-76", "", "", "1", // SearchRequest{limit, offset, query, orderField, orderBy}
			400, // Status Code
			[]int{},
		},
		{
			"Bad_offset",             // test Name
			"1", "text", "", "", "1", // SearchRequest{limit, offset, query, orderField, orderBy}
			400, // Status Code
			[]int{},
		},
		{
			"Big_limit",              // test Name
			"60", "0", "", "Id", "1", // SearchRequest{limit, offset, query, orderField, orderBy}
			200, // Status Code
			[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34},
		},
		{
			"Negative_limit",        // test Name
			"-10", "1", "", "", "1", // SearchRequest{limit, offset, query, orderField, orderBy}
			400, // Status Code
			[]int{},
		},
		{
			"Bad_limit",              // test Name
			"text", "5", "", "", "1", // SearchRequest{limit, offset, query, orderField, orderBy}
			400, // Status Code
			[]int{},
		},
		{
			"Bad_order_field",        // test Name
			"5", "0", "", "qwe", "1", // SearchRequest{limit, offset, query, orderField, orderBy}
			400, // Status Code
			[]int{},
		},
		{
			"Bad_order_by",         // test Name
			"5", "0", "", "", "10", // SearchRequest{limit, offset, query, orderField, orderBy}
			400, // Status Code
			[]int{},
		},
		{
			"Text_order_by",          // test Name
			"5", "0", "", "", "text", // SearchRequest{limit, offset, query, orderField, orderBy}
			400, // Status Code
			[]int{},
		},
		{
			"Miss_access_token",   // test Name
			"5", "0", "", "", "1", // SearchRequest{limit, offset, query, orderField, orderBy}
			401, // Status Code
			[]int{},
		},
	}
	for _, tc := range testCases {
		url := fmt.Sprintf("/?limit=%s&offset=%s&order_field=%s&query=%s&order_by=%s", tc.limit, tc.offset, tc.orderField, tc.query, tc.orderBy)
		t.Run(tc.nameCase, func(t *testing.T) {
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			if tc.nameCase != "Miss_access_token" {
				req.Header.Set("AccessToken", "token")
			}
=======
	}

	for _, tc := range testCases {
		t.Run(tc.nameCase, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?limit=5&offset=0&order_field=&query=&order_by=1", nil)
			w := httptest.NewRecorder()
>>>>>>> d593b196f5c232e53217b47cbe701bb4e49ac161

			SearchServer(w, req)

			resp := w.Result()
<<<<<<< HEAD
			defer resp.Body.Close()

			statusCode := resp.StatusCode
			if statusCode != tc.ExpectedStatusCode {
				t.Errorf("expected status code: %d, got: %d", tc.ExpectedStatusCode, statusCode)
				return
			}

			users := make([]User, 0, 26)

			err := json.NewDecoder(resp.Body).Decode(&users)
			if err != nil {
				t.Log("err decode json")
				return
			}

			if len(users) != len(tc.expectedIDs) {
				t.Errorf("users count incorrect")
				return
			}

			for i, user := range users {
				if user.ID != tc.expectedIDs[i] {
					t.Errorf("expected ID of User: %d, got: %d", tc.expectedIDs[i], user.ID)
					return
				}
			}

=======
			// body, _ := io.ReadAll(resp.Body)

			users := make([]User, 0, 1)

			_ = json.NewDecoder(resp.Body).Decode(&users)
			if len(users) != len(tc.expectedIDs) {
				t.Errorf("users count incorrect")
			}
			for _, user := range users {
				fmt.Println(user.ID)
			}
>>>>>>> d593b196f5c232e53217b47cbe701bb4e49ac161
		})
	}

}

func TestSearchServer_BadXML(t *testing.T) {
	tmp, err := os.CreateTemp("", "bad-*.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	_, err = tmp.WriteString(`
	<root>
		<row>
			<id>not-an-int</id>
		</row>
	</root>
	`)

	if err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	old := FilePath
	FilePath = tmp.Name()
	defer func() { FilePath = old }()

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	c := &SearchClient{URL: ts.URL, AccessToken: "x"}
	_, err = c.FindUsers(SearchRequest{Limit: 5, OrderBy: 1, OrderField: "Id"})

	if err == nil {
		t.Fatal("expected: error, got: nil")
	}
	if !strings.Contains(err.Error(), "SearchServer fatal error") {
		t.Errorf(`want: "SearchServer fatal error", got: %q`, err.Error())
	}
}
