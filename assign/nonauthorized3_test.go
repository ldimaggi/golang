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
	"os"
	"testing"

	"github.com/DATA-DOG/godog"
)

func FeatureContext(s *godog.Suite) {
	backlogCtx := BacklogContext{identityHelper: IdentityHelper{}, api: Api{}}
	s.BeforeSuite(backlogCtx.setUpTestData2)
	s.BeforeScenario(backlogCtx.Reset)
	s.Step(`^I\'m not authorized$`, backlogCtx.imNotAuthorized)
	s.Step(`^I send "([^"]*)" request to "([^"]*)"$`, backlogCtx.iSendRequestTo)
	s.Step(`^the response code should be (\d+)$`, backlogCtx.theResponseCodeShouldBe)
	s.Step(`^the response should contain fields:$`, backlogCtx.theResponseShouldContainFields)
	s.AfterSuite(backlogCtx.cleanUpTestData2)
}

func TestMain(m *testing.M) {
	status := godog.RunWithOptions("godogs", func(s *godog.Suite) {
		FeatureContext(s)
	}, godog.Options{
		Format: "progress",
		Paths:  []string{"../../../features/assign/nonauthorized"},
	})

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}
