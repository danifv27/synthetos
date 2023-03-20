Feature: Products tab functionality 

Background:
    Given I am logged in to creation portal

Scenario: Getting factory info data from product tab
    When I switch to "model factories" view on Product Page
    Then I can see factory info for APP product
    And user clicks on first product in "table" view on Product Page
    Then user sees Factory Details Page is loaded