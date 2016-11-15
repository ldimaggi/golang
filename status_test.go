package feature

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/almighty/almighty-core/client"
	goaclient "github.com/goadesign/goa/client"
)

type api struct {
	c    *client.Client
	resp *http.Response
	err  error
	body map[string]interface{}
}

func (a *api) newScenario(i interface{}) {
	a.c = nil
	a.resp = nil
	a.err = nil

	a.c = client.New(goaclient.HTTPClientDoer(http.DefaultClient))
	a.c.Host = "localhost:8080"
}

func (a *api) iSendRequestTo(requestMethod, endpoint string) error {
	switch endpoint {
	case "status":
		resp, err := a.c.ShowStatus(context.Background(), "/api/status")
		a.resp = resp
		a.err = err
		json.NewDecoder(a.resp.Body).Decode(&a.body)

	case "workitemtypes":
		resp, err := a.c.ShowStatus(context.Background(), "/api/workitemtypes")
		a.resp = resp
		a.err = err
		json.NewDecoder(a.resp.Body).Decode(&a.body)
	default:
		return godog.ErrPending
	}
	return nil
}

func (a *api) theResponseCodeShouldBe(statusCode int) error {
	if a.resp.StatusCode != statusCode {
		return fmt.Errorf("Expected %d but was %d", a.resp.StatusCode, statusCode)
	}
	return nil
}

func (a *api) theResponseShouldContainJSON(jsonKeys *gherkin.DocString) error {
	var keys map[string]interface{}
	json.NewDecoder(strings.NewReader(jsonKeys.Content)).Decode(&keys)

        for key := range keys {
		fmt.Printf("THE KEY = " + key)
		fmt.Printf("%+v\n", a.body)
		if _, ok := a.body[key]; !ok {
			return fmt.Errorf("Expected key %s to exist, but got %v", key, a.body)
		}
	}

	return nil
}

func FeatureContext(s *godog.Suite) {
	a := api{}

	s.BeforeScenario(a.newScenario)

	s.Step(`^I send "([^"]*)" request to "([^"]*)"$`, a.iSendRequestTo)
	s.Step(`^the response code should be (\d+)$`, a.theResponseCodeShouldBe)
	s.Step(`^the response should contain json:$`, a.theResponseShouldContainJSON)
}
