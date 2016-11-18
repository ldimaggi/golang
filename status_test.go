package feature

import (
	"context"
	//	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/almighty/almighty-core/client"
	goaclient "github.com/goadesign/goa/client"
)

// Simple test to retrieve workitem

type api struct {
	c    *client.Client
	resp *http.Response
	err  error
	body [200]string
}

func (a *api) newScenario(i interface{}) {
	a.c = nil
	a.resp = nil
	a.err = nil
	a.c = client.New(goaclient.HTTPClientDoer(http.DefaultClient))
	a.c.Host = "localhost:8080"
}

func createPayload() *client.CreateWorkItemPayload {
	return &client.CreateWorkItemPayload{
		Type: "system.bug",
		Fields: map[string]interface{}{
			"foo":    "bar",
			"blabla": -1,
		},
	}
}

func (a *api) iSendRequestTo(requestMethod, endpoint string) error {
	switch endpoint {
	case "get_workitemtypes":
		//		resp, err := a.c.ShowStatus(context.Background(), "/api/workitems/20")
		resp, err := a.c.ListWorkitemtype(context.Background(), "/api/workitemtypes", nil)
		a.resp = resp
		a.err = err
	case "get_workitems":
		resp, err := a.c.ListWorkitem(context.Background(), "/api/workitems", nil, nil)
		a.resp = resp
		a.err = err
	case "create_workitem":
		// Question for Aslak - how to create the payload?
		resp, err := a.c.AuthorizeLogin(context.Background(), "/api/login/authorize")
		resp, err = a.c.CreateWorkitem(context.Background(), "/api/workitems", createPayload(), "newType")
		a.resp = resp
		a.err = err

	default:
		return godog.ErrPending
	}
	return nil
}

func (a *api) theResponseCodeShouldBe(statusCode int) error {
	if a.resp.StatusCode != statusCode {
		return fmt.Errorf("Expected %d but was %d", statusCode, a.resp.StatusCode)
	}
	return nil
}

func (a *api) theResponseShouldContainFields(theDocString *gherkin.DocString) error {
	fmt.Println(string(theDocString.Content))

	defer a.resp.Body.Close()
	htmlData, err := ioutil.ReadAll(a.resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(os.Stdout, string(htmlData))
	return nil
}

func FeatureContext(s *godog.Suite) {
	a := api{}
	s.BeforeScenario(a.newScenario)
	s.Step(`^I send "([^"]*)" request to "([^"]*)"$`, a.iSendRequestTo)
	s.Step(`^the response code should be (\d+)$`, a.theResponseCodeShouldBe)
	s.Step(`^the response should contain fields:$`, a.theResponseShouldContainFields)
}
