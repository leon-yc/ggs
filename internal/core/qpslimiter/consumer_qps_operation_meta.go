package qpslimiter

import (
	"strings"

	"github.com/leon-yc/ggs/internal/core/common"
)

// ConsumerKeys contain consumer keys
type ConsumerKeys struct {
	MicroServiceName string
	//SchemaQualifiedName    string
	OperationQualifiedName string
}

// ProviderKeys contain provider keys
type ProviderKeys struct {
	Api    string
	Global string
}

//Prefix is const
const Prefix = "ggs.flowcontrol"

// GetConsumerKey get specific key for consumer
func GetConsumerKey(sourceName, serviceName, schemaID, OperationID string) *ConsumerKeys {
	keys := new(ConsumerKeys)
	//for mesher to govern
	if serviceName != "" {
		keys.MicroServiceName = strings.Join([]string{Prefix, common.Consumer, "qps.limit", serviceName}, ".")
	}
	//if schemaID != "" {
	//	keys.SchemaQualifiedName = strings.Join([]string{keys.MicroServiceName, schemaID}, ".")
	//}
	if OperationID != "" {
		keys.OperationQualifiedName = strings.Join([]string{keys.MicroServiceName, OperationID}, ".")
	}
	return keys
}

// GetProviderKey get specific key for provider
func GetProviderKey(operationID string) *ProviderKeys {
	keys := &ProviderKeys{}
	if operationID != "" {
		keys.Api = strings.Join([]string{Prefix, common.Provider, "qps.limit", operationID}, ".")
	}

	keys.Global = strings.Join([]string{Prefix, common.Provider, "qps.global.limit"}, ".")
	return keys
}

// GetMicroServiceSchemaOpQualifiedName get micro-service schema operation qualified name
func (op *ConsumerKeys) GetMicroServiceSchemaOpQualifiedName() string {
	return op.OperationQualifiedName
}

// GetMicroServiceName get micro-service name
func (op *ConsumerKeys) GetMicroServiceName() string {
	return op.MicroServiceName
}
