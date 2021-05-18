# KCE Cloud Controller Manager

[![Test Coverage](https://api.codeclimate.com/v1/badges/464449b5a40461b8072a/test_coverage)](https://codeclimate.com/github/krystal/kce-ccm/test_coverage)

The Katapult Container Engine Cloud Controller Manager is a set of tools that is designed to monitor nodes & provision load balancers.

This is still a work in progress.

## Other CCMs

See the following other CCMs as good guidance:

- https://github.com/kubernetes/cloud-provider-gcp/
- https://github.com/kubernetes/cloud-provider-aws/  
- https://github.com/digitalocean/digitalocean-cloud-controller-manager (old-style, take with pinch of salt)

## Configuration

The following environment variables are mandatory:

* `KATAPULT_API_TOKEN` - the API token to use to authenticate with
* `KATAPULT_ORGANIZATION_RID` - the organization RID for the cluster
* `KATAPULT_DATA_CENTER_RID` - the data centre that the cluster is deployed in
* `KATAPULT_NODE_TAG_RID` - the tag that has been applied to all worker nodes in the cluster

The following environment variables are optional:

* `KATAPULT_API_HOST` - the hostname for the API service

A set of command line arguments are also available. Use --help to view these in full.

## Token

The token requires the following scopes:

- ``load_balancers``