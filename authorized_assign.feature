Feature:
Workitems are assigned to users. It must be possible to set, remove and update workitem assignments in the Detail view.
As an unidentified user, I should be able to only view a user assignment. As an identified user, it must be possible to set, remove and update workitem assignments.

  Scenario: Authorized user is able to view workitem assignments (#342)
    When I'm authorized
    When I send "GET" request to "get_workitems" 
    Then the response code should be 200
    And the response should contain fields:
      """
      {
			"assignee":"WayneGretzky"
      }
      """

  Scenario: Authorized user is able to reassign a workitem (#341)   
    When I'm authorized
    When I send "POST" request to "update_workitem" 
    Then the response code should be 200

  Scenario: Authorized user is able to unassign a workitem (#341)   
    When I'm authorized
    When I send "POST" request to "update_workitem_unassign" 
    Then the response code should be 200

  Scenario: Authorized user is able to delete a workitem (#341)   
    When I'm authorized
    When I send "POST" request to "delete_workitem" 
    Then the response code should be 200

