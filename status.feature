# file: status.feature
Feature: get status
  In order to know the state of the running server
  As an API user
  I need to be able to request current status

  Scenario: should get work item type fields
    When I send "GET" request to "get_workitemtypes"
    Then the response code should be 200
    And the response should contain fields:
      """
      {
      "name":"system.bug"
      }
      """

  Scenario: should get work item fields
    When I send "GET" request to "get_workitems"
    Then the response code should be 200
    And the response should contain fields:
      """
      {
        "system.title":"remove this workitem"
        "system.creator":"jsmith"
      }
      """

  Scenario: should create work item 
    When I'm authorized
    When I send "POST" request to "create_workitem"
    Then the response code should be 201
    And the response should contain fields:
      """
      {
			"system.title":   "remove this workitem PLEASE",
			"system.owner":   "ldimaggi",
			"system.state":   "open",
			"system.creator": "ldimaggi",
      }
      """