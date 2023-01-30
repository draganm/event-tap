Feature: mapping events


    Scenario: mapping events
        Given one event in the buffer
        When I create a new map of events
        Then the receiver should receive that event as webhook

