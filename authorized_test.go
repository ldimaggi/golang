package feature

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/almighty/almighty-core/client"
	goaclient "github.com/goadesign/goa/client"
)

/* Simple test to verify actions performed by non-authorized users */

/* Nested structure to define a work item */
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

/* Structure for http requests */
type api struct {
	c    *client.Client
	resp *http.Response
	err  error
	body map[string]interface{}
	//body [200]string
}

/* Define the loggin levels */
var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

/* Global variables */
var savedToken string
var idString string

/* Set up the logging - Ref: https://www.goinggo.net/2013/11/using-log-package-in-go.html*/
func Init(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {
	Trace = log.New(traceHandle, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(infoHandle, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

/* Set up the http request before each scenario */
func (a *api) newScenario(i interface{}) {
	a.c = nil
	a.resp = nil
	a.err = nil
	a.c = client.New(goaclient.HTTPClientDoer(http.DefaultClient))
	a.c.Host = "localhost:8080"
}

/* The payload used to create a workitem */
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

/* The payload used to update a workitem - to reassign the workitem*/
func updatePayload() *client.UpdateWorkItemPayload {
	return &client.UpdateWorkItemPayload{
		Type: "system.bug",
		Fields: map[string]interface{}{
			"system.title":    "remove this TEST workitem PLEASE - OK",
			"system.owner":    "BobbyOrr",
			"system.state":    "open",
			"system.creator":  "GordieHowe",
			"system.assignee": "Not WayneGretzky",
		},
		Version: 0,
	}
}

/* The payload used to update a workitem - to unassign the workitem*/
func updatePayloadUnassign() *client.UpdateWorkItemPayload {
	return &client.UpdateWorkItemPayload{
		Type: "system.bug",
		Fields: map[string]interface{}{
			"system.title":    "remove this TEST workitem PLEASE - OK",
			"system.owner":    "BobbyOrr",
			"system.state":    "open",
			"system.creator":  "GordieHowe",
			"system.assignee": "Jaromir Jagr",
		},
		Version: 1,
	}
}

/* Handle the GET/POST requests */
func (a *api) iSendRequestTo(requestMethod, endpoint string) error {
	switch endpoint {
	case "get_workitemtypes":
		Info.Println("Received GET request to get workitem types")
		resp, err := a.c.ListWorkitemtype(context.Background(), "/api/workitemtypes", nil)
		a.resp = resp
		a.err = err
	case "get_workitems":
		Info.Println("Received GET request to get workitems")
		var tempString string
		tempString = "/api/workitems" + "/" + idString
		resp, err := a.c.ListWorkitem(context.Background(), tempString, nil, nil)
		a.resp = resp
		a.err = err
	case "create_workitem":
		Info.Println("Received POST request to create workitem")
		resp, err := a.c.CreateWorkitem(context.Background(), "/api/workitems", createPayload())
		a.resp = resp
		a.err = err

		defer a.resp.Body.Close()
		htmlData, err := ioutil.ReadAll(a.resp.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		data := string(htmlData)
		Info.Println("The response is:")
		Info.Println(data)

	case "update_workitem":
		Info.Println("Received POST request to update workitem")
		resp, err := a.c.UpdateWorkitem(context.Background(), "/api/workitems/"+idString, updatePayload())
		a.resp = resp
		a.err = err
	case "update_workitem_unassign":
		Info.Println("Received POST request to update/unassign workitem")
		resp, err := a.c.UpdateWorkitem(context.Background(), "/api/workitems/"+idString, updatePayloadUnassign())
		a.resp = resp
		a.err = err
		a.printResponse()

	case "delete_workitem":
		Info.Println("Received POST request to delete workitem")
		resp, err := a.c.DeleteWorkitem(context.Background(), "/api/workitems/"+idString)
		a.resp = resp
		a.err = err
	default:
		return godog.ErrPending
	}
	return nil
}

/* Check the value of the http responses */
func (a *api) theResponseCodeShouldBe(statusCode int) error {
	if a.resp.StatusCode != statusCode {
		return fmt.Errorf("Expected %d but was %d", statusCode, a.resp.StatusCode)
	}
	return nil
}

/* Check the contents of the http responses */
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

/* For authorized users - no set up is needed */
func (a *api) imAuthorized() error {

	/* Set up authorization with the token obtained earlier in the test */
	a.c.SetJWTSigner(&goaclient.APIKeySigner{
		SignQuery: false,
		KeyName:   "Authorization",
		KeyValue:  savedToken,
		Format:    "Bearer %s",
	})
	return nil
}

/* This function creates a new work item, and returns the ID of that work item */
func (a *api) setUpTestData() {
	Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	a.c = nil
	a.resp = nil
	a.err = nil
	a.c = client.New(goaclient.HTTPClientDoer(http.DefaultClient))
	a.c.Host = "localhost:8080"

	resp, err := a.c.ShowStatus(context.Background(), "api/login/generate")
	a.resp = resp
	a.err = err

	/* Retrieve the authentication token needed to create/delete workitems
	   Example of a token is:
	   "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJmdWxsTmFtZSI6IlRlc3QgRGV2ZWxvcGVyIiwiaW1hZ2VVUkwiOiIiLCJ1dWlkIjoiNGI4Zjk0YjUtYWQ4OS00NzI1LWI1ZTUtNDFkNmJiNzdkZjFiIn0.ML2N_P2qm-CMBliUA1Mqzn0KKAvb9oVMbyynVkcyQq3myumGeCMUI2jy56KPuwIHySv7i-aCUl4cfIjG-8NCuS4EbFSp3ja0zpsv1UDyW6tr-T7jgAGk-9ALWxcUUEhLYSnxJoEwZPQUFNTWLYGWJiIOgM86__OBQV6qhuVwjuMlikYaHIKPnetCXqLTMe05YGrbxp7xgnWMlk9tfaxgxAJF5W6WmOlGaRg01zgvoxkRV-2C6blimddiaOlK0VIsbOiLQ04t9QA8bm9raLWX4xOkXN4ubpdsobEzcJaTD7XW0pOeWPWZY2cXCQulcAxfIy6UmCXA14C07gyuRs86Rw"   */

	// Option 1 - Extarct the 1st token from the html Data in the reponse
	defer a.resp.Body.Close()
	htmlData, err := ioutil.ReadAll(a.resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//fmt.Println("[[[", string(htmlData), "]]]")
	lastBin := strings.LastIndex(string(htmlData), "\"},{\"token\":\"")
	Info.Println("The token to be used is:", string(htmlData)[11:lastBin])

	// Option 2 - Extract the 1st token from JSON in the response
	lastBin = strings.LastIndex(string(htmlData), ",")
	//Info.Println("The token to be used is:", string(htmlData)[11:lastBin])

	// TODO - Extract the token from the JSON map read from the html Data in the response
	byt := []byte(string(htmlData)[1:lastBin])
	var keys map[string]interface{}
	json.Unmarshal(byt, &keys)
	savedToken = fmt.Sprint(keys["token"])

	a.c.SetJWTSigner(&goaclient.APIKeySigner{
		SignQuery: false,
		KeyName:   "Authorization",
		KeyValue:  savedToken,
		Format:    "Bearer %s",
	})

	resp, err = a.c.CreateWorkitem(context.Background(), "/api/workitems", createPayload())
	//fmt.Println("body = ", resp.Body)
	//fmt.Println("error = ", resp.Status)
	a.resp = resp
	a.err = err

	defer a.resp.Body.Close()
	htmlData, err = ioutil.ReadAll(a.resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//fmt.Println(os.Stdout, string(htmlData))
	Info.Println("The newly created workitem is:", string(htmlData))

	idStart := strings.LastIndex(string(htmlData), "\"id\":\"")
	tmpString := string(htmlData)[idStart+6:]
	idEnd := strings.Index(tmpString, "\"")
	idString = tmpString[:idEnd]
	Info.Println("The ID of the newly created workitem is:", idString)
}

/* Function to delete a work item - requires authorization */
func (a *api) cleanUpTestData() {
	fmt.Println("Nothing to see here - move along")

	//	/* Set up authorization with the token obtained earlier in the test */
	//	a.c.SetJWTSigner(&goaclient.APIKeySigner{
	//		SignQuery: false,
	//		KeyName:   "Authorization",
	//		KeyValue:  savedToken,
	//		Format:    "Bearer %s",
	//	})
	//
	//	/* Delete the workitem */
	//	Info.Println("The ID of the workitem to be deleted is:", idString)
	//	resp, err := a.c.DeleteWorkitem(context.Background(), "/api/workitems/"+idString)
	//	a.resp = resp
	//	a.err = err
	//	if err != nil {
	//		fmt.Println(err)
	//		os.Exit(1)
	//	}
}

func (a *api) printResponse() {
	fmt.Println("Nothing to see here - move along")

	defer a.resp.Body.Close()
	htmlData, err := ioutil.ReadAll(a.resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	data := string(htmlData)
	Info.Println("The response is:")
	Info.Println(data)

}

func FeatureContext(s *godog.Suite) {
	a := api{}
	s.BeforeSuite(a.setUpTestData)
	s.BeforeScenario(a.newScenario)
	s.Step(`^I\'m authorized$`, a.imAuthorized)
	s.Step(`^I send "([^"]*)" request to "([^"]*)"$`, a.iSendRequestTo)
	s.Step(`^the response code should be (\d+)$`, a.theResponseCodeShouldBe)
	s.Step(`^the response should contain fields:$`, a.theResponseShouldContainFields)
	s.AfterSuite(a.cleanUpTestData)
}
