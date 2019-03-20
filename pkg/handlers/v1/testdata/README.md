# testdata #

This folder contains real test data pulled from config to ensure we correctly interpret various scenarios produced by AWS resources.

## EC2 ##

The files ec2.[0-4].json are the result of an EC2 instance being started for the first time, stopped, restarted, stopped, and then terminated in that order.
These lifecycle events result in different diff events: CREATE, UPDATE, DELETE. For each type of event, we need to inspect the current resource configuration
as well as the previous configuration contained in the diff. An example of this is shown below for ec2

```
{
  "configurationItemDiff": {
    "changedProperties": { ... }, <- Configuration diff where the keys are named after the resource configuration, and show a previous and updated value for each changed configuration property
    "changeType": "<CREATE|UPDATE|DELETE>"
  },
  "configurationItem": {
    "relatedEvents": [],
    "relationships": [ ... ], <- resources related to the current resource (e.g. volumes, network interfaces, subnets, security groups, etc.)
    "configuration": { <- EC2 specific configuration
      "amiLaunchIndex": 0,
      "imageId": "ami-0bbe6b35405ecebdb",
      "instanceId": "i-0a763ac3ee37d8d2b",
      "instanceType": "t2.micro",
      "kernelId": null,
      "keyName": "zactest2",
      "launchTime": "2019-02-22T20:50:14.000Z",
      "monitoring": {
        "state": "disabled"
      },
      "placement": {
        "availabilityZone": "us-west-2a",
        "affinity": null,
        "groupName": "",
        "partitionNumber": null,
        "hostId": null,
        "tenancy": "default",
        "spreadDomain": null
      },
      "platform": null,
      "privateDnsName": "ip-172-31-30-79.us-west-2.compute.internal",
      "privateIpAddress": "172.31.30.79",
      "productCodes": [],
      "publicDnsName": "ec2-34-219-72-29.us-west-2.compute.amazonaws.com",
      "publicIpAddress": "34.219.72.29",
      "ramdiskId": null,
      "state": {
        "code": 16,
        "name": "running"
      },
      "stateTransitionReason": "",
      "subnetId": "subnet-3d0b8c5a",
      "vpcId": "vpc-b290fcd5",
      "architecture": "x86_64",
      "blockDeviceMappings": [
        {
          "deviceName": "/dev/sda1",
          "ebs": {
            "attachTime": "2019-02-22T20:30:11.000Z",
            "deleteOnTermination": true,
            "status": "attached",
            "volumeId": "vol-0da7faa5400c54c4c"
          }
        }
      ],
      "clientToken": "",
      "ebsOptimized": false,
      "enaSupport": true,
      "hypervisor": "xen",
      "iamInstanceProfile": null,
      "instanceLifecycle": null,
      "elasticGpuAssociations": [],
      "elasticInferenceAcceleratorAssociations": [],
      "networkInterfaces": [
        {
          "association": {
            "ipOwnerId": "amazon",
            "publicDnsName": "ec2-34-219-72-29.us-west-2.compute.amazonaws.com",
            "publicIp": "34.219.72.29"
          },
          "attachment": {
            "attachTime": "2019-02-22T20:30:10.000Z",
            "attachmentId": "eni-attach-0195433bad822bc2f",
            "deleteOnTermination": true,
            "deviceIndex": 0,
            "status": "attached"
          },
          "description": "",
          "groups": [
            {
              "groupName": "launch-wizard-2",
              "groupId": "sg-05d1b37a375dcca8e"
            }
          ],
          "ipv6Addresses": [],
          "macAddress": "02:17:59:ed:8b:0a",
          "networkInterfaceId": "eni-05721fa8354d07b8c",
          "ownerId": "515665915980",
          "privateDnsName": "ip-172-31-30-79.us-west-2.compute.internal",
          "privateIpAddress": "172.31.30.79",
          "privateIpAddresses": [
            {
              "association": {
                "ipOwnerId": "amazon",
                "publicDnsName": "ec2-34-219-72-29.us-west-2.compute.amazonaws.com",
                "publicIp": "34.219.72.29"
              },
              "primary": true,
              "privateDnsName": "ip-172-31-30-79.us-west-2.compute.internal",
              "privateIpAddress": "172.31.30.79"
            }
          ],
          "sourceDestCheck": true,
          "status": "in-use",
          "subnetId": "subnet-3d0b8c5a",
          "vpcId": "vpc-b290fcd5"
        }
      ],
      "rootDeviceName": "/dev/sda1",
      "rootDeviceType": "ebs",
      "securityGroups": [
        {
          "groupName": "launch-wizard-2",
          "groupId": "sg-05d1b37a375dcca8e"
        }
      ],
      "sourceDestCheck": true,
      "spotInstanceRequestId": null,
      "sriovNetSupport": null,
      "stateReason": null,
      "tags": [
        {
          "key": "business_unit",
          "value": "CISO-Security"
        },
        {
          "key": "service_name",
          "value": "foo-bar"
        }
      ],
      "virtualizationType": "hvm",
      "cpuOptions": {
        "coreCount": 1,
        "threadsPerCore": 1
      },
      "capacityReservationId": null,
      "capacityReservationSpecification": null,
      "hibernationOptions": {
        "configured": false
      },
      "licenses": []
    },
    "supplementaryConfiguration": {},
    "tags": {
      "service_name": "foo-bar",
      "business_unit": "CISO-Security"
    },
    "configurationItemVersion": "1.3",
    "configurationItemCaptureTime": "2019-02-22T21:02:18.758Z",
    "configurationStateId": 1550869338758,
    "awsAccountId": "515665915980",
    "configurationItemStatus": "OK",
    "resourceType": "AWS::EC2::Instance",
    "resourceId": "i-0a763ac3ee37d8d2b",
    "resourceName": null,
    "ARN": "arn:aws:ec2:us-west-2:515665915980:instance/i-0a763ac3ee37d8d2b",
    "awsRegion": "us-west-2",
    "availabilityZone": "us-west-2a",
    "configurationStateMd5Hash": "",
    "resourceCreationTime": "2019-02-22T20:50:14.000Z"
  },
  "notificationCreationTime": "2019-02-22T21:02:19.690Z",
  "messageType": "ConfigurationItemChangeNotification",
  "recordVersion": "1.3"
}
```