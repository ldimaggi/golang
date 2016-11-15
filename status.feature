# file: status.feature
Feature: get status
  In order to know the state of the running server
  As an API user
  I need to be able to request current status

  Scenario: should get workitem types
    When I send "GET" request to "workitemtypes"
    Then the response code should be 200
    And the response should contain json:
      """
      {
        "fields":""
      }
      """

  Scenario: unauthorized access to GET
    When I send "GET" request to "status"
    Then the response code should be 200

  Scenario: should get commit sha
    When I send "GET" request to "status"
    Then the response code should be 200
    And the response should contain json:
      """
      {
        "commit": ""
      }
      """
  Scenario: should get build time
    When I send "GET" request to "status"
    Then the response code should be 200
    And the response should contain json:
      """
      {
        "buildTime": ""
      }
      """
  Scenario: should get start time
    When I send "GET" request to "status"
    Then the response code should be 200
    And the response should contain json:
      """
      {
        "startTime": ""
      }
      """

