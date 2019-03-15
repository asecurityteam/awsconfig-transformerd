<a id="markdown-AWSConfig Transformer" name="AWSConfig Transformer"></a>
# AWS Config Transformer - an service that receives AWS config changes and transforms them to a more consumable form, POSTing to a configured endpoint

<https://github.com/asecurityteam/awsconfig-transformer>

- [AWSConfig Transformer - description](#AWSConfig Transformer)
    - [Overview](#overview)
    - [Quick Start](#quick-start)
    - [Configuration](#configuration)
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
which publishes to SQS.  The awsconfig-transformerd service provides an API which accepts the SQS
payload, transforms the data, and POSTs the transformed data to an HTTP endpoint, where they may
be further transformed or otherwise processed and consumed by other services.

Example topic payloads can be seen at AWS's Developer Guide page, but beware, the data is old.  To
gain a complete understanding of the variances in notification payloads, it is recommended to
gather real notifications of actual change events.

The current implementation of this filter only observes `"changeType": "UPDATE"` events, and only
records and transforms `Configuration.NetworkInterfaces.*` changes.  All other config change types
are ignored.  The payload emitted from this transformer adheres to the following JSON specification:

```
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "Schema for cloud asset change events",
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
    "startedAt": {
      "type": "string",
      "title": "time at which the asset is discovered",
      "format": "date-time"
    },
    "stoppedAt": {
      "type": "string",
      "title": "time at which the asset is offline",
      "format": "date-time"
    },
    "resourceType": {
      "type": "string",
      "title": "the AWS resource type"
    },
    "businessUnit": {
      "type": "string",
      "title": "the business unit to which the asset belongs"
    },
    "resourceOwner": {
      "type": "string",
      "title": "the asset owner"
    },
    "serviceName": {
      "type": "string",
      "title": "the name of the related service"
    },
    "microsServiceId": {
        "type": "string",
        "title": "the ID of the service in micros"
    }
  },
  "required": [
    "businessUnit",
    "resourceOwner",
    "serviceName"
  ]
}
```


<a id="markdown-quick-start" name="quick-start"></a>
## Quick Start

<Hello world style example.>

<a id="markdown-configuration" name="configuration"></a>
## Configuration

<Details of how to actually work with the project>

<a id="markdown-status" name="status"></a>
## Status

This project is in incubation which means we are not yet operating this tool in production
and the interfaces are subject to change.

<a id="markdown-contributing" name="contributing"></a>
## Contributing

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
