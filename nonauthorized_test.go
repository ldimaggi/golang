package feature

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/almighty/almighty-core/client"
	goaclient "github.com/goadesign/goa/client"
)

// Simple test to create and retrieve workitem

type Structworkitem struct {
	Fields  Structworkitemfields `json:"fields"`
	ID      string               `json:"id"`
	Type    string               `json:"type"`
	Version string               `json:"version"`
}
type Structworkitemfields struct {
	Assignee    string `json:"system.assignee"`
	Creator     string `json:"system.creator"`
	Description string `json:"system.description"`
	State       string `json:"system.state"`
	Title       string `json:"system.title"`
}

type api struct {
	c    *client.Client
	resp *http.Response
	err  error
	body map[string]interface{}
	//body [200]string
}

var savedToken string

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
			"system.title":    "remove this TEST workitem PLEASE - OK",
			"system.owner":    "BobbyOrr",
			"system.state":    "open",
			"system.creator":  "GordieHowe",
			"system.assignee": "WayneGretzky",
		},
	}
}

func (a *api) iSendRequestTo(requestMethod, endpoint, extraParameter string) error {
	switch endpoint {
	case "get_workitemtypes":
		resp, err := a.c.ListWorkitemtype(context.Background(), "/api/workitemtypes", nil)
		a.resp = resp
		a.err = err
	case "get_workitems":
		var tempString string
		tempString = "/api/workitems" + "/" + extraParameter
		//fmt.Println(tempString)
		resp, err := a.c.ListWorkitem(context.Background(), tempString, nil, nil)
		a.resp = resp
		a.err = err
	case "create_workitem":
		resp, err := a.c.CreateWorkitem(context.Background(), "/api/workitems", createPayload())
		//fmt.Println("body = ", resp.Body)
		//fmt.Println("error = ", resp.Status)
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
	//fmt.Println("the string ", string(theDocString.Content))
	defer a.resp.Body.Close()
	htmlData, err := ioutil.ReadAll(a.resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	data := string(htmlData)
	w := Structworkitem{}
	json.Unmarshal([]byte(data), &w)

	byt := []byte(string(string(theDocString.Content)))
	var keys map[string]interface{}
	json.Unmarshal(byt, &keys)
	for key, value := range keys {
		//		fmt.Printf("the key = %v", key)
		//		fmt.Printf("the value = %v", value)
		if key == "assignee" {
			if value != w.Fields.Assignee {
				return fmt.Errorf("Expected %s but was %s", value, w.Fields.Assignee)
			}
		}
	}
	return nil
}

func imNotAuthorized() error {
	//	return godog.ErrPending
	// fmt.Println("Nothing to see here - move along")
	return nil
}

func FeatureContext(s *godog.Suite) {
	a := api{}
	s.BeforeScenario(a.newScenario)
	s.Step(`^I\'m not authorized$`, imNotAuthorized)
	s.Step(`^I send "([^"]*)" request to "([^"]*)" "([^"]*)"$`, a.iSendRequestTo)
	s.Step(`^the response code should be (\d+)$`, a.theResponseCodeShouldBe)
	s.Step(`^the response should contain fields:$`, a.theResponseShouldContainFields)
}
