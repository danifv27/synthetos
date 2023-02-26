 # Uxperi a Cucumber based Prometheus Exporter

`Uxperi` is a tool designed to monitor the performance and functionality of a website. This exporter execute predefined test scenarios, as defined in Cucumber feature files, and generate metrics that describe the latency of each test, as well as the test result (pass or fail). These metrics could be used to track the performance and functionality of the website over time, identify areas that need improvement, and trigger alerts if certain thresholds are exceeded.

Similar to the blackbox exporter, this exporter operate from outside the system being monitored. It sends HTTP requests to the website being monitored, execute the Cucumber-defined test scenarios, and collect metrics related to the test results and the time it takes for the tests to complete. The exporter publish these metrics to Prometheus so that they can be visualized and used for monitoring and alerting.

The metrics collected by this exporter could include the total number of tests executed, the number of tests that passed or failed, the average latency of each test, and the total execution time of the tests.

Overall, a Prometheus exporter that runs Cucumber-defined tests against a website and publishes its results to Prometheus is a powerful tool for monitoring the performance and functionality of a website, and can help ensure that the website is meeting its performance and quality goals.
