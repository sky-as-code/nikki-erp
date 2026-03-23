package modelmapper

import (
	stdErr "errors"
	"reflect"
	"strings"

	"go.bryk.io/pkg/errors"
)

const mapperTag = "mapper"
const tagOmit = "-"

// converterRegistry stores custom field converters indexed by (srcType → destType).
var converterRegistry = make(map[reflect.Type]map[reflect.Type]Converter)

// mapType is the canonical reflect.Type for map[string]any.
var mapType = reflect.TypeOf((*map[string]any)(nil)).Elem()

type fieldPair struct {
	meta  reflect.StructField
	value reflect.Value
}

type Converter func(in reflect.Value) (reflect.Value, error)

// AddConversion registers a custom converter used by Copy when no automatic rule applies.
// When both TSrcField and TDestField are non-pointer base types, all three pointer
// variants (*Src→Dest, Src→*Dest, *Src→*Dest) are automatically registered as well.
func AddConversion[TSrcField any, TDestField any](converter Converter) {
	srcType := reflect.TypeOf((*TSrcField)(nil)).Elem()
	destType := reflect.TypeOf((*TDestField)(nil)).Elem()
	storeConverter(srcType, destType, converter)
	if srcType.Kind() == reflect.Ptr || destType.Kind() == reflect.Ptr {
		return
	}
	srcPtr, destPtr := reflect.PointerTo(srcType), reflect.PointerTo(destType)
	storeConverter(srcPtr, destType, synthesizePointerConverter(srcPtr, destType, destType, converter))
	storeConverter(srcType, destPtr, synthesizePointerConverter(srcType, destPtr, destType, converter))
	storeConverter(srcPtr, destPtr, synthesizePointerConverter(srcPtr, destPtr, destType, converter))
}

func storeConverter(srcType, destType reflect.Type, conv Converter) {
	if _, ok := converterRegistry[srcType]; !ok {
		converterRegistry[srcType] = make(map[reflect.Type]Converter)
	}
	converterRegistry[srcType][destType] = conv
}

// Copy copies all exported fields from src into a new TDest value, returning
// joined errors for every field that could not be mapped. TDest must be a pointer
// type or a map-based type (convertible to map[string]any). src may be a pointer,
// a plain struct, or any map-based type. A nil pointer src returns a zero TDest.
// Fields tagged `mapper:"-"` are skipped.
func Copy[TDest any](src any) (TDest, error) {
	var zero TDest
	destType := reflect.TypeOf((*TDest)(nil)).Elem()
	srcVal := reflect.ValueOf(src)
	if srcVal.IsValid() && isMapBased(srcVal.Type()) {
		if isMapBased(destType) {
			return shallowCopyMapAs[TDest](srcVal, destType)
		}
		if destType.Kind() == reflect.Ptr && isMapBased(destType.Elem()) {
			return shallowCopyMapAsPtr[TDest](srcVal, destType.Elem())
		}
		return MapToStruct[TDest](asMap(srcVal))
	}
	ptrDestType, err := requirePtrDest[TDest]()
	if err != nil {
		return zero, err
	}
	srcResolved, valid, err := resolveSrcToStruct(src)
	if err != nil {
		return zero, err
	}
	if !valid {
		return zero, nil
	}
	dest := reflect.New(ptrDestType.Elem())
	errs := mapFields(srcResolved, dest.Elem())
	destResult, _ := dest.Interface().(TDest)
	return destResult, stdErr.Join(errs...)
}

// CastCopy performs a fast cast when TSrc and TDest share the same element type,
// or falls back to field-by-field copy otherwise. TDest may be a pointer type or
// a map-based type (convertible to map[string]any). For map-based src and dest,
// a reflect.Convert cast is attempted first (zero-copy reinterpretation), falling
// back to a shallow entry copy when types are not directly convertible.
// A nil pointer src returns a zero TDest.
func CastCopy[TDest any](src any) (TDest, error) {
	var zero TDest
	destType := reflect.TypeOf((*TDest)(nil)).Elem()
	srcVal := reflect.ValueOf(src)
	if srcVal.IsValid() && isMapBased(srcVal.Type()) {
		if isMapBased(destType) {
			return castMapAs[TDest](srcVal, destType)
		}
		if destType.Kind() == reflect.Ptr && isMapBased(destType.Elem()) {
			return castMapAsPtr[TDest](srcVal, destType.Elem())
		}
		return MapToStruct[TDest](asMap(srcVal))
	}
	ptrDestType, err := requirePtrDest[TDest]()
	if err != nil {
		return zero, err
	}
	if srcVal.Kind() == reflect.Ptr {
		return castCopyFromPtr[TDest](srcVal, ptrDestType)
	}
	if srcVal.Kind() == reflect.Struct {
		return castCopyViaReflect[TDest](srcVal, srcVal.Type(), ptrDestType.Elem(), src)
	}
	return zero, errors.Errorf("src must be a pointer, struct, or map-based type, got %T", src)
}

// castCopyFromPtr handles the pointer-specific fast paths: nil guard and same-type
// unsafe cast. Falls through to castCopyViaReflect for all other cases.
func castCopyFromPtr[TDest any](srcVal reflect.Value, destType reflect.Type) (TDest, error) {
	var zero TDest
	if srcVal.IsNil() {
		return zero, nil
	}
	destElemType := destType.Elem()
	if srcVal.Type().Elem() == destElemType {
		result := reflect.NewAt(destElemType, srcVal.UnsafePointer())
		return result.Interface().(TDest), nil
	}
	return castCopyViaReflect[TDest](srcVal.Elem(), srcVal.Type().Elem(), destElemType, srcVal.Interface())
}

// castCopyViaReflect tries a reflect-level cast; falls back to Copy on failure.
func castCopyViaReflect[TDest any](srcElem reflect.Value, srcType, destElemType reflect.Type, orig any) (TDest, error) {
	if result, ok := tryCastReflect(srcType, destElemType, srcElem); ok {
		ptr := reflect.New(destElemType)
		ptr.Elem().Set(result)
		return ptr.Interface().(TDest), nil
	}
	return Copy[TDest](orig)
}

// MapToStruct converts a string-keyed map into a new TDest value using "json" field
// tags as key names, following standard json conventions: "-" skips the field,
// "name,omitempty" uses "name", and no tag falls back to the struct field name.
// Embedded structs without a json tag are expanded. TDest must be a pointer type.
// A nil map returns a nil TDest. Values are assigned by direct cast first, then
// reflect conversion, then a registered converter.
func MapToStruct[TDest any](src map[string]any) (TDest, error) {
	var zero TDest
	if src == nil {
		return zero, nil
	}
	destType := reflect.TypeOf((*TDest)(nil)).Elem()
	if destType.Kind() != reflect.Ptr {
		return zero, errors.Errorf("MapToStruct: TDest must be a pointer type, got %v", destType)
	}
	dest := reflect.New(destType.Elem())
	errs := assignMapToStruct(src, dest.Elem())
	destResult, _ := dest.Interface().(TDest)
	return destResult, stdErr.Join(errs...)
}

// isMapBased reports whether t is a map kind that is convertible to map[string]any.
func isMapBased(t reflect.Type) bool {
	return t.Kind() == reflect.Map && t.ConvertibleTo(mapType)
}

// asMap reinterprets a map-based reflect.Value as a plain map[string]any.
func asMap(v reflect.Value) map[string]any {
	return v.Convert(mapType).Interface().(map[string]any)
}

// castMapAs performs a fast reflect.Convert cast from a map-based src to TDest.
// Falls back to shallowCopyMapAs when the types are not directly convertible.
func castMapAs[TDest any](srcVal reflect.Value, destType reflect.Type) (TDest, error) {
	var zero TDest
	if srcVal.Type().ConvertibleTo(destType) {
		result, ok := srcVal.Convert(destType).Interface().(TDest)
		if !ok {
			return zero, errors.Errorf("cannot cast %v to %v", srcVal.Type(), destType)
		}
		return result, nil
	}
	return shallowCopyMapAs[TDest](srcVal, destType)
}

// shallowCopyMapAs allocates a new map of destType and copies all entries from
// srcVal into it, then returns the result as TDest.
func shallowCopyMapAs[TDest any](srcVal reflect.Value, destType reflect.Type) (TDest, error) {
	var zero TDest
	src := resolvePtr(srcVal)
	if !src.IsValid() || src.IsNil() {
		return zero, nil
	}
	newMap := reflect.MakeMap(destType)
	if err := fillMapEntries(src, newMap, destType); err != nil {
		return zero, err
	}
	result, ok := newMap.Interface().(TDest)
	if !ok {
		return zero, errors.Errorf("cannot convert map %v to %v", src.Type(), destType)
	}
	return result, nil
}

// shallowCopyMapAsPtr allocates a new map of destElemType, copies all entries from
// srcVal into it, wraps the result in a pointer, and returns it as TDest.
func shallowCopyMapAsPtr[TDest any](srcVal reflect.Value, destElemType reflect.Type) (TDest, error) {
	var zero TDest
	src := resolvePtr(srcVal)
	if !src.IsValid() || src.IsNil() {
		return zero, nil
	}
	newMap := reflect.MakeMap(destElemType)
	if err := fillMapEntries(src, newMap, destElemType); err != nil {
		return zero, err
	}
	ptr := reflect.New(destElemType)
	ptr.Elem().Set(newMap)
	result, ok := ptr.Interface().(TDest)
	if !ok {
		return zero, errors.Errorf("cannot convert *map %v to %v", destElemType, reflect.TypeOf((*TDest)(nil)).Elem())
	}
	return result, nil
}

// castMapAsPtr performs a fast reflect.Convert cast from a map-based src to *destElemType
// (TDest). Falls back to a shallow entry copy when types are not directly convertible.
func castMapAsPtr[TDest any](srcVal reflect.Value, destElemType reflect.Type) (TDest, error) {
	var zero TDest
	var mapVal reflect.Value
	if srcVal.Type().ConvertibleTo(destElemType) {
		mapVal = srcVal.Convert(destElemType)
	} else {
		src := resolvePtr(srcVal)
		if !src.IsValid() || src.IsNil() {
			return zero, nil
		}
		newMap := reflect.MakeMap(destElemType)
		if err := fillMapEntries(src, newMap, destElemType); err != nil {
			return zero, err
		}
		mapVal = newMap
	}
	ptr := reflect.New(destElemType)
	ptr.Elem().Set(mapVal)
	result, ok := ptr.Interface().(TDest)
	if !ok {
		return zero, errors.Errorf("cannot cast *map %v to %v", destElemType, reflect.TypeOf((*TDest)(nil)).Elem())
	}
	return result, nil
}

func assignMapToStruct(src map[string]any, destVal reflect.Value) []error {
	if destVal.Kind() != reflect.Struct {
		return []error{errors.Errorf("dest must be a struct, got %v", destVal.Kind())}
	}
	var errs []error
	t := destVal.Type()
	for i := 0; i < t.NumField(); i++ {
		field, fieldVal := t.Field(i), destVal.Field(i)
		if !field.IsExported() || !fieldVal.CanSet() {
			continue
		}
		if field.Anonymous && field.Tag.Get("json") == "" {
			if embedded := resolvePtr(fieldVal); embedded.IsValid() && embedded.Kind() == reflect.Struct {
				errs = append(errs, assignMapToStruct(src, embedded)...)
				continue
			}
		}
		key, skip := jsonKey(field)
		if skip {
			continue
		}
		val, exists := src[key]
		if !exists || val == nil {
			continue
		}
		if err := assignFromAny(val, fieldVal); err != nil {
			errs = append(errs, errors.Errorf("field %s: %w", field.Name, err))
		}
	}
	return errs
}

func jsonKey(field reflect.StructField) (key string, skip bool) {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", true
	}
	if idx := strings.Index(tag, ","); idx != -1 {
		tag = tag[:idx]
	}
	if tag == "" {
		return field.Name, false
	}
	return tag, false
}

func assignFromAny(val any, destField reflect.Value) error {
	srcVal := reflect.ValueOf(val)
	if err, ok := tryDirectAssign(srcVal, destField); ok {
		return err
	}
	if srcVal.Type().ConvertibleTo(destField.Type()) {
		destField.Set(srcVal.Convert(destField.Type()))
		return nil
	}
	if conv, ok := lookupConverter(srcVal.Type(), destField.Type()); ok {
		return applyConverter(conv, srcVal, destField)
	}
	return errors.Errorf("cannot assign %T to %v", val, destField.Type())
}

func requirePtrArgs[TDest any](src any) (reflect.Value, reflect.Type, error) {
	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() != reflect.Ptr {
		return reflect.Value{}, nil, errors.Errorf("src must be a pointer, got %T", src)
	}
	destType := reflect.TypeOf((*TDest)(nil)).Elem()
	if destType.Kind() != reflect.Ptr {
		return reflect.Value{}, nil, errors.Errorf("TDest must be a pointer type, got %v", destType)
	}
	return srcVal, destType, nil
}

func requirePtrDest[TDest any]() (reflect.Type, error) {
	destType := reflect.TypeOf((*TDest)(nil)).Elem()
	if destType.Kind() != reflect.Ptr {
		return nil, errors.Errorf("TDest must be a pointer type, got %v", destType)
	}
	return destType, nil
}

// resolveSrcToStruct resolves src to its underlying struct Value, accepting both
// pointer and non-pointer structs. Returns (zero, false, nil) for nil pointers.
func resolveSrcToStruct(src any) (reflect.Value, bool, error) {
	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() == reflect.Ptr {
		if srcVal.IsNil() {
			return reflect.Value{}, false, nil
		}
		resolved := resolvePtr(srcVal.Elem())
		return resolved, resolved.IsValid(), nil
	}
	if srcVal.Kind() == reflect.Struct {
		return srcVal, true, nil
	}
	return reflect.Value{}, false, errors.Errorf("src must be a pointer, struct, or map[string]any, got %T", src)
}

// tryCastReflect attempts a zero-copy (or minimal-copy) cast via reflect
// for pointer-indirection and type-convertible cases.
func tryCastReflect(srcType, destType reflect.Type, srcVal reflect.Value) (reflect.Value, bool) {
	if srcType.Kind() == reflect.Ptr && srcType.Elem() == destType {
		if srcVal.IsNil() {
			return reflect.Zero(destType), true
		}
		return srcVal.Elem(), true
	}
	if destType.Kind() == reflect.Ptr && destType.Elem() == srcType {
		newPtr := reflect.New(srcType)
		newPtr.Elem().Set(srcVal)
		return newPtr, true
	}
	if srcType.ConvertibleTo(destType) {
		return srcVal.Convert(destType), true
	}
	return reflect.Value{}, false
}

// mapFields iterates exported fields of srcVal and assigns them to matching
// fields in destVal, returning one error per unmappable field.
func mapFields(srcVal, destVal reflect.Value) []error {
	if srcVal.Kind() != reflect.Struct || destVal.Kind() != reflect.Struct {
		return []error{errors.Errorf("both src and dest must be structs, got %v → %v", srcVal.Kind(), destVal.Kind())}
	}
	var errs []error
	for _, pair := range collectFields(srcVal) {
		if pair.meta.Tag.Get(mapperTag) == tagOmit {
			continue
		}
		destField := destVal.FieldByName(pair.meta.Name)
		if !destField.IsValid() || !destField.CanSet() {
			continue
		}
		if err := assignField(pair.value, destField); err != nil {
			errs = append(errs, errors.Errorf("field %s: %w", pair.meta.Name, err))
		}
	}
	return errs
}

// collectFields returns a flat list of all exported fields, expanding
// anonymous (embedded) struct fields recursively.
func collectFields(val reflect.Value) []fieldPair {
	var pairs []fieldPair
	t := val.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := val.Field(i)
		if !f.IsExported() {
			continue
		}
		if f.Anonymous {
			resolved := resolvePtr(fv)
			if resolved.IsValid() && resolved.Kind() == reflect.Struct {
				pairs = append(pairs, collectFields(resolved)...)
				continue
			}
		}
		pairs = append(pairs, fieldPair{meta: f, value: fv})
	}
	return pairs
}

// assignField dispatches a single field assignment using the best available strategy.
func assignField(srcField, destField reflect.Value) error {
	srcBase := baseType(srcField.Type())
	destBase := baseType(destField.Type())
	if srcBase.Kind() == reflect.Map && destBase.Kind() == reflect.Map {
		return shallowCopyMap(srcField, destField)
	}
	if srcBase.Kind() == reflect.Struct && srcBase == destBase {
		return copyNestedStruct(srcField, destField)
	}
	if err, ok := tryDirectAssign(srcField, destField); ok {
		return err
	}
	if conv, ok := lookupConverter(srcField.Type(), destField.Type()); ok {
		return applyConverter(conv, srcField, destField)
	}
	return errors.Errorf("no mapping strategy from %v to %v", srcField.Type(), destField.Type())
}

// shallowCopyMap copies all entries from srcField into a freshly allocated map set on destField.
func shallowCopyMap(srcField, destField reflect.Value) error {
	srcMap := resolvePtr(srcField)
	if !srcMap.IsValid() || srcMap.IsNil() {
		return nil
	}
	destMapType := baseType(destField.Type())
	newMap := reflect.MakeMap(destMapType)
	if err := fillMapEntries(srcMap, newMap, destMapType); err != nil {
		return err
	}
	return setMapField(destField, newMap)
}

// fillMapEntries copies each key/value pair from srcMap into newMap.
func fillMapEntries(srcMap, newMap reflect.Value, destMapType reflect.Type) error {
	for _, key := range srcMap.MapKeys() {
		copiedKey, err := assignMapEntry(key, destMapType.Key())
		if err != nil {
			return errors.Errorf("map key %v: %w", key, err)
		}
		copiedVal, err := assignMapEntry(srcMap.MapIndex(key), destMapType.Elem())
		if err != nil {
			return errors.Errorf("map value at key %v: %w", key, err)
		}
		newMap.SetMapIndex(copiedKey, copiedVal)
	}
	return nil
}

// assignMapEntry assigns val into a temporary slot of targetType using the full
// assignField dispatch chain.
func assignMapEntry(val reflect.Value, targetType reflect.Type) (reflect.Value, error) {
	tmp := reflect.New(targetType).Elem()
	if err := assignField(val, tmp); err != nil {
		return reflect.Value{}, err
	}
	return tmp, nil
}

// setMapField assigns newMap into destField, wrapping in a pointer when required.
func setMapField(destField, newMap reflect.Value) error {
	if destField.Type().Kind() == reflect.Ptr {
		ptr := reflect.New(newMap.Type())
		ptr.Elem().Set(newMap)
		destField.Set(ptr)
		return nil
	}
	destField.Set(newMap)
	return nil
}

// tryDirectAssign covers same-type, pointer-to-value, value-to-pointer, and assignable cases.
func tryDirectAssign(srcField, destField reflect.Value) (error, bool) {
	srcType, destType := srcField.Type(), destField.Type()
	if srcType == destType {
		destField.Set(srcField)
		return nil, true
	}
	if srcType.Kind() == reflect.Ptr && srcType.Elem() == destType {
		if !srcField.IsNil() {
			destField.Set(srcField.Elem())
		}
		return nil, true
	}
	if destType.Kind() == reflect.Ptr && destType.Elem() == srcType {
		ptr := reflect.New(srcType)
		ptr.Elem().Set(srcField)
		destField.Set(ptr)
		return nil, true
	}
	if srcType.AssignableTo(destType) {
		destField.Set(srcField)
		return nil, true
	}
	return nil, false
}

// copyNestedStruct recursively copies a struct field, handling pointer wrapping/unwrapping.
func copyNestedStruct(srcField, destField reflect.Value) error {
	srcStruct := resolvePtr(srcField)
	if !srcStruct.IsValid() {
		return nil
	}
	if destField.Type().Kind() == reflect.Ptr {
		newDest := reflect.New(destField.Type().Elem())
		errs := mapFields(srcStruct, newDest.Elem())
		destField.Set(newDest)
		return stdErr.Join(errs...)
	}
	return stdErr.Join(mapFields(srcStruct, destField)...)
}

// applyConverter invokes a registered converter and sets the result on destField.
func applyConverter(conv func(reflect.Value) (reflect.Value, error), srcField, destField reflect.Value) error {
	result, err := conv(srcField)
	if err != nil {
		return err
	}
	destField.Set(result)
	return nil
}

// lookupConverter searches the registry for a converter from srcType to destType.
func lookupConverter(srcType, destType reflect.Type) (Converter, bool) {
	if m, ok := converterRegistry[srcType]; ok {
		if conv, ok := m[destType]; ok {
			return conv, true
		}
	}
	return nil, false
}

// synthesizePointerConverter wraps baseConv so it handles any combination of
// pointer src/dest, dereferencing src and re-wrapping dest as needed.
func synthesizePointerConverter(srcType, destType, baseDest reflect.Type, baseConv Converter) Converter {
	srcIsPtr := srcType.Kind() == reflect.Ptr
	destIsPtr := destType.Kind() == reflect.Ptr
	return func(in reflect.Value) (reflect.Value, error) {
		src := derefConverterSrc(in, srcIsPtr)
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

// derefConverterSrc dereferences a pointer src for converter dispatch.
// Returns an invalid Value when the pointer is nil.
func derefConverterSrc(in reflect.Value, isPtr bool) reflect.Value {
	if !isPtr {
		return in
	}
	if in.IsNil() {
		return reflect.Value{}
	}
	return in.Elem()
}

// resolvePtr dereferences pointer chains, returning an invalid Value for nil pointers.
func resolvePtr(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

// baseType strips all pointer indirection to reach the underlying type.
func baseType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}
