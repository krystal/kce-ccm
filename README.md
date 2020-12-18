# KCE Cloud Controller Manager

The Katapult Container Engine Cloud Controller Manager is a set of tools that is designed to monitor nodes & provision load balancers.

This is still a work in progress.

## Configuration

The following environment variables are needed in order to communicate with the Katapult backend...

* `KATAPULT_API_HOST` - the hostname for the API service
* `KATAPULT_API_TOKEN` - the API token to use to authenticate with
* `KATAPULT_ORGANIZATION_RID` - the organization RID for the cluster
* `KATAPULT_DATA_CENTER_RID` - the data centre that the cluster is deployed in
* `KATAPULT_NETWORK_RID` - the network that the cluster is deployed on
