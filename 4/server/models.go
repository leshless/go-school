package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type User struct {
	ID     int `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName string `xml:"last_name"`
	Age    int `xml:"age"`
	About  string `xml:"about"`
	Gender string `xml:"gender"`
}

type UserRows struct {
	Users []User `xml:"row"`
}

type UserAPI struct {
	ID     int `json:"id"`
	Name   string `json:"name"`
	Age    int `json:"age"`
	About  string `json:"about"`
	Gender string `json:"gender"`
}

func (u User) ToAPI() UserAPI {
	return UserAPI{
		ID: u.ID,
		Name: u.FirstName + u.LastName,
		Age: u.Age,
		About: u.About,
		Gender: u.Gender,
	}
}

// This should be enum
const (
	OrderByAsc = -1
	OrderByAsIs = 0
	OrderByDesc = 1
)

// This also should be enum
const (
	OrderFieldID = "id"
	OrderFieldAge = "age"
	OrderFieldName = "name"
)

// This should be just http status instead of error in response body
// I do not understand why mr.Romanov wants us to write it like this
var (
	ErrorBadOrderField = fmt.Errorf("ErrorBadOrderField")
	ErrorBadRequest = fmt.Errorf("ErrorBadRequest")
)

type SearchRequest struct{
	Limit      int
	Offset     int
	Query      string
	OrderField string
	OrderBy    int
}

func SearchRequestFromAPI(req *http.Request) (SearchRequest, error) {
	params, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return SearchRequest{}, ErrorBadRequest
	}
	
	query := params.Get("query")
	
	orderField := params.Get("order_field")
	if orderField == ""{
		orderField = OrderFieldName
	}

	limit, err := strconv.Atoi(params.Get("limit"))
	if err != nil {
		return SearchRequest{}, ErrorBadRequest
	}
	offset, err := strconv.Atoi(params.Get("offset"))
	if err != nil {
		return SearchRequest{}, ErrorBadRequest
	}
	orderBy, err := strconv.Atoi(params.Get("order_by"))
	if err != nil {
		return SearchRequest{}, ErrorBadRequest
	}

	return SearchRequest{
		Limit: limit,
		Offset: offset,
		Query: query,
		OrderField: orderField,
		OrderBy: orderBy,
	}, nil
}

func (req SearchRequest) Validate() error {
	if req.Limit <= 0 {
		return ErrorBadRequest
	}

	if req.Offset < 0 {
		return ErrorBadRequest
	}

	if (req.OrderField != OrderFieldID) && (req.OrderField != OrderFieldAge) && (req.OrderField != OrderFieldName) {
		return ErrorBadOrderField
	}

	if (req.OrderBy != OrderByAsc) && (req.OrderBy != OrderByAsIs) && (req.OrderBy != OrderByDesc) {
		return ErrorBadRequest
	}

	return nil
}

// SearchResponse is not needed because its just []UserAPI

type SearchErrorResponse struct {
	Error string `json:"error"`
}