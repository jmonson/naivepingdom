# Naive Pingdom

Pingdom...sort of ;)

## Getting Started

This project probes websites to monitor response duration, content type and response codes. The Probe service provides a `/metrics` endpoint which is consumed by Prometheus. Grafana then provides visualizations of the site response duration as well as the latest content-type and HTTP response code.

The Probe service has a RESTful API endpoint (`/api/site`) which allows adding, removing and viewing monitored sites. Grafana will automatically incorporate any site changes made via the API.

### Prerequisites

* [Docker](https://docs.docker.com/v17.09/engine/installation/)
* [Go](https://golang.org/doc/install)

## Deployment

* The project folder should be deployed within $GOPATH/src
* Build custom Prometheus collector

```
$ docker-compose build
```

* Launch services

```
$ docker-compose up -d
```

## Usage

Service locations:
* Probe metrics endpoint - http://localhost:8080/metrics
* Probe API endpoint - http://localhost:8080/api/site
* Prometheus - http://localhost:9090
* Grafana - http://localhost:3000/login

1. Login to [Grafana](http://localhost:3000/login)
2. View `Site Probe` dashboard
3. Add new site:

```
curl --request POST \
  --url http://localhost:8080/api/site \
  --header 'content-type: application/json' \
  --data '{"name":"twitter","address":"https://twitter.com"}'
```
4. Refresh Grafana

## Destory

```
$ docker-compose stop
$ docker-compose down -v
```

## Notes

I chose to implement Exporter/API in Go. I started learning Go last week with the goal to build a custom Kubernetes/Terraform client to manage deployments. Therefore, the code is a bit novice and there are many areas to be improved (some detailed below).

Rather than build the Grafana dashboard configuration from scratch, I took a shortcut to export the JSON and kept the output as-is. I made the dashboards dynamic/templated, such that any new additional metrics will automatically appear, but I would have liked to spend more time refining this configuration.

The Probe service will persist the list of sites to monitor to a flat JSON file on disk. I added this so configurations were saved across multiple launches of the application.

## Future Enhancements

- Test Driven Development! Normally, I would have built unit tests prior to coding. However, I skipped this important step in the interest of saving time.
- HTTP Tracing for more accurate response times. My response duration is more of an estimate. A better measure would have utilized the [httptrace](https://golang.org/pkg/net/http/httptrace/) package to understand true response times.
- Production HTTP service:
  - Security / Authentication
  - MVC pattern with Router
- Adequate datastore 
- Terraform / Kubernetes
- Swagger. Implement a swagger library to provide an API documentation endpoint
