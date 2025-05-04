package convert

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// FromBytes converts byte array to a struct.
// Output parameter must be a pointer.
func FromBytes(input []byte, output any) error {
	var buf bytes.Buffer
	buf.Write(input)
	return FromByteBuffer(&buf, output)
}

// FromBytes converts byte array to a struct.
// Output parameter must be a pointer.
func FromByteBuffer(input *bytes.Buffer, output any) error {
	enc := gob.NewDecoder(input)
	err := enc.Decode(output)
	return err
}

// ToBytes converts an interface to bytes.
// Input must be a pointer.
func ToBytes(input any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(input)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// StringToDate attempts to parse a string into a time.Time type using a
// predefined list of formats.  If no suitable format is found, an error is
// returned.
func StringToDate(s string) (time.Time, error) {
	return parseDateWith(s, []string{
		time.RFC3339,
		"2006-01-02T15:04:05", // iso8601 without timezone
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC850,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		"2006-01-02 15:04:05.999999999 -0700 MST", // Time.String()
		"2006-01-02",
		"02 Jan 2006",
		"2006-01-02T15:04:05-0700", // RFC3339 without timezone hh:mm colon
		"2006-01-02 15:04:05 -07:00",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05Z07:00", // RFC3339 without T
		"2006-01-02 15:04:05Z0700",  // RFC3339 without T or timezone hh:mm colon
		"2006-01-02 15:04:05",
		time.Kitchen,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
	})
}

func parseDateWith(s string, dates []string) (d time.Time, e error) {
	for _, dateType := range dates {
		if d, e = time.Parse(dateType, s); e == nil {
			return
		}
	}
	return d, fmt.Errorf("unable to parse date: %s", s)
}

func GetStringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func GetStringValueByEnum(e any) string {
	if ptr, ok := e.(*string); ok {
		return GetStringValue(ptr)
	}
	return ""
}

func GetStringValueByTime(t *time.Time) string {
	if t != nil {
		return t.String()
	}
	return ""
}

func GetString(out interface{}) string {
	value := reflect.ValueOf(out)

	// if out is pointer
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return "nil"
		}
		// Get value of pointer
		value = value.Elem()
		if value.Kind() != reflect.Struct {
			return fmt.Sprintf("%v", value)
		}
	}
	// If object (struct) has Sprintf
	if value.Kind() == reflect.Struct {
		method := value.MethodByName("Sprintf")
		if method.IsValid() {
			results := method.Call(nil)
			if len(results) > 0 {
				return results[0].Interface().(string)
			}
		}
	}
	// if value is object
	if value.Kind() == reflect.Struct {
		if value.Type() == reflect.TypeOf(time.Time{}) {
			return fmt.Sprintf("%v", value)
		} else {
			var result string
			for i := 0; i < value.NumField(); i++ {
				field := value.Field(i)
				fieldName := value.Type().Field(i).Name

				if field.CanInterface() {
					fieldValue := field.Interface()

					valueFieldReflect := reflect.ValueOf(fieldValue)
					if valueFieldReflect.Kind() == reflect.Ptr && !valueFieldReflect.IsNil() {
						valueFieldReflect = valueFieldReflect.Elem()
					}

					result += fmt.Sprintf("%s: %v\n", fieldName, valueFieldReflect)
				} else {
					result += fmt.Sprintf("%s: <unexported>\n", fieldName)
				}
			}
			return result
		}

	}

	return fmt.Sprintf("%v", out)
}

func NewString(value string) *string {
	return &value
}

type GRN struct {
	Partition        string
	Service          string
	Region           string
	AccountId        string
	ResourceType     string
	ResourceId       string
	ResourceOriginal string
}

func ParseGRN(arn string) (*GRN, error) {
	parts := strings.SplitN(arn, ":", 6)
	if len(parts) != 6 {
		return nil, fmt.Errorf("invalid ARN format")
	}

	partition := parts[1]
	service := parts[2]
	region := parts[3]
	accountId := parts[4]
	resource := parts[5]

	var resourceType, resourceId string

	// some service do not have resourceType ????
	if service == "s3" {
		resourceType = ""
		resourceId = resource
	} else {
		if strings.Contains(resource, "/") {
			resourceParts := strings.SplitN(resource, "/", 2)
			resourceType = resourceParts[0]
			resourceId = resourceParts[1]
		} else if strings.Contains(resource, ":") {
			resourceParts := strings.SplitN(resource, ":", 2)
			resourceType = resourceParts[0]
			resourceId = resourceParts[1]
		} else {
			resourceType = ""
			resourceId = resource
		}
	}

	return &GRN{
		Partition:        partition,
		Service:          service,
		Region:           region,
		AccountId:        accountId,
		ResourceType:     resourceType,
		ResourceId:       resourceId,
		ResourceOriginal: resource,
	}, nil
}

func GenerateGRN(grn *GRN) (string, error) {
	if grn.Partition == "" || grn.Service == "" || grn.ResourceOriginal == "" {
		return "", fmt.Errorf("partition, service, and OriginalResourceFormat are required")
	}

	var arnBuilder strings.Builder
	arnBuilder.WriteString("arn:")
	arnBuilder.WriteString(grn.Partition)
	arnBuilder.WriteString(":")
	arnBuilder.WriteString(grn.Service)
	arnBuilder.WriteString(":")
	arnBuilder.WriteString(grn.Region)
	arnBuilder.WriteString(":")
	arnBuilder.WriteString(grn.AccountId)
	arnBuilder.WriteString(":")

	arnBuilder.WriteString(grn.ResourceOriginal)

	return arnBuilder.String(), nil
}

type IDENTITY_RESOURCE_TYPE string

const (
	IRT_USER   IDENTITY_RESOURCE_TYPE = "user"
	IRT_GROUP  IDENTITY_RESOURCE_TYPE = "group"
	IRT_ROLE   IDENTITY_RESOURCE_TYPE = "role"
	IRT_POLICY IDENTITY_RESOURCE_TYPE = "policy"
)

func GenerateIdentityGRN(accountId string, resourceType IDENTITY_RESOURCE_TYPE, resourceName string) (grn string, err error) {
	var resourceOriginal string
	if resourceName == "root" {
		resourceOriginal = resourceName
	} else {
		resourceOriginal = fmt.Sprintf("%s/%s", string(resourceType), resourceName)
	}
	return GenerateGRN(&GRN{
		Partition:        "gws",
		AccountId:        accountId,
		Service:          "iam",
		ResourceType:     string(resourceType),
		ResourceId:       resourceName,
		ResourceOriginal: resourceOriginal,
	})
}

func GeneratePolicyGRN(accountId string, resourceName string) (grn string, err error) {
	resourceType := IRT_POLICY
	resourceOriginal := fmt.Sprintf("%s/%s", string(resourceType), resourceName)
	return GenerateGRN(&GRN{
		Partition:        "gws",
		AccountId:        accountId,
		Service:          "iam",
		ResourceType:     string(resourceType),
		ResourceId:       resourceName,
		ResourceOriginal: resourceOriginal,
	})
}
