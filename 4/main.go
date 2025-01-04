package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
)

const (
	DatasetPath = "./dataset.xml"
)

type UserRow struct {
	Id     int `xml:"id"`
	FirstName   string `xml:"first_name"`
	LastName string `xml:"last_name"`
	Age    int `xml:"age"`
	About  string `xml:"about"`
	Gender string `xml:"gender"`
}

type UserRows struct {
	Users []UserRow `xml:"row"`
}

func ReadDataset(path string) ([]UserRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

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

func Handler(req *http.Request, res http.ResponseWriter) (SearchResponse, error) {
	users, err := ReadDataset(DatasetPath)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("failed to read dataset")
	}

	filtered := make([]UserRow, 0)
	for _, user := range users {
		name := user.FirstName + user.LastName
		about := user.About

		if strings.Contains(name, req.Query) || strings.Contains(about, req.Query){
			filtered = append(filtered, user)
		}
	}
	users = filtered

	sort.Slice(users, func (i, j int) bool {
		swi
	})


}

func main() {
	userRows, err := ReadDataset("dataset.xml")
	if err != nil {
		panic(err)
	}

	fmt.Println(userRows)
}