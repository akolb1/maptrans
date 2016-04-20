package maptrans

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/goinggo/mapstructure"
)

// TranslationType identifies type of element translation to perform.
type TranslationType int

const (
	// CustomTranslation (default) means that a function should be provided for a
	// translation
	CustomTranslation TranslationType = iota
	// MapTranslation means that translation defines an embedded object
	MapTranslation
	// MapArrayTranslation means that the translation defines an array of
	// objects
	MapArrayTranslation
	// ModifyTranslation modifies the map based on the input value
	ModifyTranslation
	// InsertTranslation inserts a missing value
	InsertTranslation
)

// MapFunc is a function that converts one interface to another
type MapFunc func(interface{}) (interface{}, error)

// ModFunc takes an map and a value and modifies the map. It
// returns the resulting (modified) map).
type ModFunc func(map[string]interface{}, map[string]interface{},
	interface{}) error

// InsertFunc is used to insert a new element into the map.
// Parameters:
//   Original map
//   Name of the destination element
// Returns: a value that will be inserted in the map using TargetName.
type InsertFunc func(map[string]interface{}, map[string]interface{},
	string) (interface{}, error)

// MapElement defines translation definition
type MapElement struct {
	InsertFunc     InsertFunc             // Function to insert element
	Mandatory      bool                   // The field must be present if true
	MapFunc        MapFunc                // Function that maps value to new value
	ModFunc        ModFunc                // Function for object modification
	SubTranslation map[string]interface{} // Subtranslation map for children
	TargetName     string                 // Name of destination field
	Type           TranslationType        // Type of translation
}

// Custom errors

// InternalError is a programming error - it should never happen
type InternalError struct {
	Reason string
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("internal error: %s", e.Reason)
}

// NewInternalError returns an instance of an internal error with specified reason
func NewInternalError(reason string) *InternalError {
	return &InternalError{Reason: reason}
}

// MissingAttributeError is caused by a map attribute that is mandatory but is
// missing
type MissingAttributeError struct {
	Name string
}

func (e *MissingAttributeError) Error() string {
	return fmt.Sprintf("missing mandatory attribute '%s'", e.Name)
}

// NewMissingAttributeError returns an instance of an error for a missing attribute
func NewMissingAttributeError(name string) *MissingAttributeError {
	return &MissingAttributeError{Name: name}
}

// InvalidPropertyError is an error indicating that a user-provided parameter
// is bad.
type InvalidPropertyError struct {
	Name   string
	Reason string
}

func (e *InvalidPropertyError) Error() string {
	return fmt.Sprintf("property '%s' is invalid: %s", e.Name, e.Reason)
}

// NewInvalidProp returns an instance of InvalidPropertyError
func NewInvalidProp(name string, reason string) *InvalidPropertyError {
	return &InvalidPropertyError{Name: name, Reason: reason}
}

var (
	validUUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

	validID = regexp.MustCompile(`^[a-zA-Z_]+[0-9a-zA-Z_]*$`)
)

// Translate converts source map[string]interface{} to destination
// map[string]interface{} using specified description.
func Translate(src map[string]interface{},
	descr map[string]interface{}) (map[string]interface{}, error) {
	if descr == nil {
		// nil description interpreted as 'no translation'
		return src, nil
	}
	result := map[string]interface{}{}
	// Check whether any mandatory field is missing
	for k, v := range descr {
		_, isString := v.(string)
		if isString {
			continue // Nothing to do
		}
		md, ok := v.(MapElement)
		if !ok {
			return nil, NewInternalError(fmt.Sprintf("%v is not MapElement", v))
		}
		if !md.Mandatory {
			continue // Nothing to do
		}
		_, isPresent := src[k]
		if !isPresent {
			return nil, NewMissingAttributeError(k)
		}
	}

	// Walk over all fields present in the source and translate them
	for k, v := range src {
		mapDescr, ok := descr[k]
		// Ignore any description that isn't present in the source
		if !ok {
			continue
		}
		// The description can be either a string or MapElement
		// For string do string conversion
		if stringConversion, ok := mapDescr.(string); ok {
			dstStr, err := StringMap(v)
			if err != nil {
				return result, NewInvalidProp(k, err.Error())
			}
			// Save destination in the specified string
			result[stringConversion] = dstStr
			continue
		}
		md, ok := mapDescr.(MapElement)
		if !ok {
			return nil, NewInternalError(
				fmt.Sprintf("%v is not a MapElement", mapDescr))
		}
		if md.TargetName == "" {
			// By default preserve the attribute name
			md.TargetName = k
		}
		switch md.Type {
		case CustomTranslation:
			if md.MapFunc == nil {
				return nil,
					NewInternalError("missing translation func for " + k)
			}
			dstStr, err := md.MapFunc(v)
			if err != nil {
				return nil, NewInvalidProp(k, err.Error())
			}
			// Save destination in the specified string
			result[md.TargetName] = dstStr
		case MapTranslation:
			// v can be either map[string]interface{} or map[string]interface{}
			srcMap, ok := v.(map[string]interface{})
			if !ok {
				srcMap, ok := v.(map[string]interface{})
				if !ok {
					return nil, NewInternalError(
						fmt.Sprintf("invalid type for %v: %T", v, v))
				}
				srcMap = map[string]interface{}(srcMap)
			}
			trans, transErr := Translate(srcMap, md.SubTranslation)
			if transErr != nil {
				return nil, transErr
			}
			result[md.TargetName] = trans
		case MapArrayTranslation:
			// Translate [ {... }, {...} ]
			srcMaps := []map[string]interface{}{}
			err := mapstructure.Decode(v, &srcMaps)
			if err != nil {
				return nil, NewInternalError(err.Error())
			}
			res := make([]map[string]interface{}, len(srcMaps))
			for i, val := range srcMaps {
				trans, transErr := Translate(val, md.SubTranslation)
				if transErr != nil {
					return nil, transErr
				}
				res[i] = trans
			}
			result[md.TargetName] = res
		case ModifyTranslation:
			// Modify result based on value
			if md.ModFunc == nil {
				return nil,
					NewInternalError("missing translation func for " + k)
			}
			err := md.ModFunc(src, result, v)
			if err != nil {
				return nil, NewInvalidProp(k, err.Error())
			}
		case InsertTranslation:
			// InsertTranslation is only used for missing fields
			continue
		default:
			return nil, NewInternalError("Invalid Translation type")
		}
	}

	// Now check whether any value should be inserted
	for k, v := range descr {
		_, isString := v.(string)
		if isString {
			continue // Nothing to do
		}
		md, ok := v.(MapElement)
		if !ok {
			return nil, NewInternalError(
				fmt.Sprintf("%v is not a MapElement", v))
		}
		if md.Type != InsertTranslation {
			continue
		}
		if md.InsertFunc == nil {
			return nil,
				NewInternalError("missing translation func for " + k)
		}

		// Skip anything that is already present
		if _, isPresent := result[md.TargetName]; isPresent {
			continue
		}

		// Get the value to insert
		val, err := md.InsertFunc(src, result, k)
		if err != nil {
			return nil, err
		}
		// Insert result
		result[md.TargetName] = val
	}
	return result, nil
}

// IDMap translates an object to itself. This is the easiest way to deal with
// embedded objects.
func IDMap(src interface{}) (interface{}, error) {
	return src, nil
}

// StringMap translates string interface into a string (trimming spaces)
func StringMap(src interface{}) (interface{}, error) {
	if srcStr, ok := src.(string); ok {
		return strings.TrimSpace(srcStr), nil
	}
	return "", fmt.Errorf("invalid type %T for %v", src, src)
}

// StringToLowerMap translates string interface into a string with lower case
func StringToLowerMap(src interface{}) (interface{}, error) {
	if srcStr, ok := src.(string); ok {
		return strings.TrimSpace(strings.ToLower(srcStr)), nil
	}
	return "", fmt.Errorf("invalid type %T for %v", src, src)
}

// StringToUpperMap translates string interface into a string with upper case
func StringToUpperMap(src interface{}) (interface{}, error) {
	if srcStr, ok := src.(string); ok {
		return strings.TrimSpace(strings.ToUpper(srcStr)), nil
	}
	return "", fmt.Errorf("invalid type %T for %v", src, src)
}

// IdentifierMap is similar to StringMap but verifies that the string
// contains only valid characters for identifiers
func IdentifierMap(src interface{}) (interface{}, error) {
	srcStr, ok := src.(string)
	if !ok {
		return "", fmt.Errorf("%v is not a string", srcStr)
	}
	if !validID.MatchString(srcStr) {
		return "", fmt.Errorf("%s is not a valid identifier", srcStr)
	}

	return strings.TrimSpace(srcStr), nil
}

// IPAddrMap verifies that the argument is a valid IP address
func IPAddrMap(src interface{}) (interface{}, error) {
	srcStr, ok := src.(string)
	if !ok {
		return "", fmt.Errorf("%v is not a string", srcStr)
	}
	srcStr = strings.TrimSpace(srcStr)
	if net.ParseIP(srcStr) == nil {
		return "", fmt.Errorf("%s is not a valid IP address", srcStr)
	}

	return srcStr, nil
}

// CIDRMap verifies that the argument is a valid IP address in CIDR notation
// notation
func CIDRMap(src interface{}) (interface{}, error) {
	srcStr, ok := src.(string)
	if !ok {
		return "", fmt.Errorf("%v is not a string", srcStr)
	}
	srcStr = strings.TrimSpace(srcStr)
	if _, _, err := net.ParseCIDR(srcStr); err == nil {
		return srcStr, nil // Valid address
	}
	return "", fmt.Errorf("%s is not a valid CIDR address", srcStr)
}

// BoolMap translates boolean interface into a boolean
func BoolMap(src interface{}) (interface{}, error) {
	val, ok := src.(bool)
	if ok {
		return val, nil
	}
	strVal, ok := src.(string)
	if !ok {
		return "", errors.New("invalid type")
	}
	result, err := strconv.ParseBool(strVal)
	if err != nil {
		return false, fmt.Errorf("invalid value '%s' for boolean", strVal)
	}
	return result, nil
}

// BoolToStrMap translates boolean interface into a string
func BoolToStrMap(src interface{}) (interface{}, error) {
	b, err := BoolMap(src)
	if err != nil {
		return nil, err
	}
	if val, _ := b.(bool); val {
		return "True", nil
	}
	return "False", nil
}

// IntegerMap Converts numbers to strings
func IntegerMap(val interface{}) (interface{}, error) {
	switch val := val.(type) {
	case int:
		if val < 0 {
			return "", fmt.Errorf("%v should be non-negative", val)
		}
		i := uint64(val)
		return strconv.FormatUint(i, 10), nil // convert to string
	case uint32:
		return strconv.FormatUint(uint64(val), 10), nil // convert to string
	case string:
		result, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return false, fmt.Errorf("invalid value '%s' for an integer",
				val)
		}
		return strconv.FormatUint(result, 10), nil
	case float64:
		if val < 0 {
			return "", fmt.Errorf("%v should be non-negative", val)
		}
		i := uint64(val)
		return strconv.FormatUint(i, 10), nil // convert to string
	}
	return nil, fmt.Errorf("invalid type %t for value %v", val, val)
}

// UUIDMap translates UUID values and verifies that they are legal
func UUIDMap(src interface{}) (interface{}, error) {
	srcStr, ok := src.(string)
	if !ok {
		return "", fmt.Errorf("%v is not a string", srcStr)
	}
	srcStr = strings.TrimSpace(srcStr)
	if !validUUID.MatchString(srcStr) {
		return "", fmt.Errorf("%v is not a valid UUID", srcStr)
	}
	return srcStr, nil
}

// StringArrayMap translates array of strings
func StringArrayMap(src interface{}) (interface{}, error) {
	if src == nil {
		return nil, nil
	}
	result := []string{}
	if err := mapstructure.Decode(src, &result); err != nil {
		return "", fmt.Errorf("invalid argument type: %v", err)
	}
	return result, nil
}

// IsSimilar verifies that dst object matches src object according to
// description
func IsSimilar(src map[string]interface{}, dst map[string]interface{},
	descr map[string]interface{}) (bool, error) {

	for k, vSrc := range src {
		mapDescr, ok := descr[k]
		if !ok {
			continue
		}

		// The description can be either a string or MapElement
		// For string do string conversion
		if stringConversion, ok := mapDescr.(string); ok {
			srcStr, ok := vSrc.(string)
			if !ok {
				return false,
					NewInternalError(
						fmt.Sprintf("Invalid description value %v", vSrc))
			}
			dstStr, ok := dst[stringConversion].(string)
			if !ok {
				return false,
					NewInternalError(
						fmt.Sprintf("Missing value for %s", stringConversion))
			}
			if srcStr != dstStr {
				return false,
					fmt.Errorf("Values %s and %s don't match", srcStr, dstStr)
			}
			continue
		}
		md, ok := mapDescr.(MapElement)
		if !ok {
			return false, NewInternalError(
				fmt.Sprintf("invalid description %v", mapDescr))
		}
		switch md.Type {
		case MapTranslation:
			srcMap, ok := vSrc.(map[string]interface{})
			if !ok {
				return false, fmt.Errorf("Invalid source object %v", vSrc)
			}
			dstMapVal, ok := dst[md.TargetName]
			if !ok {
				return false,
					fmt.Errorf("Missing value for %s in %v", md.TargetName, dst)
			}
			dstMap, ok := dstMapVal.(map[string]interface{})
			if !ok {
				dstMap, ok = dstMapVal.(map[string]interface{})
				if !ok {
					return false,
						fmt.Errorf("Invalid Type for %s: %T",
							md.TargetName, dstMapVal)
				}
			}
			r, err := IsSimilar(srcMap, dstMap, md.SubTranslation)
			if !r {
				return false, err
			}
		case MapArrayTranslation:
			srcMaps := []map[string]interface{}{}
			err := mapstructure.Decode(vSrc, &srcMaps)
			if err != nil {
				return false, fmt.Errorf("Invalid source object %v: %v", vSrc, err)
			}
			_, ok := dst[md.TargetName]
			if !ok {
				return false,
					fmt.Errorf("Missing value for %s in %v", md.TargetName, dst)
			}
			dstMaps := []map[string]interface{}{}
			e2 := mapstructure.Decode(dst[md.TargetName], &dstMaps)
			if e2 != nil {
				return false, fmt.Errorf("Invalid destination object %v",
					dst[md.TargetName])
			}
			if len(srcMaps) != len(dstMaps) {
				return false,
					fmt.Errorf("Source and destination length: %d!= %d",
						len(srcMaps), len(dstMaps))
			}
			for i, val := range srcMaps {
				r, err := IsSimilar(val, dstMaps[i], md.SubTranslation)
				if !r {
					return false, err
				}
			}
		default:
			return false, fmt.Errorf("Unsupported translation type %v", md.Type)
		}
	}
	return true, nil
}
