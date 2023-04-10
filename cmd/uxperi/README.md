 # Uxperi a Cucumber based Prometheus Exporter

`Uxperi` is a tool designed to monitor the performance and functionality of a website. This exporter execute predefined test scenarios, as defined in Cucumber feature files, and generate metrics that describe the latency of each test, as well as the test result (pass or fail). These metrics could be used to track the performance and functionality of the website over time, identify areas that need improvement, and trigger alerts if certain thresholds are exceeded.

One of the key features of the application is its Prometheus exporter, which is based on Gherkin tests. The exporter allows users to monitor the health and performance of the system using Prometheus, a popular open-source monitoring and alerting system. The exporter works by running godog test suites, which are written in Gherkin syntax and describe the behavior of the system in a human-readable format. The exporter then converts the results of the test suites into Prometheus metrics, which can be visualized and analyzed using a variety of tools.

## Available metrics

 * `stepDurationGaugeVec`: This is a Gauge vector that measures the duration of test steps in seconds. It has four label dimensions: feature_name, scenario_name, step_name, and step_status. The feature_name and scenario_name labels identify the feature file and scenario that the step belongs to, while the step_name label identifies the name of the step itself. The step_status label indicates whether the step passed or failed. This metric can be used to identify slow-running or failing steps in the test suite.

* `stepSuccessGaugeVec`: This is a Gauge vector that displays whether or not the test was a success. It has two label dimensions: feature_name and scenario_name. The feature_name and scenario_name labels identify the feature file and scenario that the test belongs to. The value of the metric is 1 if the test succeeded, and 0 if it failed. This metric can be used to track the overall success rate of the test suite over time.

## How to add a new plugin

1. Features folder: The application expects a folder containing all the Gherkin feature definitions. These feature files describe the behavior of the system in a human-readable format.

2. Implementing steps: The application uses the godog library to implement the steps defined in the feature files. The steps are the actions that the system must perform in order to satisfy the requirements described in the feature files.

3. Registering tests: Once the steps have been implemented, the tests must be registered with the CLI application. This involves defining a new command that will execute the tests and output the results.

4. Running tests: Once the tests have been registered, they can be run by making a GET request to the exporter. The exporter will execute the tests and return the results in a format that can be consumed by Prometheus or other monitoring systems.

## Run docker image

```
docker run -p 8082:8082 -p 8081:8081 --mount type=bind,source=./internal/infrastructure/exporters/features/login.feature,target=/app/features/login.feature,readonly danifv27/uxperi:local test --logging.level=debug --test.features-folder="/app/features"
```
