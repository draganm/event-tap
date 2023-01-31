Feature: Feature name

    Scenario: listing when there are no taps
        Given there are no taps
        When I list the taps
        Then the result should be empty

    Scenario: listing one tap
        Given there is one tap
        When I list the taps
        Then the result should have one tap