package copier

import (
	"encoding/json"
	stdErr "errors"
	"reflect"
	"strings"

	"github.com/labstack/gommon/log"
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

func (c *cCopier) Copy(fromValue, toValue interface{}) (err error) {
	return Copy(fromValue, toValue)
}

func (c *cCopier) CopyWithJSONTag(toValue, fromValue interface{}) (err error) {
	return CopyWithJSONTag(toValue, fromValue)
}

func (c *cCopier) CopyExWithJSONTag(toValue, fromValue interface{}) (err error) {
	return CopyExWithJSONTag(toValue, fromValue)
}

var Instance ICopier = &cCopier{}

// Copy objects
func Copy[TSrc any, TDest any](fromValue TSrc, toValue TDest) (err error) {
	errors := model.Copy(toValue, fromValue)
	err = stdErr.Join(errors...)
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

func init() {
	// mapType := reflect.TypeOf((map[string]string)(nil)).Elem()
	mapType := reflect.ValueOf((map[string]string)(nil)).Type()
	mapPtr := make(map[string]string)
	// mapPtrType := reflect.TypeOf((*map[string]string)(nil)).Elem()
	mapPtrType := reflect.ValueOf(&mapPtr).Type()
	model.AddConversionByType(mapType, mapPtrType, func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf((*map[string]string)(nil)), nil
		}

		result := in.Interface().(map[string]string)
		return reflect.ValueOf(&result), nil
	})

	model.AddConversionByType(mapPtrType, mapType, func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf((map[string]string)(nil)), nil
		}

		result := *in.Interface().(*map[string]string)
		return reflect.ValueOf(result), nil
	})

	stringType := reflect.TypeOf("")
	stringPtrType := reflect.TypeOf((*string)(nil))
	model.AddConversionByType(stringType, stringPtrType, func(in reflect.Value) (reflect.Value, error) {
		result := in.Interface().(string)
		return reflect.ValueOf(&result), nil
	})
	model.AddConversionByType(stringPtrType, stringType, func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf(""), nil
		}

		result := *in.Interface().(*string)
		return reflect.ValueOf(result), nil
	})
}
