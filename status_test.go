package feature

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/almighty/almighty-core/client"
	goaclient "github.com/goadesign/goa/client"
)

// Simple test to create and retrieve workitem

type api struct {
	c    *client.Client
	resp *http.Response
	err  error
	body [200]string
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
			"system.title":   "remove this workitem PLEASE- OK",
			"system.owner":   "ldimaggi",
			"system.state":   "open",
			"system.creator": "ldimaggi",
		},
	}
}

func (a *api) iSendRequestTo(requestMethod, endpoint string) error {
	switch endpoint {
	case "get_workitemtypes":
		resp, err := a.c.ListWorkitemtype(context.Background(), "/api/workitemtypes", nil)
		a.resp = resp
		a.err = err
	case "get_workitems":
		resp, err := a.c.ListWorkitem(context.Background(), "/api/workitems", nil, nil)
		a.resp = resp
		a.err = err
	case "create_workitem":
		resp, err := a.c.CreateWorkitem(context.Background(), "/api/workitems", createPayload(), "newType")
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

func (a *api) imAuthorized() error {
	resp, err := a.c.ShowStatus(context.Background(), "api/login/generate")
	a.resp = resp
	a.err = err

	// Option 1 - Extarct the 1st token from the html Data in the reponse
	defer a.resp.Body.Close()
	htmlData, err := ioutil.ReadAll(a.resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//fmt.Println("[[[", string(htmlData), "]]]")
	lastBin := strings.LastIndex(string(htmlData), "\"},{\"token\":\"")
	fmt.Printf("The token to use is: %v\n", string(htmlData)[11:lastBin])

	// Option 2 - Extract the 1st token from JSON in the response
	lastBin = strings.LastIndex(string(htmlData), ",")
	fmt.Printf("The token to use is: %v\n", string(htmlData)[1:lastBin])

	// TODO - Extract the token from the JSON map read from the html Data in the response
	byt := []byte(string(htmlData)[1:lastBin])
	var keys map[string]interface{}
	json.Unmarshal(byt, &keys)
	savedToken = fmt.Sprint(keys["token"])

	//key := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJmdWxsTmFtZSI6IlRlc3QgRGV2ZWxvcGVyIiwiaW1hZ2VVUkwiOiIiLCJ1dWlkIjoiNGI4Zjk0YjUtYWQ4OS00NzI1LWI1ZTUtNDFkNmJiNzdkZjFiIn0.ML2N_P2qm-CMBliUA1Mqzn0KKAvb9oVMbyynVkcyQq3myumGeCMUI2jy56KPuwIHySv7i-aCUl4cfIjG-8NCuS4EbFSp3ja0zpsv1UDyW6tr-T7jgAGk-9ALWxcUUEhLYSnxJoEwZPQUFNTWLYGWJiIOgM86__OBQV6qhuVwjuMlikYaHIKPnetCXqLTMe05YGrbxp7xgnWMlk9tfaxgxAJF5W6WmOlGaRg01zgvoxkRV-2C6blimddiaOlK0VIsbOiLQ04t9QA8bm9raLWX4xOkXN4ubpdsobEzcJaTD7XW0pOeWPWZY2cXCQulcAxfIy6UmCXA14C07gyuRs86Rw" // call api to get key
	a.c.SetJWTSigner(&goaclient.APIKeySigner{
		SignQuery: false,
		KeyName:   "Authorization",
		KeyValue:  savedToken,
		Format:    "Bearer %s",
	})
	return nil
}

func FeatureContext(s *godog.Suite) {
	a := api{}
	s.BeforeScenario(a.newScenario)
	s.Step(`^I\'m authorized$`, a.imAuthorized)
	s.Step(`^I send "([^"]*)" request to "([^"]*)"$`, a.iSendRequestTo)
	s.Step(`^the response code should be (\d+)$`, a.theResponseCodeShouldBe)
	s.Step(`^the response should contain fields:$`, a.theResponseShouldContainFields)
}
