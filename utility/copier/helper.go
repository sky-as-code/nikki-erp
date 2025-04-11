package copier

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/labstack/gommon/log"
	"github.com/sky-as-code/nikki-erp/utility/fault"
	"gopkg.in/jeevatkm/go-model.v1"
)

type ICopier interface {
	Copy(toValue, fromValue interface{}) (err error)
	JSONCopy(dst, src interface{}) (err error)
	CopyWithJSONTag(dst, src interface{}) (err error)
	CopyExWithJSONTag(dst, src interface{}) (err error)
}

type cCopier struct{}

func (c *cCopier) JSONCopy(dst, src interface{}) (err error) {
	return JSONCopy(dst, src)
}

func (c *cCopier) Copy(toValue, fromValue interface{}) (err error) {
	return Copy(toValue, fromValue)
}

func (c *cCopier) CopyWithJSONTag(toValue, fromValue interface{}) (err error) {
	return CopyWithJSONTag(toValue, fromValue)
}

func (c *cCopier) CopyExWithJSONTag(toValue, fromValue interface{}) (err error) {
	return CopyExWithJSONTag(toValue, fromValue)
}

var Instance ICopier = &cCopier{}

// Copy things
func Copy(toValue, fromValue interface{}) (err error) {
	errors := model.Copy(toValue, fromValue)
	err = fault.JoinErrors(errors)
	return err
}

// JSONCopy ...
func JSONCopy(dst, src interface{}) error {
	bytes, err := json.Marshal(src)
	if err != nil {
		log.Info("JSONCopy: err marshal json = ", err)
		return err
	}

	err = json.Unmarshal(bytes, dst)
	if err != nil {
		log.Info("JSONCopy: err unmarshal json = ", err)
		return err
	}

	return nil
}

func CopyWithJSONTag(dest interface{}, source interface{}) error {
	sValue := reflect.ValueOf(source).Elem()
	dValue := reflect.ValueOf(dest).Elem()

	for i := 0; i < sValue.NumField(); i++ {
		sField := sValue.Type().Field(i)
		sFieldName := sField.Tag.Get("json")
		if sFieldName == "" {
			continue
		}
		sFieldName = strings.TrimSuffix(sFieldName, ",omitempty")

		dField, found := dValue.Type().FieldByNameFunc(func(name string) bool {
			structField, ok := dValue.Type().FieldByName(name)
			if !ok {
				return false
			}
			nameJson := strings.TrimSuffix(structField.Tag.Get("json"), ",omitempty")
			return nameJson == sFieldName
		})
		if !found {
			log.Infof("CopyWithJSONTag Error sField: %v, sFieldName: %v, dField: %v", sField, sFieldName, dField)
			continue
		}

		dFieldName := dField.Tag.Get("json")
		if dFieldName == "" {
			continue
		}
		dFieldName = strings.TrimSuffix(dFieldName, ",omitempty")
		if dFieldName != sFieldName {
			continue
		}

		//log.Infof("CopyWithJSONTag sField: %v, sFieldName: %v, dField: %v, dFieldName: %v", sField, sFieldName, dField, dFieldName)

		if sField.Type != dField.Type {
			continue
		}

		dValue.FieldByName(dField.Name).Set(sValue.Field(i))
	}

	return nil
}

func CopyExWithJSONTag(dest interface{}, source interface{}) error {
	sValue := reflect.ValueOf(source).Elem()
	dValue := reflect.ValueOf(dest).Elem()

	for i := 0; i < sValue.NumField(); i++ {
		sField := sValue.Type().Field(i)
		sFieldName := sField.Tag.Get("json")
		if sFieldName == "" {
			continue
		}
		sFieldName = strings.TrimSuffix(sFieldName, ",omitempty")

		dField, found := dValue.Type().FieldByNameFunc(func(name string) bool {
			structField, ok := dValue.Type().FieldByName(name)
			if !ok {
				return false
			}
			nameJson := strings.TrimSuffix(structField.Tag.Get("json"), ",omitempty")
			return nameJson == sFieldName
		})
		if !found {
			log.Infof("CopyExWithJSONTag Error sField: %v, sFieldName: %v, dField: %v", sField, sFieldName, dField)
			continue
		}

		dFieldName := dField.Tag.Get("json")
		if dFieldName == "" {
			continue
		}
		dFieldName = strings.TrimSuffix(dFieldName, ",omitempty")
		if dFieldName != sFieldName {
			continue
		}

		//log.Infof("CopyExWithJSONTag sField: %v, sFieldName: %v, dField: %v, dFieldName: %v", sField, sFieldName, dField, dFieldName)

		// Nếu trường là con trỏ và trường tương ứng trong struct dest là giá trị
		if sField.Type == reflect.PtrTo(dField.Type) && sValue.Field(i).IsNil() == false {
			dValue.FieldByName(dField.Name).Set(sValue.Field(i).Elem())
		} else if sField.Type == dField.Type {
			dValue.FieldByName(dField.Name).Set(sValue.Field(i))
		} else {
			continue
		}

	}

	return nil
}
