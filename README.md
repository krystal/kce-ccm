# KCE Cloud Controller Manager

[![Test Coverage](https://api.codeclimate.com/v1/badges/464449b5a40461b8072a/test_coverage)](https://codeclimate.com/github/krystal/kce-ccm/test_coverage)
[![Go Report Card](https://goreportcard.com/badge/github.com/krystal/kce-ccm)](https://goreportcard.com/report/github.com/krystal/kce-ccm)
[![GitHub last commit](https://img.shields.io/github/last-commit/krystal/go-katapult.svg?style=flat&logo=github&logoColor=white)](https://github.com/krystal/go-katapult/commits/main)
[![GitHub issues](https://img.shields.io/github/issues-raw/krystal/go-katapult.svg?style=flat&logo=github&logoColor=white)](https://github.com/krystal/go-katapult/issues)

The Katapult Container Engine Cloud Controller Manager is a set of tools that is
designed to monitor nodes & provision load balancers.

As it stands, kce-ccm creates LoadBalancers and LoadBalancerRule objects in
Katapult to direct traffic to k8s LoadBalancer type services.

## Other CCMs

See the following other CCMs as good guidance:

- https://github.com/kubernetes/cloud-provider-gcp/
- https://github.com/kubernetes/cloud-provider-aws/  
- https://github.com/digitalocean/digitalocean-cloud-controller-manager
  (uses older CCM framework, so take with a pinch of salt)

## Configuration

The following environment variables are mandatory:

* `KATAPULT_API_TOKEN` - the API token to use to authenticate with
* `KATAPULT_ORGANIZATION_RID` - the organization RID for the cluster
* `KATAPULT_DATA_CENTER_RID` - the data centre that the cluster is deployed in
* `KATAPULT_NODE_TAG_RID` - the tag that has been applied to all worker nodes in
  the cluster

The following environment variables are optional:

* `KATAPULT_API_HOST` - the hostname for the API service

A set of command line arguments are also available. Use --help to view these in
full.

## Token

The token requires the following scopes:

- ``load_balancers``