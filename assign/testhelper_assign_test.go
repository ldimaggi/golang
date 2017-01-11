package feature

/* Simple GoDog BDD test to verify actions performed by authorized users
   This test requires the "testhelper_assign_test.go" file for function definitions
   and the ./../../features/assign/authorized file for feature definitions
   This test also requires the server to be running in development mode as the
   test generates tokens from: api/login/generate

   @author ldimaggi

   TODO
   - The "backlog" prefix on the context structure used in the test is adapted
     from the backlog management test and "helper" file - (author Vineet Reynolds) we
	 should look at creating a single heloper package that can be shared by all tests
*/

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/almighty/almighty-core/account"
	"github.com/almighty/almighty-core/app"
	"github.com/almighty/almighty-core/client"
	"github.com/almighty/almighty-core/configuration"
	"github.com/almighty/almighty-core/gormsupport"
	"github.com/almighty/almighty-core/migration"
	"github.com/almighty/almighty-core/models"
	testsupport "github.com/almighty/almighty-core/test"
	almtoken "github.com/almighty/almighty-core/token"
	"github.com/almighty/almighty-core/workitem"
	"github.com/goadesign/goa"
	goaclient "github.com/goadesign/goa/client"
	"github.com/jinzhu/gorm"
	"github.com/mitchellh/mapstructure"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

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
}

/* Copied from workitem_blackbox_test.go - we need this in order to be able
   generate user IDs to be used in the assigning of work items to users
*/
type WorkItem2Suite struct {
	suite.Suite
	db             *gorm.DB
	clean          func()
	wiCtrl         app.WorkitemController
	wi2Ctrl        app.WorkitemController
	pubKey         *rsa.PublicKey
	priKey         *rsa.PrivateKey
	svc            *goa.Service
	wi             *app.WorkItem2
	minimumPayload *app.UpdateWorkitemPayload
}

/* Used for http request/response */
type Api struct {
	c    *client.Client
	resp *http.Response
	err  error
	body map[string]interface{}
}

/* Opens a new connection at localhost:8080 */
func (a *Api) Reset() {
	a.c = nil
	a.resp = nil
	a.err = nil

	a.c = client.New(goaclient.HTTPClientDoer(http.DefaultClient))
	a.c.Host = "localhost:8080"
}

/* Used to store the generated token required for write operations */
type IdentityHelper struct {
	savedToken string
}

/* Create a new token from api/login/generate */
func (i *IdentityHelper) GenerateToken(a *Api) error {
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
	//fmt.Printf("The token to use is: %v\n", string(htmlData)[11:lastBin])

	// Option 2 - Extract the 1st token from JSON in the response
	lastBin = strings.LastIndex(string(htmlData), ",")
	//fmt.Printf("The token to use is: %v\n", string(htmlData)[1:lastBin])

	// TODO - Extract the token from the JSON map read from the html Data in the response
	byt := []byte(string(htmlData)[1:lastBin])
	var keys map[string]interface{}
	json.Unmarshal(byt, &keys)
	token := fmt.Sprint(keys["token"])
	if token == "" {
		return fmt.Errorf("Failed to obtain a login token")
	}
	i.savedToken = token

	//key := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJmdWxsTmFtZSI6IlRlc3QgRGV2ZWxvcGVyIiwiaW1hZ2VVUkwiOiIiLCJ1dWlkIjoiNGI4Zjk0YjUtYWQ4OS00NzI1LWI1ZTUtNDFkNmJiNzdkZjFiIn0.ML2N_P2qm-CMBliUA1Mqzn0KKAvb9oVMbyynVkcyQq3myumGeCMUI2jy56KPuwIHySv7i-aCUl4cfIjG-8NCuS4EbFSp3ja0zpsv1UDyW6tr-T7jgAGk-9ALWxcUUEhLYSnxJoEwZPQUFNTWLYGWJiIOgM86__OBQV6qhuVwjuMlikYaHIKPnetCXqLTMe05YGrbxp7xgnWMlk9tfaxgxAJF5W6WmOlGaRg01zgvoxkRV-2C6blimddiaOlK0VIsbOiLQ04t9QA8bm9raLWX4xOkXN4ubpdsobEzcJaTD7XW0pOeWPWZY2cXCQulcAxfIy6UmCXA14C07gyuRs86Rw" // call api to get key
	a.c.SetJWTSigner(&goaclient.APIKeySigner{
		SignQuery: false,
		KeyName:   "Authorization",
		KeyValue:  i.savedToken,
		Format:    "Bearer %s",
	})

	userResp, userErr := a.c.ShowUser(context.Background(), "/api/user")
	var user map[string]interface{}
	json.NewDecoder(userResp.Body).Decode(&user)

	if userErr != nil {
		fmt.Printf("Error: %s", userErr)
	}
	return nil
}

/* Reset the token - this is used to simulate an unauthorized user */
func (i *IdentityHelper) Reset() {
	i.savedToken = ""
}

/* Context with which all the tests run - adapted from backlog management tests */
type BacklogContext struct {
	api            Api
	identityHelper IdentityHelper
	space          client.SpaceSingle
	spaceCreated   bool
	iteration      client.IterationSingle
	workItem       client.WorkItem2Single
	iterationName  string
	spaceName      string
}

/* Create a new connection and get a new token */
func (i *BacklogContext) Reset(v interface{}) {
	i.api.Reset()
	i.generateToken()
}

/* Needed for authorized users */
func (i *BacklogContext) aUserWithPermissions() error {
	return i.generateToken()
}

/* Create a new token */
func (i *BacklogContext) generateToken() error {
	err := i.identityHelper.GenerateToken(&i.api)
	if err != nil {
		return err
	}
	return nil
}

/* Create new space */
func (i *BacklogContext) anExistingSpace() error {
	if i.spaceCreated == false {
		a := i.api
		resp, err := a.c.CreateSpace(context.Background(), client.CreateSpacePath(), i.createSpacePayload())
		a.resp = resp
		a.err = err
		dec := json.NewDecoder(a.resp.Body)
		if err := dec.Decode(&i.space); err == io.EOF {
			return i.verifySpace()
		} else if err != nil {
			panic(err)
		}
		return i.verifySpace()
	}
	return nil
}

/* And verify a space */
func (i *BacklogContext) verifySpace() error {
	if len(i.space.Data.ID) < 1 {
		return fmt.Errorf("Expected a space with ID, but ID was [%s]", i.space.Data.ID)
	}
	expectedTitle := i.spaceName
	actualTitle := i.space.Data.Attributes.Name
	if *actualTitle != expectedTitle {
		return fmt.Errorf("Expected a space with title %s, but title was [%s]", expectedTitle, *actualTitle)
	}
	i.spaceCreated = true
	return nil
}

/* Payload used to create a space */
func (i *BacklogContext) createSpacePayload() *client.CreateSpacePayload {
	i.spaceName = "Test space" + uuid.NewV4().String()
	return &client.CreateSpacePayload{
		Data: &client.Space{
			Attributes: &client.SpaceAttributes{
				Name: &i.spaceName,
			},
			Type: "spaces",
		},
	}
}

/* Create new iteration */
func (i *BacklogContext) theUserCreatesANewIterationWithStartDateAndEndDate(startDate string, endDate string) error {
	a := i.api
	spaceIterationsPath := fmt.Sprintf("/api/spaces/%v/iterations", i.space.Data.ID)
	resp, err := a.c.CreateSpaceIterations(context.Background(), spaceIterationsPath, i.createSpaceIterationPayload(startDate, endDate))
	a.resp = resp
	a.err = err
	dec := json.NewDecoder(a.resp.Body)
	if err := dec.Decode(&i.iteration); err == io.EOF {
		return nil
	} else if err != nil {
		panic(err)
	}
	return nil
}

/* Payload for iteration */
func (i *BacklogContext) createSpaceIterationPayload(startDate string, endDate string) *client.CreateSpaceIterationsPayload {
	iterationName := "Test iteration"
	i.iterationName = iterationName
	const longForm = "2006-01-02"
	t1, _ := time.Parse(longForm, startDate)
	t2, _ := time.Parse(longForm, endDate)
	return &client.CreateSpaceIterationsPayload{
		Data: &client.Iteration{
			Attributes: &client.IterationAttributes{
				Name:    &iterationName,
				StartAt: &t1,
				EndAt:   &t2,
			},
			Type: "iterations",
		},
	}
}

/* Verify creation of new iteration */
func (i *BacklogContext) aNewIterationShouldBeCreated() error {
	createdIteration := i.iteration
	if len(createdIteration.Data.ID) < 1 {
		return fmt.Errorf("Expected an iteration with ID, but ID was [%s]", createdIteration.Data.ID)
	}
	expectedName := i.iterationName
	actualName := createdIteration.Data.Attributes.Name
	if *actualName != expectedName {
		return fmt.Errorf("Expected a space with title %s, but title was [%s]", expectedName, *actualName)
	}

	return nil
}

/* Add workitem to backlog */
func (b *BacklogContext) theUserAddsAnItemToTheBacklogWithTitleAndDescription() error {
	a := b.api
	resp, err := a.c.CreateWorkitem(context.Background(), "/api/workitems", createWorkItemPayload())
	a.resp = resp
	a.err = err
	json.NewDecoder(a.resp.Body).Decode(&a.body)
	mapError := mapstructure.Decode(a.body, &b.workItem)
	if mapError != nil {
		panic(mapError)
	}
	return nil
}

/* Payload is needed to create work item */
func createWorkItemPayload() *client.CreateWorkitemPayload {
	return &client.CreateWorkitemPayload{
		Data: &client.WorkItem2{
			Attributes: map[string]interface{}{
				workitem.SystemTitle: "Test bug",
				workitem.SystemState: workitem.SystemStateNew,
			},
			Relationships: &client.WorkItemRelationships{
				BaseType: &client.RelationBaseType{
					Data: &client.BaseTypeData{
						ID:   workitem.SystemBug,
						Type: "workitemtypes",
					},
				},
			},
			Type: "workitems",
		},
	}
}

/* Verify workitem was added to backlog */
func (i *BacklogContext) aNewWorkItemShouldBeCreatedInTheBacklog() error {
	createdWorkItem := i.workItem
	if len(*createdWorkItem.Data.ID) < 1 {
		return fmt.Errorf("Expected a work item with ID, but ID was [%p]", createdWorkItem.Data.ID)
	}
	expectedTitle := "Test bug"
	actualTitle := createdWorkItem.Data.Attributes["system.title"]
	if actualTitle != expectedTitle {
		return fmt.Errorf("Expected a work item with title %s, but title was [%s]", expectedTitle, actualTitle)
	}
	expectedState := "new"
	actualState := createdWorkItem.Data.Attributes["system.state"]
	if expectedState != actualState {
		return fmt.Errorf("Expected a work item with state %s, but state was [%s]", expectedState, actualState)
	}
	return nil
}

/* Verify user who created workitem */
func (i *BacklogContext) theCreatorOfTheWorkItemMustBeTheSaidUser() error {
	// TODO: Generate an identity for every call to /api/login/generate and verify the identity here against system.creator
	return godog.ErrPending
}

/* Define the logging levels */
var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

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

/* The payload used to create a workitem */
func createPayload() *client.CreateWorkitemPayload {
	return &client.CreateWorkitemPayload{
		Data: &client.WorkItem2{
			Attributes: map[string]interface{}{
				workitem.SystemTitle:   "the title",
				workitem.SystemState:   workitem.SystemStateOpen,
				workitem.SystemCreator: "GordieHowe",
			},
			Relationships: &client.WorkItemRelationships{
				BaseType: &client.RelationBaseType{
					Data: &client.BaseTypeData{
						ID:   "system.bug",
						Type: "workitemtypes",
					},
				},
			},
			Type: "workitems",
		},
	}
}

/* Copied from workitem_blackbox_test.go - needed to creaet the user which in turn is
   needed to assign the workitem */
func (s *WorkItem2Suite) SetupSuite() {
	var err error

	if err = configuration.Setup(""); err != nil {
		panic(fmt.Errorf("Failed to setup the configuration: %s", err.Error()))
	}

	s.db, err = gorm.Open("postgres", configuration.GetPostgresConfigString())
	if err != nil {
		panic("Failed to connect database: " + err.Error())
	}
	s.pubKey, _ = almtoken.ParsePublicKey([]byte(almtoken.RSAPublicKey))
	s.priKey, _ = almtoken.ParsePrivateKey([]byte(almtoken.RSAPrivateKey))
	s.svc = testsupport.ServiceAsUser("TestUpdateWI2-Service", almtoken.NewManager(s.pubKey, s.priKey), account.TestIdentity)
	require.NotNil(s.T(), s.svc)

	// s.wiCtrl = NewWorkitemController(s.svc, gormapplication.NewGormDB(s.db))
	// require.NotNil(s.T(), s.wiCtrl)
	//
	//	s.wi2Ctrl = NewWorkitemController(s.svc, gormapplication.NewGormDB(s.db))
	//	require.NotNil(s.T(), s.wi2Ctrl)

	// Make sure the database is populated with the correct types (e.g. system.bug etc.)
	if configuration.GetPopulateCommonTypes() {
		if err := models.Transactional(s.db, func(tx *gorm.DB) error {
			return migration.PopulateCommonTypes(context.Background(), tx, workitem.NewWorkItemTypeRepository(tx))
		}); err != nil {
			panic(err.Error())
		}
	}
	s.clean = gormsupport.DeleteCreatedEntities(s.db)
}

/* Copied from workitem_blackbox_test.go */
func (s *WorkItem2Suite) TearDownSuite() {
	s.clean()
	if s.db != nil {
		s.db.Close()
	}
}

/* Copied from workitem_blackbox_test.go */
func createOneRandomUserIdentity(ctx context.Context, db *gorm.DB) *account.Identity {
	newUserUUID := uuid.NewV4()
	identityRepo := account.NewIdentityRepository(db)
	identity := account.Identity{
		FullName: "Test User Integration Random",
		ImageURL: "http://images.com/42",
		ID:       newUserUUID,
	}
	err := identityRepo.Create(ctx, &identity)
	if err != nil {
		fmt.Println("should not happen off.")
		return nil
	}
	return &identity
}

/* Copied from workitem_blackbox_test.go */
func ident(id uuid.UUID) *client.GenericData {
	APIStringTypeUser := "identities"
	ut := APIStringTypeUser
	i := id.String()
	return &client.GenericData{
		Type: &ut,
		ID:   &i,
	}
}

/* Global variables */
var idString string

/* The payload used to update a workitem - to reassign the workitem */
func updatePayload() *client.UpdateWorkitemPayload {

	/* Copied from workitem_blackbox_test.go */
	bs := WorkItem2Suite{}
	bs.SetupSuite()
	newUser := createOneRandomUserIdentity(bs.svc.Context, bs.db)
	Info.Println(newUser)

	return &client.UpdateWorkitemPayload{
		Data: &client.WorkItem2{
			Attributes: map[string]interface{}{
				"version":            "0",
				workitem.SystemTitle: "the title updated",
				workitem.SystemState: workitem.SystemStateOpen,
			},
			Relationships: &client.WorkItemRelationships{
				BaseType: &client.RelationBaseType{
					Data: &client.BaseTypeData{
						ID:   "system.bug",
						Type: "workitemtypes",
					},
				},
				Assignees: &client.RelationGenericList{
					Data: []*client.GenericData{
						ident(newUser.ID),
					},
				},
			},
			ID:   &idString,
			Type: "workitems",
		},
	}
}

/* The payload used to update a workitem - to unassign the workitem*/
func updatePayloadUnassign() *client.UpdateWorkitemPayload {
	return &client.UpdateWorkitemPayload{
		Data: &client.WorkItem2{
			Attributes: map[string]interface{}{
				"version":            "1",
				workitem.SystemTitle: "the title updated again",
				workitem.SystemState: workitem.SystemStateOpen,
			},
			Relationships: &client.WorkItemRelationships{
				BaseType: &client.RelationBaseType{
					Data: &client.BaseTypeData{
						ID:   "system.bug",
						Type: "workitemtypes",
					},
				},
				Assignees: &client.RelationGenericList{
					Data: []*client.GenericData{},
				},
			},
			ID:   &idString,
			Type: "workitems",
		},
	}
}

/* Handle the GET/POST requests */
func (i *BacklogContext) iSendRequestTo(requestMethod, endpoint string) error {
	switch endpoint {
	case "get_workitemtypes":
		Info.Println("Received GET request to get workitem types")
		resp, err := i.api.c.ListWorkitemtype(context.Background(), "/api/workitemtypes", nil)
		i.api.resp = resp
		i.api.err = err
	case "get_workitems":
		Info.Println("Received GET request to get workitems")
		var tempString string
		tempString = "/api/workitems" + "/" + idString
		resp, err := i.api.c.ListWorkitem(context.Background(), tempString, nil, nil, nil, nil)
		i.api.resp = resp
		i.api.err = err
	case "create_workitem":
		Info.Println("Received POST request to create workitem")
		resp, err := i.api.c.CreateWorkitem(context.Background(), "/api/workitems", createPayload())
		i.api.resp = resp
		i.api.err = err

		defer i.api.resp.Body.Close()
		htmlData, err := ioutil.ReadAll(i.api.resp.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		data := string(htmlData)
		Info.Println("The response is:")
		Info.Println(data)

	case "update_workitem":
		Info.Println("Received POST request to update workitem")
		resp, err := i.api.c.UpdateWorkitem(context.Background(), "/api/workitems/"+idString, updatePayload())
		i.api.resp = resp
		i.api.err = err

		defer i.api.resp.Body.Close()
		htmlData, err := ioutil.ReadAll(i.api.resp.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		data := string(htmlData)
		Info.Println("The response is:")
		Info.Println(data)

	case "update_workitem_unassign":
		Info.Println("Received POST request to update/unassign workitem")
		resp, err := i.api.c.UpdateWorkitem(context.Background(), "/api/workitems/"+idString, updatePayloadUnassign())
		i.api.resp = resp
		i.api.err = err
		i.printResponse()

	case "delete_workitem":
		Info.Println("Received POST request to delete workitem")
		resp, err := i.api.c.DeleteWorkitem(context.Background(), "/api/workitems/"+idString)
		i.api.resp = resp
		i.api.err = err
	default:
		return godog.ErrPending
	}
	return nil
}

/* Check the value of the http responses */
func (i *BacklogContext) theResponseCodeShouldBe(statusCode int) error {
	if i.api.resp.StatusCode != statusCode {
		return fmt.Errorf("Expected %d but was %d", statusCode, i.api.resp.StatusCode)
	}
	return nil
}

/* Check the contents of the http responses */
func (i *BacklogContext) theResponseShouldContainFields(theDocString *gherkin.DocString) error {
	//fmt.Println("the string ", string(theDocString.Content))
	defer i.api.resp.Body.Close()
	htmlData, err := ioutil.ReadAll(i.api.resp.Body)
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

/* Authorized users require a token */
func (i *BacklogContext) imAuthorized() error {
	return i.generateToken()
}

/* A new work item is created for the test to use */
func (i *BacklogContext) setUpTestData2() {
	i.api.Reset()
	i.generateToken()
	Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	resp, err := i.api.c.CreateWorkitem(context.Background(), "/api/workitems", createPayload())
	//fmt.Println("body = ", resp.Body)
	//fmt.Println("error = ", resp.Status)
	i.api.resp = resp
	i.api.err = err

	defer i.api.resp.Body.Close()
	htmlData, err := ioutil.ReadAll(i.api.resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//fmt.Println(os.Stdout, string(htmlData))
	Info.Println("The newly created workitem is:", string(htmlData))

	idStart := strings.Index(string(htmlData), "\"id\":\"")
	tmpString := string(htmlData)[idStart+6:]
	idEnd := strings.Index(tmpString, "\"")
	idString = tmpString[:idEnd]
	Info.Println("The ID of the newly created workitem is:", idString)
}

/* Function to delete a work item - requires authorization */
func (i *BacklogContext) cleanUpTestData2() {
	fmt.Println("Nothing to see here - move along")

	i.api.Reset()
	i.generateToken()

	/* Delete the workitem */
	Info.Println("The ID of the workitem to be deleted is:", idString)
	resp, err := i.api.c.DeleteWorkitem(context.Background(), "/api/workitems/"+idString)
	i.api.resp = resp
	i.api.err = err
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

/* Formatted printing of the http response */
func (i *BacklogContext) printResponse() {
	//fmt.Println("Nothing to see here - move along")

	defer i.api.resp.Body.Close()
	htmlData, err := ioutil.ReadAll(i.api.resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	data := string(htmlData)
	Info.Println("The response is:")
	Info.Println(data)
}
