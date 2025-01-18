// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	liberrors "errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/lib/vapi/std"
	"github.com/vmware/vsphere-automation-sdk-go/lib/vapi/std/errors"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data/serializers/cleanjson"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func printAPIError(apiError model.ApiError) string {
	msg := ""

	if apiError.ErrorMessage != nil {
		msg = *apiError.ErrorMessage
	}

	if apiError.Details != nil {
		msg += fmt.Sprintf(": %s", *apiError.Details)
	}

	if apiError.ErrorCode != nil {
		msg += fmt.Sprintf(" (code %v)", *apiError.ErrorCode)
	}

	return msg
}

// TODO: Remove duplicate code when sdk implements composition of API inheritance model
func printRelatedAPIError(apiError model.RelatedApiError) string {
	if apiError.ErrorMessage != nil && apiError.ErrorCode != nil {
		return fmt.Sprintf("%s (code %v)", *apiError.ErrorMessage, *apiError.ErrorCode)
	}

	if apiError.ErrorMessage != nil {
		return *apiError.ErrorMessage
	}

	if apiError.ErrorCode != nil {
		return fmt.Sprintf("(code %v)", *apiError.ErrorCode)
	}

	return ""
}

func logRawVapiErrorData(message string, vapiType *errors.ErrorTypeEnum, apiErrorDataValue *data.StructValue) error {
	dataValueToJSONEncoder := cleanjson.NewDataValueToJsonEncoder()
	errorStr, convErr := dataValueToJSONEncoder.Encode(apiErrorDataValue)
	if convErr != nil {
		log.Printf("[ERROR]: Failed to encode error details: %s", convErr)
		if vapiType != nil {
			return fmt.Errorf("%s (%s)", message, *vapiType)
		}
		return fmt.Errorf("%s (no additional details provided)", message)
	}

	log.Printf("[ERROR]: %s: %s", message, errorStr)
	return fmt.Errorf("%s: %s", message, errorStr)
}

func logVapiErrorData(message string, vapiMessages []std.LocalizableMessage, vapiType *errors.ErrorTypeEnum, apiErrorDataValue *data.StructValue) error {

	if apiErrorDataValue == nil {
		if len(vapiMessages) > 0 {
			return fmt.Errorf("%s (%s)", message, vapiMessages[0].DefaultMessage)
		}
		if vapiType != nil {
			return fmt.Errorf("%s (%s)", message, *vapiType)
		}

		return fmt.Errorf("%s (no additional details provided)", message)
	}

	// ApiError type is identical in all three relevant SDKs - policy, MP and GM
	var typeConverter = bindings.NewTypeConverter()
	data, err := typeConverter.ConvertToGolang(apiErrorDataValue, model.ApiErrorBindingType())

	// As of today, we cannot trust converter to return error in case target type doesn't
	// match the actual error. This issue is being looked into on VAPI level.
	// For now, we check both conversion error and actual contents of converted struct
	if err != nil {
		// This is likely not an error coming from NSX
		return logRawVapiErrorData(message, vapiType, apiErrorDataValue)
	}

	apiError, ok := data.(model.ApiError)
	if !ok {
		// This is likely not an error coming from NSX
		return logRawVapiErrorData(message, vapiType, apiErrorDataValue)
	}

	details := fmt.Sprintf(" %s: %s", message, printAPIError(apiError))

	if len(apiError.RelatedErrors) > 0 {
		details += "\nRelated errors:\n"
		for _, relatedErr := range apiError.RelatedErrors {
			details += fmt.Sprintf("%s ", printRelatedAPIError(relatedErr))
		}
	}
	log.Printf("[ERROR]: %s", details)
	return liberrors.New(details)
}

func logAPIError(message string, err error) error {
	if vapiError, ok := err.(errors.InvalidRequest); ok {
		// Connection errors end up here
		return logVapiErrorData(message, vapiError.Messages, vapiError.ErrorType, vapiError.Data)
	}
	if vapiError, ok := err.(errors.NotFound); ok {
		return logVapiErrorData(message, vapiError.Messages, vapiError.ErrorType, vapiError.Data)
	}
	if vapiError, ok := err.(errors.Unauthorized); ok {
		return logVapiErrorData(message, vapiError.Messages, vapiError.ErrorType, vapiError.Data)
	}
	if vapiError, ok := err.(errors.Unauthenticated); ok {
		return logVapiErrorData(message, vapiError.Messages, vapiError.ErrorType, vapiError.Data)
	}
	if vapiError, ok := err.(errors.InternalServerError); ok {
		return logVapiErrorData(message, vapiError.Messages, vapiError.ErrorType, vapiError.Data)
	}
	if vapiError, ok := err.(errors.ServiceUnavailable); ok {
		return logVapiErrorData(message, vapiError.Messages, vapiError.ErrorType, vapiError.Data)
	}
	if vapiError, ok := err.(errors.ConcurrentChange); ok {
		return logVapiErrorData(message, vapiError.Messages, vapiError.ErrorType, vapiError.Data)
	}

	return err
}

func isNotFoundError(err error) bool {
	if _, ok := err.(errors.NotFound); ok {
		return true
	}

	return false
}

func isServiceUnavailableError(err error) bool {
	if _, ok := err.(errors.ServiceUnavailable); ok {
		return true
	}
	return false
}

func isTimeoutError(err error) bool {
	if _, ok := err.(errors.TimedOut); ok {
		return true
	}
	return false
}

func handleCreateError(resourceType string, resource string, err error) error {
	msg := fmt.Sprintf("Failed to create %s %s", resourceType, resource)
	return logAPIError(msg, err)
}

func handleUpdateError(resourceType string, resourceID string, err error) error {
	msg := fmt.Sprintf("Failed to update %s %s", resourceType, resourceID)
	return logAPIError(msg, err)
}

func handleListError(resourceType string, err error) error {
	msg := fmt.Sprintf("Failed to read %ss", resourceType)
	return logAPIError(msg, err)
}

func handleReadError(d *schema.ResourceData, resourceType string, resource string, err error) error {
	msg := fmt.Sprintf("Failed to read %s %s", resourceType, resource)
	if isNotFoundError(err) {
		d.SetId("")
		log.Print(msg)
		return nil
	}
	return logAPIError(msg, err)
}

func handleDataSourceReadError(d *schema.ResourceData, resourceType string, resourceID string, err error) error {
	msg := fmt.Sprintf("Failed to read %s %s", resourceType, resourceID)
	return logAPIError(msg, err)
}

func handleDeleteError(resourceType string, resourceID string, err error) error {
	if isNotFoundError(err) {
		log.Printf("[WARNING] %s %s not found on backend", resourceType, resourceID)
		// We don't want to fail apply on this
		return nil
	}
	msg := fmt.Sprintf("Failed to delete %s %s", resourceType, resourceID)
	return logAPIError(msg, err)
}

func handleMultitenancyTier0Error() error {
	return fmt.Errorf("context use not supported with Tier0 gateways")
}
