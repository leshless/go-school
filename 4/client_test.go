package hw4

import (
	"hw4/server"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)


type ClientTestSuite struct {
	suite.Suite

	someServer *httptest.Server
	someRequest SearchRequest
	someResponse SearchResponse

	client *SearchClient
}

func TestClient(t *testing.T){
	suite.Run(t, new(ClientTestSuite))
}

func (t *ClientTestSuite) SetupTest(){
	handler := server.AuthorizeMiddleware(server.NewSearchHandler())
	t.someServer = httptest.NewServer(handler)

	t.client = &SearchClient{
		AccessToken: "top_secret",
		URL: t.someServer.URL,
	}

	t.someRequest = SearchRequest{
		Limit: 2,
		Offset: 0,
		Query: "",
		OrderField: "id",
		OrderBy: OrderByAsIs,
	}

	t.someResponse = SearchResponse{
		Users: []User{
			{
				Id: 0,
				Name: "BoydWolf",
				Age: 22,
				About: "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
				Gender: "male",
			},
			{
				Id: 1,
				Name: "HildaMayer",
				Age: 21,
				About: "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
				Gender: "female",
			},
		},
		NextPage: true,
	}
}

func (t *ClientTestSuite) TearDownTest(){
	t.someServer.Close()
}

func (t *ClientTestSuite) Test_Client_GreenPath(){
	response, err := t.client.FindUsers(t.someRequest)
	
	t.Assert().NoError(err)
	t.Assert().Equal(t.someResponse, *response)
}

func (t *ClientTestSuite) Test_Client_WhenLimitIsGreaterThan25(){
	t.someRequest.Limit = 26

	response, err := t.client.FindUsers(t.someRequest)
	
	t.Assert().NoError(err)
	t.Assert().Len(response.Users, 25)
	t.Assert().True(response.NextPage)
}

func (t *ClientTestSuite) Test_Client_WhenNoNextPage(){
	t.someRequest.Limit = 0
	t.someRequest.Offset = 35
	t.someResponse = SearchResponse{
		Users: []User{},
		NextPage: false,
	}

	response, err := t.client.FindUsers(t.someRequest)
	
	t.Assert().NoError(err)
	t.Assert().Equal(t.someResponse, *response)
}

func (t *ClientTestSuite) Test_Client_WhenInvalidLimit(){
	t.someRequest.Limit = -1

	response, err := t.client.FindUsers(t.someRequest)
	
	t.Assert().Error(err)
	t.Assert().Equal(err.Error(), "limit must be > 0")
	t.Assert().Empty(response)
}

func (t *ClientTestSuite) Test_Client_WhenInvalidOffset(){
	t.someRequest.Offset = -1

	response, err := t.client.FindUsers(t.someRequest)
	
	t.Assert().Error(err)
	t.Assert().Contains(err.Error(), "offset must be > 0")
	t.Assert().Empty(response)
}

func (t *ClientTestSuite) Test_Client_WhenInvalidURL(){
	t.client.URL = "some bad url"

	response, err := t.client.FindUsers(t.someRequest)
	
	t.Assert().Error(err)
	t.Assert().Contains(err.Error(), "unknown error")
	t.Assert().Empty(response)
}


type MockBusyHandler struct{}

func (h *MockBusyHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	time.Sleep(2 * time.Second)
}

func (t *ClientTestSuite) Test_Client_WhenServerTimeout(){
	t.someServer.Close()
	
	t.someServer = httptest.NewServer(&MockBusyHandler{})
	t.client.URL = t.someServer.URL

	response, err := t.client.FindUsers(t.someRequest)
	
	t.Assert().Error(err)
	t.Assert().Contains(err.Error(), "timeout for")
	t.Assert().Empty(response)
}

func (t *ClientTestSuite) Test_Client_WhenWrongAccessToken(){
	t.client.AccessToken = "some_access_token"
	
	response, err := t.client.FindUsers(t.someRequest)

	t.Assert().Error(err)
	t.Assert().Contains(err.Error(), "Bad AccessToken")
	t.Assert().Empty(response)
}

type MockBrokenHandler struct{}

func (h *MockBrokenHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusInternalServerError)
}

func (t *ClientTestSuite) Test_Client_WhenServerInternalError(){
	t.someServer.Close()

	t.someServer = httptest.NewServer(&MockBrokenHandler{})
	t.client.URL = t.someServer.URL
	
	response, err := t.client.FindUsers(t.someRequest)

	t.Assert().Error(err)
	t.Assert().Contains(err.Error(), "SearchServer fatal error")
	t.Assert().Empty(response)
}

type MockBadErrorFormatHandler struct{}

func (h *MockBadErrorFormatHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte("i'm not a json!!! :)"))
}

func (t *ClientTestSuite) Test_Client_WhenBadServerErrorFormat(){
	t.someServer.Close()

	t.someServer = httptest.NewServer(&MockBadErrorFormatHandler{})
	t.client.URL = t.someServer.URL
	
	response, err := t.client.FindUsers(t.someRequest)

	t.Assert().Error(err)
	t.Assert().Contains(err.Error(), "cant unpack error json:")
	t.Assert().Empty(response)
}

func (t *ClientTestSuite) Test_Client_WhenBadOrderFieldError(){
	t.someRequest.OrderField = "job"
	
	response, err := t.client.FindUsers(t.someRequest)

	t.Assert().Error(err)
	t.Assert().Contains(err.Error(), "OrderFeld job invalid")
	t.Assert().Empty(response)
}

func (t *ClientTestSuite) Test_Client_WhenBadRequestError(){
	t.someRequest.OrderBy = 10
	
	response, err := t.client.FindUsers(t.someRequest)

	t.Assert().Error(err)
	t.Assert().Contains(err.Error(), "unknown bad request error:")
	t.Assert().Empty(response)
}

type MockBadResponseFormatHandler struct{}

func (h *MockBadResponseFormatHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("i'm not a json!!! :)"))
}

func (t *ClientTestSuite) Test_Client_WhenBadServerResponseFormat(){
	t.someServer.Close()

	t.someServer = httptest.NewServer(&MockBadResponseFormatHandler{})
	t.client.URL = t.someServer.URL

	response, err := t.client.FindUsers(t.someRequest)

	t.Assert().Error(err)
	t.Assert().Contains(err.Error(), "cant unpack result json:")
	t.Assert().Empty(response)
}
