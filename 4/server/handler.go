package server

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
)

const (
	DatasetPath = "./server/dataset.xml"
)

type SearchHandler struct {}

func NewSearchHandler() *SearchHandler {
	return &SearchHandler{}
}

var _ http.Handler = &SearchHandler{}

func (h *SearchHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	users, err := h.readDataset(DatasetPath)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	searchReq, err := SearchRequestFromAPI(req)
	if err != nil {
		errResponse := SearchErrorResponse{
			Error: err.Error(),
		}
		body, _ := json.Marshal(errResponse)

		res.WriteHeader(http.StatusBadRequest)
		res.Write(body)		
		
		return
	}

	err = searchReq.Validate()
	if err != nil {
		errResponse := SearchErrorResponse{
			Error: err.Error(),
		}
		body, _ := json.Marshal(errResponse)

		res.WriteHeader(http.StatusBadRequest)
		res.Write(body)		
		
		return
	}

	users = h.filterUsers(users, searchReq.Query)
	users = h.sortUsers(users, searchReq.OrderBy, searchReq.OrderField)
	users = h.paginateUsers(users, searchReq.Offset, searchReq.Limit)

	usersAPI := make([]UserAPI, 0, len(users))
	for _, user := range users {
		usersAPI = append(usersAPI, user.ToAPI())
	}

	body, _ := json.Marshal(usersAPI)
	res.Write(body)
}

// HELPERS

func (h *SearchHandler) readDataset(path string) ([]User, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	userRows := UserRows{}
	err = xml.Unmarshal(data, &userRows)
	if err != nil {
		return nil, err
	}

	return userRows.Users, err
}

func (h *SearchHandler) filterUsers(users []User, query string) []User {
	filtered := make([]User, 0)
	for _, user := range users {
		name := user.FirstName + user.LastName
		about := user.About

		if strings.Contains(name, query) || strings.Contains(about, query){
			filtered = append(filtered, user)
		}
	}
	return filtered
}

func (h *SearchHandler) sortUsers(users []User, orderBy int, orderField string) []User {
	if orderBy != OrderByAsIs {
		isReverse := (orderBy == OrderByDesc)
	
		sort.Slice(users, func (i, j int) bool {
			switch orderField {
			case OrderFieldID:
				id1 := users[i].ID
				id2 := users[j].ID
				return (id1 <= id2) != isReverse			
			case OrderFieldAge:
				age1 := users[i].Age
				age2 := users[j].Age
				return (age1 <= age2) != isReverse	
			case OrderFieldName:
				name1 := users[i].FirstName + users[i].LastName
				name2 := users[j].FirstName + users[j].LastName
				return (strings.Compare(name1, name2) <= 0) != isReverse	
			}
	
			return false
		})
	}

	return users
}

func (h *SearchHandler) paginateUsers(users []User, offset int, limit int) []User {
	paginated := make([]User, 0, limit)
	for i := offset; i < offset + limit; i++ {
		if i >= len(users) {
			break
		}

		paginated = append(paginated, users[i])
	}

	return paginated
}