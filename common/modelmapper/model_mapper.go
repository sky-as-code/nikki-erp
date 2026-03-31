package modelmapper

import (
	stdErr "errors"
	"reflect"
	"strings"

	"go.bryk.io/pkg/errors"
)

const mapperTag = "mapper"
const tagOmit = "-"

var converterRegistry = make(map[reflect.Type]map[reflect.Type]Converter)
var mapType = reflect.TypeOf((*map[string]any)(nil)).Elem()

type fieldPair struct {
	meta  reflect.StructField
	value reflect.Value
}

type Converter func(in reflect.Value) (reflect.Value, error)

// AddConversion registers a custom converter used by MapToStruct,
// StructToMap and StructToStruct when no automatic rule applies. When
// both TSrcField and TDestField are non-pointer base types, all three
// pointer variants (*Src->Dest, Src->*Dest, *Src->*Dest) are
// automatically registered as well.
func AddConversion[TSrcField any, TDestField any](
	converter Converter,
) {
	srcType := reflect.TypeOf((*TSrcField)(nil)).Elem()
	destType := reflect.TypeOf((*TDestField)(nil)).Elem()
	storeConverter(srcType, destType, converter)
	if srcType.Kind() == reflect.Ptr ||
		destType.Kind() == reflect.Ptr {
		return
	}
	srcPtr := reflect.PointerTo(srcType)
	destPtr := reflect.PointerTo(destType)
	storeConverter(srcPtr, destType,
		synthesizePtrConv(srcPtr, destType, destType, converter))
	storeConverter(srcType, destPtr,
		synthesizePtrConv(srcType, destPtr, destType, converter))
	storeConverter(srcPtr, destPtr,
		synthesizePtrConv(srcPtr, destPtr, destType, converter))
}

// MapToStruct copies values from src map into dest struct fields
// matched by "json" tags. dest must be a non-nil pointer to a struct.
func MapToStruct(src map[string]any, dest any) error {
	if src == nil {
		return nil
	}
	structVal, err := requireStructPtr(dest, "dest", "MapToStruct")
	if err != nil {
		return err
	}
	return stdErr.Join(assignMapFields(src, structVal, true)...)
}

// StructToMap converts a struct into a map[string]any using "json"
// tags as keys. src must be a non-nil pointer to a struct.
func StructToMap(src any) (map[string]any, error) {
	srcVal, err := requireStructPtr(src, "src", "StructToMap")
	if err != nil {
		return nil, err
	}
	result := make(map[string]any)
	writeStructToMap(srcVal, result, true)
	return result, nil
}

// StructToStruct copies fields from src to dest matched by "json" tag
// keys. Both src and dest must be non-nil pointers to structs.
func StructToStruct(src any, dest any) error {
	srcVal, err := requireStructPtr(src, "src", "StructToStruct")
	if err != nil {
		return err
	}
	destVal, err := requireStructPtr(dest, "dest", "StructToStruct")
	if err != nil {
		return err
	}
	srcFields := gatherJsonFields(srcVal, true)
	return stdErr.Join(
		applyJsonFields(srcFields, destVal, true)...,
	)
}

// ---------- validation ----------

func requireStructPtr(
	v any, role, fn string,
) (reflect.Value, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return reflect.Value{}, errors.Errorf(
			"%s: %s must be a non-nil pointer to struct", fn, role,
		)
	}
	elem := rv.Elem()
	if elem.Kind() == reflect.Interface && !elem.IsNil() {
		elem = promoteIfaceToAddr(elem)
	}
	if elem.Kind() == reflect.Ptr && !elem.IsNil() {
		elem = elem.Elem()
	}
	if elem.Kind() != reflect.Struct {
		return reflect.Value{}, errors.Errorf(
			"%s: %s must point to a struct, got %v",
			fn, role, elem.Kind(),
		)
	}
	return elem, nil
}

// promoteIfaceToAddr unwraps an interface reflect.Value. When the
// concrete value is a struct (stored by value), it is copied into a
// freshly allocated pointer so the returned value is always
// addressable and its fields are settable via reflection.
func promoteIfaceToAddr(iface reflect.Value) reflect.Value {
	concrete := iface.Elem()
	if concrete.Kind() == reflect.Struct {
		ptr := reflect.New(concrete.Type())
		ptr.Elem().Set(concrete)
		iface.Set(ptr)
		return ptr
	}
	return concrete
}

// ---------- MapToStruct internals ----------

func assignMapFields(
	src map[string]any, destVal reflect.Value, expandEmbed bool,
) []error {
	destVal = ensureAddr(destVal)
	var errs []error
	t := destVal.Type()
	for i := 0; i < t.NumField(); i++ {
		field, fv := t.Field(i), destVal.Field(i)
		if !field.IsExported() {
			continue
		}
		if expandEmbed && isInlineEmbed(field) {
			errs = append(
				errs, assignEmbedMapFields(src, fv)...,
			)
			continue
		}
		if err := assignOneMapField(src, field, fv); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func assignEmbedMapFields(
	src map[string]any, fv reflect.Value,
) []error {
	embedded := initEmbedded(fv)
	if embedded.IsValid() && embedded.Kind() == reflect.Struct {
		return assignMapFields(src, embedded, false)
	}
	return nil
}

func assignOneMapField(
	src map[string]any, field reflect.StructField,
	fv reflect.Value,
) error {
	key, skip, _ := parseJsonTag(field)
	if skip {
		return nil
	}
	val, ok := src[key]
	if !ok || val == nil {
		return nil
	}
	if err := assignValue(reflect.ValueOf(val), fv); err != nil {
		return errors.Errorf("assignOneMapField: field %s: %w", field.Name, err)
	}
	return nil
}

// ---------- StructToMap internals ----------

func writeStructToMap(
	srcVal reflect.Value, out map[string]any, expandEmbed bool,
) {
	t := srcVal.Type()
	for i := 0; i < t.NumField(); i++ {
		field, fv := t.Field(i), srcVal.Field(i)
		if !field.IsExported() {
			continue
		}
		if expandEmbed && isInlineEmbed(field) {
			resolved := derefValue(fv)
			if resolved.IsValid() &&
				resolved.Kind() == reflect.Struct {
				writeStructToMap(resolved, out, false)
			}
			continue
		}
		writeOneFieldToMap(field, fv, out)
	}
}

func writeOneFieldToMap(
	field reflect.StructField, fv reflect.Value,
	out map[string]any,
) {
	key, skip, omitempty := parseJsonTag(field)
	if skip || (omitempty && fv.IsZero()) {
		return
	}
	out[key] = fv.Interface()
}

// ---------- StructToStruct internals ----------

func gatherJsonFields(
	val reflect.Value, expandEmbed bool,
) map[string]fieldPair {
	result := make(map[string]fieldPair)
	t := val.Type()
	for i := 0; i < t.NumField(); i++ {
		field, fv := t.Field(i), val.Field(i)
		if !field.IsExported() {
			continue
		}
		if expandEmbed && isInlineEmbed(field) {
			gatherEmbeddedFields(fv, result)
			continue
		}
		key, skip, _ := parseJsonTag(field)
		if skip {
			continue
		}
		result[key] = fieldPair{meta: field, value: fv}
	}
	return result
}

func gatherEmbeddedFields(
	fv reflect.Value, result map[string]fieldPair,
) {
	resolved := derefValue(fv)
	if !resolved.IsValid() || resolved.Kind() != reflect.Struct {
		return
	}
	for k, v := range gatherJsonFields(resolved, false) {
		if _, exists := result[k]; !exists {
			result[k] = v
		}
	}
}

func applyJsonFields(
	srcFields map[string]fieldPair, destVal reflect.Value,
	expandEmbed bool,
) []error {
	destVal = ensureAddr(destVal)
	var errs []error
	t := destVal.Type()
	for i := 0; i < t.NumField(); i++ {
		field, fv := t.Field(i), destVal.Field(i)
		if !field.IsExported() {
			continue
		}
		if expandEmbed && isInlineEmbed(field) {
			errs = append(
				errs, applyEmbedFields(srcFields, fv)...,
			)
			continue
		}
		if err := applyOneField(srcFields, field, fv); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func applyEmbedFields(
	srcFields map[string]fieldPair, fv reflect.Value,
) []error {
	embedded := initEmbedded(fv)
	if embedded.IsValid() && embedded.Kind() == reflect.Struct {
		return applyJsonFields(srcFields, embedded, false)
	}
	return nil
}

func applyOneField(
	srcFields map[string]fieldPair,
	field reflect.StructField, fv reflect.Value,
) error {
	key, skip, _ := parseJsonTag(field)
	if skip {
		return nil
	}
	pair, ok := srcFields[key]
	if !ok {
		return nil
	}
	if err := assignValue(pair.value, fv); err != nil {
		return errors.Errorf("applyOneField: field %s: %w", field.Name, err)
	}
	return nil
}

// ---------- value assignment ----------

func assignValue(srcVal, destField reflect.Value) error {
	srcVal = unwrapIface(srcVal)
	if !srcVal.IsValid() {
		return nil
	}
	srcType, destType := srcVal.Type(), destField.Type()
	if srcType == destType {
		destField.Set(srcVal)
		return nil
	}
	if srcType.Kind() == reflect.Ptr && srcVal.IsNil() {
		return nil
	}
	if err, ok := tryPtrShortcut(srcVal, destField); ok {
		return err
	}
	realSrc := derefValue(srcVal)
	if !realSrc.IsValid() {
		return nil
	}
	realDestType, destIsPtr := stripPtrType(destType)
	return convertAndSet(
		srcVal, realSrc, destField, realDestType, destIsPtr,
	)
}

func tryPtrShortcut(
	srcVal, destField reflect.Value,
) (error, bool) {
	srcType, destType := srcVal.Type(), destField.Type()
	if srcType.Kind() == reflect.Ptr &&
		srcType.Elem() == destType {
		destField.Set(srcVal.Elem())
		return nil, true
	}
	if destType.Kind() == reflect.Ptr &&
		destType.Elem() == srcType {
		ptr := reflect.New(srcType)
		ptr.Elem().Set(srcVal)
		destField.Set(ptr)
		return nil, true
	}
	return nil, false
}

func convertAndSet(
	srcVal, realSrc, destField reflect.Value,
	realDestType reflect.Type, destIsPtr bool,
) error {
	realSrcType := realSrc.Type()
	if realSrcType.AssignableTo(realDestType) {
		return setOrWrap(
			realSrc, destField, destIsPtr, realDestType,
		)
	}
	if realSrcType.ConvertibleTo(realDestType) {
		return setOrWrap(
			realSrc.Convert(realDestType),
			destField, destIsPtr, realDestType,
		)
	}
	if conv, ok := lookupConverter(
		srcVal.Type(), destField.Type(),
	); ok {
		return applyConverter(conv, srcVal, destField)
	}
	return assignNested(realSrc, destField, realDestType, destIsPtr)
}

func assignNested(
	realSrc, destField reflect.Value,
	realDestType reflect.Type, destIsPtr bool,
) error {
	srcKind := realSrc.Type().Kind()
	if srcKind == reflect.Map &&
		realDestType.Kind() == reflect.Struct &&
		isMapBased(realSrc.Type()) {
		m := realSrc.Convert(mapType).Interface().(map[string]any)
		return nestedMapToStruct(
			m, destField, realDestType, destIsPtr,
		)
	}
	if srcKind == reflect.Struct &&
		realDestType.Kind() == reflect.Struct {
		return nestedStructToStruct(
			realSrc, destField, realDestType, destIsPtr,
		)
	}
	if srcKind == reflect.Slice &&
		realDestType.Kind() == reflect.Slice {
		return assignSlice(
			realSrc, destField, realDestType, destIsPtr,
		)
	}
	return errors.Errorf(
		"cannot convert %v to %v", realSrc.Type(), destField.Type(),
	)
}

func nestedMapToStruct(
	m map[string]any, destField reflect.Value,
	realDestType reflect.Type, destIsPtr bool,
) error {
	if destIsPtr {
		ptr := reflect.New(realDestType)
		errs := assignMapFields(m, ptr.Elem(), true)
		destField.Set(ptr)
		return stdErr.Join(errs...)
	}
	return stdErr.Join(assignMapFields(m, destField, true)...)
}

func nestedStructToStruct(
	src, destField reflect.Value,
	realDestType reflect.Type, destIsPtr bool,
) error {
	srcFields := gatherJsonFields(src, true)
	if destIsPtr {
		ptr := reflect.New(realDestType)
		errs := applyJsonFields(srcFields, ptr.Elem(), true)
		destField.Set(ptr)
		return stdErr.Join(errs...)
	}
	return stdErr.Join(
		applyJsonFields(srcFields, destField, true)...,
	)
}

func assignSlice(
	realSrc, destField reflect.Value,
	realDestType reflect.Type, destIsPtr bool,
) error {
	destElemType := realDestType.Elem()
	srcLen := realSrc.Len()
	result := reflect.MakeSlice(realDestType, srcLen, srcLen)
	for i := 0; i < srcLen; i++ {
		srcItem := unwrapIface(realSrc.Index(i))
		if !srcItem.IsValid() {
			continue
		}
		if err := convertSliceItem(
			srcItem, result.Index(i), destElemType, i,
		); err != nil {
			return err
		}
	}
	return setOrWrap(
		result, destField, destIsPtr, realDestType,
	)
}

// convertSliceItem applies the 4-step conversion for a single
// slice element: (1) direct assign, (2) reflect convert,
// (3) custom converter, (4) error. No recursive descent into
// struct, slice, or map element values.
func convertSliceItem(
	srcItem, destItem reflect.Value,
	destElemType reflect.Type, idx int,
) error {
	srcType := srcItem.Type()
	if srcType.AssignableTo(destElemType) {
		destItem.Set(srcItem)
		return nil
	}
	if srcType.ConvertibleTo(destElemType) {
		destItem.Set(srcItem.Convert(destElemType))
		return nil
	}
	if conv, ok := lookupConverter(
		srcType, destElemType,
	); ok {
		res, err := conv(srcItem)
		if err != nil {
			return errors.Errorf(
				"convertSliceItem: element [%d]: %w", idx, err,
			)
		}
		destItem.Set(res)
		return nil
	}
	return errors.Errorf(
		"convertSliceItem: element [%d]: cannot convert %v to %v",
		idx, srcType, destElemType,
	)
}

func setOrWrap(
	val, destField reflect.Value,
	wrapPtr bool, innerType reflect.Type,
) error {
	if wrapPtr {
		ptr := reflect.New(innerType)
		ptr.Elem().Set(val)
		destField.Set(ptr)
		return nil
	}
	destField.Set(val)
	return nil
}

// ---------- json tag handling ----------

func parseJsonTag(
	field reflect.StructField,
) (key string, skip bool, omitempty bool) {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", true, false
	}
	if idx := strings.Index(tag, ","); idx != -1 {
		omitempty = strings.Contains(tag[idx+1:], "omitempty")
		tag = tag[:idx]
	}
	if tag == "" {
		return field.Name, false, omitempty
	}
	return tag, false, omitempty
}

func isInlineEmbed(field reflect.StructField) bool {
	return field.Anonymous && field.Tag.Get("json") == ""
}

// ---------- embedded struct handling ----------

func initEmbedded(fv reflect.Value) reflect.Value {
	if fv.Kind() == reflect.Ptr {
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		return fv.Elem()
	}
	return fv
}

// ---------- reflect utilities ----------

// ensureAddr guarantees that a struct Value is addressable (and
// therefore its exported fields are settable). When v is already
// addressable it is returned as-is. Otherwise a heap-allocated copy
// is created via reflect.New and its Elem is returned.
func ensureAddr(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Struct || v.CanAddr() {
		return v
	}
	ptr := reflect.New(v.Type())
	ptr.Elem().Set(v)
	return ptr.Elem()
}

func unwrapIface(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return reflect.Value{}
		}
		return v.Elem()
	}
	return v
}

func derefValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

func stripPtrType(t reflect.Type) (reflect.Type, bool) {
	if t.Kind() == reflect.Ptr {
		return t.Elem(), true
	}
	return t, false
}

func isMapBased(t reflect.Type) bool {
	return t.Kind() == reflect.Map && t.ConvertibleTo(mapType)
}

// ---------- converter registry ----------

func storeConverter(
	srcType, destType reflect.Type, conv Converter,
) {
	if _, ok := converterRegistry[srcType]; !ok {
		converterRegistry[srcType] = make(map[reflect.Type]Converter)
	}
	converterRegistry[srcType][destType] = conv
}

func synthesizePtrConv(
	srcType, destType, baseDest reflect.Type, baseConv Converter,
) Converter {
	srcIsPtr := srcType.Kind() == reflect.Ptr
	destIsPtr := destType.Kind() == reflect.Ptr
	return func(in reflect.Value) (reflect.Value, error) {
		src := derefConvSrc(in, srcIsPtr)
		if !src.IsValid() {
			return reflect.Zero(destType), nil
		}
		result, err := baseConv(src)
		if err != nil {
			return reflect.Value{}, err
		}
		if destIsPtr {
			ptr := reflect.New(baseDest)
			ptr.Elem().Set(result)
			return ptr, nil
		}
		return result, nil
	}
}

func derefConvSrc(in reflect.Value, isPtr bool) reflect.Value {
	if !isPtr {
		return in
	}
	if in.IsNil() {
		return reflect.Value{}
	}
	return in.Elem()
}

func applyConverter(
	conv func(reflect.Value) (reflect.Value, error),
	srcField, destField reflect.Value,
) error {
	result, err := conv(srcField)
	if err != nil {
		return err
	}
	destField.Set(result)
	return nil
}

func lookupConverter(
	srcType, destType reflect.Type,
) (Converter, bool) {
	if m, ok := converterRegistry[srcType]; ok {
		if conv, ok := m[destType]; ok {
			return conv, true
		}
	}
	return nil, false
}
