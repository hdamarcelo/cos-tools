/*
Connector Service Fleet Manager Admin APIs

Connector Service Fleet Manager Admin is a Rest API to manage connector clusters.

API version: 0.0.3
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package admin

// ConnectorDeploymentAdminStatusOperators struct for ConnectorDeploymentAdminStatusOperators
type ConnectorDeploymentAdminStatusOperators struct {
	Assigned  ConnectorOperator `json:"assigned,omitempty"`
	Available ConnectorOperator `json:"available,omitempty"`
}