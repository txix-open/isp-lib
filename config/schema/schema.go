package schema

/*import (
	"reflect"
	"errors"
	"github.com/integration-system/isp-lib/utils"
	"strings"
	"github.com/asaskevich/govalidator"
	"unicode"
)

const (
	Integer = "INT"
	Double  = "DOUBLE"
	Boolean = "BOOLEAN"
	String  = "STRING"
	Json    = "JSON"
	Any     = "ANY"
	Object  = "OBJECT"
	Array   = "ARRAY"
)

var (
	emptyValidators []Validator
)

type Schema map[string]interface{}

type ConfigSchema struct {
	Version string `json:"version"`
	Schema  Schema `json:"schema"`
}

type Validator struct {
	Function string        `json:"function"`
	Args     []interface{} `json:"args"`
}

type FieldDescriptor struct {
	Type       interface{} `json:"type"`
	Descriptor interface{} `json:"descriptor"`
	Validators []Validator `json:"validators"`
}

func GenerateConfigSchema(cfgPtr interface{}) Schema {
	rt := reflect.TypeOf(cfgPtr)
	if rt.Kind() != reflect.Ptr || rt.Elem().Kind() != reflect.Struct {
		panic(errors.New("Expecting poiter to configu struct"))
	}
	s := generateSchema(rt.Elem(), nil).(*FieldDescriptor)
	return s.Descriptor.(Schema)
}

func generateSchema(rt reflect.Type, field *reflect.StructField) interface{} {
	kind := rt.Kind()
	switch kind {
	case reflect.Bool:
		return &FieldDescriptor{Type: Boolean, Validators: generateValidators(field)}
	case reflect.Int, reflect.Int8,
		reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint,
		reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		return &FieldDescriptor{Type: Integer, Validators: generateValidators(field)}
	case reflect.Float32, reflect.Float64:
		return &FieldDescriptor{Type: Double, Validators: generateValidators(field)}
	case reflect.Array, reflect.Slice:
		elemType := rt.Elem()
		return &FieldDescriptor{
			Type:       Array,
			Validators: generateValidators(field),
			Descriptor: generateSchema(elemType, field),
		}
	case reflect.Interface:
		return &FieldDescriptor{Type: Any, Validators: generateValidators(field)}
	case reflect.Map:
		return &FieldDescriptor{Type: Json, Validators: generateValidators(field)}
	case reflect.Ptr:
		return generateSchema(rt.Elem(), field)
	case reflect.String:
		return &FieldDescriptor{Type: String, Validators: generateValidators(field)}
	case reflect.Struct:
		num := rt.NumField()
		s := make(Schema, num)
		for i := 0; i < num; i++ {
			f := rt.Field(i)
			field := &f
			if name, accept := utils.GetFieldName(f); accept {
				if descriptor := generateSchema(field.Type, field); descriptor != nil {
					s[name] = descriptor
				}
			}
		}
		return &FieldDescriptor{Type: Object, Validators: generateValidators(field), Descriptor: s}
	default:
		return nil
	}

}

func generateValidators(field *reflect.StructField) []Validator {
	if field == nil {
		return emptyValidators
	}
	value, ok := field.Tag.Lookup("valid")
	value = strings.TrimSpace(value)
	if !ok || value == "" || value == "-" {
		return emptyValidators
	}
	tm := parseTagIntoMap(value)
	result := make([]Validator, len(tm))
	i := 0
	for val := range tm {
		if v := generateValidator(val); v != nil {
			result[i] = *v
			i++
		}
	}
	return result
}

func generateValidator(val string) (*Validator) {
	for key, value := range govalidator.ParamTagRegexMap {
		ps := value.FindStringSubmatch(val)
		l := len(ps)
		if l < 2 {
			continue
		}
		args := make([]interface{}, l-1)
		for i := 1; i < l; i++ {
			args[i-1] = ps[i]
		}
		return &Validator{Function: key, Args: args}
	}
	return &Validator{Function: val}
}

type tagOptionsMap map[string]string

func parseTagIntoMap(tag string) tagOptionsMap {
	optionsMap := make(tagOptionsMap)
	options := strings.Split(tag, ",")

	for _, option := range options {
		option = strings.TrimSpace(option)

		validationOptions := strings.Split(option, "~")
		if !isValidTag(validationOptions[0]) {
			continue
		}
		if len(validationOptions) == 2 {
			optionsMap[validationOptions[0]] = validationOptions[1]
		} else {
			optionsMap[validationOptions[0]] = ""
		}
	}
	return optionsMap
}

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("\\'\"!#$%&()*+-./:<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		default:
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}
*/
