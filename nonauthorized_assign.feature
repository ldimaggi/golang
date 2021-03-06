Feature:
Workitems are assigned to users. It must be possible to set, remove and update workitem assignments in the Detail view.
As an unidentified user, I should be able to only view a user assignment. As an identified user, it must be possible to set, remove and update workitem assignments.

  Scenario: Nonauthorized user is able to view workitem assignments (#342)
    When I'm not authorized
    When I send "GET" request to "get_workitems" 
    Then the response code should be 200
    And the response should contain fields:
      """
      {
			"assignee":"WayneGretzky"
      }
      """

  Scenario: Nonauthorized user is not able to create a new workitem (#341)   
    When I'm not authorized
    When I send "POST" request to "create_workitem" 
    Then the response code should be 401

  Scenario: Nonauthorized user is not able to reassign/modify a workitem (#341)   
    When I'm not authorized
    When I send "POST" request to "update_workitem" 
    Then the response code should be 401

  Scenario: Nonauthorized user is not able to delete a workitem (#341)   
    When I'm not authorized
    When I send "POST" request to "delete_workitem" 
    Then the response code should be 401

