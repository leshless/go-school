package hw3

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mailru/easyjson"
)

//easyjson:json
type User struct {
	Browsers []string `json:"browsers"`
	Company string `json:"company"`
	Country string `json:"country"`
	Email string `json:"email"`
	Job string `json:"job"`
	Name string `json:"name"`
	Phone string `json:"phone"`
}

func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	
	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	
	file.Close()
	
	lines := strings.Split(string(fileContents), "\n")
	
	users := make([]User, 0, len(lines))
	for _, line := range lines {
		var user User
		err := easyjson.Unmarshal([]byte(line), &user)
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	seenBrowsers := make(map[string]struct{})
	var foundUsersB strings.Builder
	
	for i, user := range users {

		containsAndroid := false
		containsMSIE := false

		for _, browser := range user.Browsers {
			isAndroid := strings.Contains(browser, "Android")
			containsAndroid = containsAndroid || isAndroid
			
			isMSIE := strings.Contains(browser, "MSIE")
			containsMSIE = containsMSIE || isMSIE
			
			if isAndroid || isMSIE {
				seenBrowsers[browser] = struct{}{}
			}
		}

		if !(containsAndroid && containsMSIE) {
			continue
		}

		foundUsersB.WriteString(fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, user.Email))
	}

	foundUsers := strings.ReplaceAll(foundUsersB.String(), "@", " [at] ")

	fmt.Fprintf(out, "found users:\n%s\n", foundUsers)
	fmt.Fprintf(out, "Total unique browsers %d\n", len(seenBrowsers))
}
 