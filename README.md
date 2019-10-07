<a id="markdown-AWSConfig Transformer" name="AWSConfig Transformer"></a>
# AWS Config Transformer - a lambda handler that receives AWS Config changes and returns a transformed version of the Config change
[![GoDoc](https://godoc.org/github.com/asecurityteam/awsconfig-transformerd?status.svg)](https://godoc.org/github.com/asecurityteam/awsconfig-transformerd)

<https://github.com/asecurityteam/awsconfig-transformer>

  - [Overview](#overview)
  - [Quick Start](#quick-start)
  - [Configuration](#configuration)
  - [Supported Resources](#supported-resources)
  - [Status](#status)
  - [Contributing](#contributing)
      - [Building And Testing](#building-and-testing)
      - [Quality Gates](#quality-gates)
      - [License](#license)
      - [Contributing Agreement](#contributing-agreement)

<!-- /TOC -->

<a id="markdown-overview" name="overview"></a>
## Overview<!-- TOC -->

AWS Config provides a detailed view of the configuration of AWS resources, potentially across
multiple AWS accounts, and can provide a stream of configuration change events via an SNS topic
which publishes to SQS.  The awsconfig-transformerd service provides a lambda handler which accepts
the [configuration item change notification](https://docs.aws.amazon.com/config/latest/developerguide/example-sns-notification.html) payload,
extracts the changed network information, and returns a transformed version of the configuration change.
The goal of the transformation is to highlight changes to the network interfaces associated with AWS resources.

Example topic payloads can be seen at AWS's Developer Guide page, but beware, the data is old.  To
gain a complete understanding of the variances in notification payloads, it is recommended to
gather real notifications of actual change events.

For the current list of supported resource types, see [Supported Resources](#supported-resources).
All other config change types are ignored. The payload emitted from this transformer adheres to the
following JSON specification:

```
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "Schema for cloud asset change events",
  "properties": {
    "changeTime": {
      "type": "string",
      "title": "time at which the asset change occurred",
      "format": "date-time"
    },
    "resourceType": {
      "type": "string",
      "title": "the AWS resource type"
    },
    "accountId": {
      "type": "string",
      "title": "the ID of the AWS account"
    },
    "region": {
      "type": "string",
      "title": "the AWS region"
    },
    "resourceId": {
      "type": "string",
      "title": "the ID of the AWS resource"
    },
    "tags": {
      "type": "object",
      "title": "AWS tags",
      "additionalProperties": {
        "type": "string"
      }
    },
    "changes": {
      "type": "array",
      "title": "list of changes which occurred",
      "items": {
        "type": "object",
        "properties": {
          "publicIpAddresses": {
            "type": "array",
            "title": "public IP addresses for the asset",
            "items": {
              "type": "string"
            }
          },
          "privateIpAddresses": {
            "type": "array",
            "title": "private IP addresses for the asset",
            "items": {
              "type": "string"
            }
          },
          "hostnames": {
            "type": "array",
            "title": "hostnames of the asset",
            "items": {
              "type": "string"
            }
          },
          "changeType": {
            "type": "string",
            "title": "the type of change which occurred",
            "enum": ["ADDED", "DELETED"]
          }
        }
      }
    }
  },
  "required": [
    "changeTime",
    "resourceType",
    "accountId",
    "region",
    "resourceId",
    "tags",
    "changes"
  ]
}
```


<a id="markdown-quick-start" name="quick-start"></a>
## Quick Start

Install docker and docker-compose.

The app can be run locally by running `make run`.

This will run `docker-compose` for the serverfull project
as well as the supplied serverfull-gateway configuration.
The sample configration provided assumes there will be a stats
collector running. To disable this, remove the stats configuration
lines from the server configuration and the serverfull-gateway
configuration.

The app should now be running on port 8080.

`curl -vX POST "http://localhost:8080" -H "Content-Type:application/json" -d @pkg/handlers/v1/testdata/ec2.0.json`

<a id="markdown-configuration" name="configuration"></a>
## Configuration

Images of this project are built, and hosted on [DockerHub](https://cloud.docker.com/u/asecurityteam/repository/docker/asecurityteam/awsconfig-transformerd).

This code functions as a stand-alone Lambda function, and can be deployed to AWS Lambda directly. To run in the AWS Lambda environment,
create a new Go project, import this project as a dependency, and run the lambda using the aws-lambda-sdk:

```
func main() {
  transformer := &v1.Transformer{
		LogFn:  <LOGGER_PROVIDER>,
		StatFn: <STATS_PROVIDER>,
	}
  lambda.Start(transformer.Handle)
}
```

For those who do not have access to AWS Lambda, you can run your own configuration by composing this
image with your own custom configuration of serverfull-gateway.

### Logging

This project makes use of [logevent](https://github.com/asecurityteam/logevent) which provides structured logging
using Go structs and tags. By default the project will set a logger value in the context for each request. The handler
uses the `LogFn` function defined in `pkg/domain/alias.go` to extract the logger instance from the context.

The built in logger can be configured through the serverfull runtime [configuration](https://github.com/asecurityteam/serverfull#configuration).

### Stats

This project uses [xstats](https://github.com/rs/xstats) as its underlying stats library. By default the project will
set a stat client value in the context for each request. The handler uses the `StatFn` function defined in
`pkg/domain/alias.go` to extract the logger instance from the context.

The built in stats client can be configured through the serverfull runtime [configuration](https://github.com/asecurityteam/serverfull#configuration).

Additional resources:

* [serverfull](https://github.com/asecurityteam/serverfull)
* [serverfull-gateway](https://github.com/asecurityteam/serverfull-gateway)

<a id="markdown-supported-resources" name="supported-resources"></a>
## Supported Resources

The current version only supports extracting network changes from:
* EC2 instances
* Elastic Load Balancers
* Application Load Balancers

<a id="markdown-status" name="status"></a>
## Status

This project is in incubation which means we are not yet operating this tool in production
and the interfaces are subject to change.

<a id="markdown-contributing" name="contributing"></a>
## Contributing

If you are interested in contributing to the project, feel free to open an issue or PR.

<a id="markdown-building-and-testing" name="building-and-testing"></a>
### Building And Testing

We publish a docker image called [SDCLI](https://github.com/asecurityteam/sdcli) that
bundles all of our build dependencies. It is used by the included Makefile to help make
building and testing a bit easier. The following actions are available through the Makefile:

-   make dep

    Install the project dependencies into a vendor directory

-   make lint

    Run our static analysis suite

-   make test

    Run unit tests and generate a coverage artifact

-   make integration

    Run integration tests and generate a coverage artifact

-   make coverage

    Report the combined coverage for unit and integration tests

-   make build

    Generate a local build of the project (if applicable)

-   make run

    Run a local instance of the project (if applicable)

-   make doc

    Generate the project code documentation and make it viewable
    locally.

<a id="markdown-quality-gates" name="quality-gates"></a>
### Quality Gates

Our build process will run the following checks before going green:

-   make lint
-   make test
-   make integration
-   make coverage (combined result must be 85% or above for the project)

Running these locally, will give early indicators of pass/fail.

<a id="markdown-license" name="license"></a>
### License

This project is licensed under Apache 2.0. See LICENSE.txt for details.

<a id="markdown-contributing-agreement" name="contributing-agreement"></a>
### Contributing Agreement

Atlassian requires signing a contributor's agreement before we can accept a
patch. If you are an individual you can fill out the
[individual CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d).
If you are contributing on behalf of your company then please fill out the
[corporate CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b).
