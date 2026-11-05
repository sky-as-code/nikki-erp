package services

import (
	"fmt"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// Small plumbing shared by the promotion and voucher domain services.

// resetServerFields drops what only the server writes, so a client cannot set it (BR §8).
func resetServerFields(cmd dmodel.DynamicFields, fields []string) {
	for _, field := range fields {
		delete(cmd, field)
	}
}

func firstCreateOptions(options []dynamicresource.CreateOptions) dynamicresource.CreateOptions {
	if len(options) > 0 {
		return options[0]
	}
	return dynamicresource.CreateOptions{}
}

func firstUpdateOptions(options []dynamicresource.UpdateOptions) dynamicresource.UpdateOptions {
	if len(options) > 0 {
		return options[0]
	}
	return dynamicresource.UpdateOptions{}
}

func firstDeleteOptions(options []dynamicresource.DeleteOptions) dynamicresource.DeleteOptions {
	if len(options) > 0 {
		return options[0]
	}
	return dynamicresource.DeleteOptions{}
}

// clientErrorsOf turns a create result into the refusal a refusable transaction rolls back on.
func clientErrorsOf(result *dynamicresource.CreateResult, err error) (*ft.ClientErrors, error) {
	if err != nil {
		return nil, err
	}
	if result != nil && result.ClientErrors.Count() > 0 {
		return &result.ClientErrors, nil
	}
	return nil, nil
}

func mutateErrorsOf(result *dynamicresource.MutateResult, err error) (*ft.ClientErrors, error) {
	if err != nil {
		return nil, err
	}
	if result != nil && result.ClientErrors.Count() > 0 {
		return &result.ClientErrors, nil
	}
	return nil, nil
}

func createRefusal(vErrs *ft.ClientErrors, err error) (*dynamicresource.CreateResult, error) {
	if err != nil {
		return nil, err
	}
	return &dynamicresource.CreateResult{ClientErrors: *vErrs}, nil
}

func mutateRefusal(vErrs *ft.ClientErrors, err error) (*dynamicresource.MutateResult, error) {
	if err != nil {
		return nil, err
	}
	return &dynamicresource.MutateResult{ClientErrors: *vErrs}, nil
}

// changedFields names the given fields the command sets to a value other than the stored one.
func changedFields(stored, cmd dmodel.DynamicFields, fields ...string) []string {
	changed := make([]string, 0, len(fields))
	for _, field := range fields {
		value, sent := cmd[field]
		if !sent {
			continue
		}
		if textOf(value) != textOf(stored[field]) {
			changed = append(changed, field)
		}
	}
	return changed
}

// textOf reads an id or enum value whatever shape it arrived in, "" for none.
func textOf(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case *string:
		if typed != nil {
			return *typed
		}
		return ""
	}
	return fmt.Sprint(value)
}
