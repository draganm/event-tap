Feature: deleting taps

    Scenario: deleting an existing tap
        Given there is one tap
        When I delete the tap
        Then the list of taps should not contain the deleted tap