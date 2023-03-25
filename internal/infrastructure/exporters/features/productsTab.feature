Feature: Products tab functionality 

Background:
    Given I am logged in to creation portal

Scenario: Getting article info data from product tab
When the user switches to the "model" view with basic filter
Then the model info for the APP product should be displayed
When the user clicks on the first product in the "table" view on Product Page
Then the Product Details Page should be loaded