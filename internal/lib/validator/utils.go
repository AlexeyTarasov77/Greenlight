package validator

import (
	"fmt"
	"greenlight/proj/internal/domain/models"
	"greenlight/proj/internal/utils"
	"reflect"
	"strings"

	govalidator "github.com/go-playground/validator/v10"
)

func getFieldName(obj any, origFieldName string) (fieldName string) {
	t := reflect.TypeOf(obj)
	field, found := t.FieldByName(origFieldName)
	if !found {
		panic(fmt.Sprintf("Field %s not found in type %s", origFieldName, t.Name()))
	}
	if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
		jsonName := strings.Split(tag, ",")[0]
		if jsonName != "" {
			fieldName = jsonName
		}
	} else {
		fieldName = utils.CamelToSnake(origFieldName)
	}
	return
}

func ProcessValidationErrors(obj any, errs govalidator.ValidationErrors) map[string]string {
	processedErrors := make(map[string]string)
	for _, e := range errs {
		processedErrors[getFieldName(obj, e.StructField())] = GetErrorMsgForField(obj, e)
	}
	return processedErrors
}

func ValidateStruct(validator *govalidator.Validate, obj any) (validationErrs map[string]string) {
	if err := validator.Struct(obj); err != nil {
		validationErrs = ProcessValidationErrors(obj, err.(govalidator.ValidationErrors))
	}
	return
}

func GetErrorMsgForField(obj any, err govalidator.FieldError) (errorMsg string) {
	t := reflect.TypeOf(obj)
	field, found := t.FieldByName(err.StructField())
	if !found {
		panic(fmt.Sprintf("Field %s not found in type %s", err.StructField(), t.Name()))
	}
	errorMsg = field.Tag.Get("errorMsg")
	if errorMsg == "" {
		switch err.Tag() {
		case "required":
			errorMsg = "This field is required"
		case "max":
			errorMsg = fmt.Sprintf("The maximum value is %s", err.Param())
		case "min":
			errorMsg = fmt.Sprintf("The minimum value is %s", err.Param())
		case "gte":
			errorMsg = fmt.Sprintf("Value should be greater than or equal to %s", err.Param())
		case "lte":
			errorMsg = fmt.Sprintf("Value should be less than or equal to %s", err.Param())
		case "lt":
			errorMsg = fmt.Sprintf("Value should be less than %s", err.Param())
		case "gt":
			errorMsg = fmt.Sprintf("Value should be greater than %s", err.Param())
		case "eqfield", "eq":
			errorMsg = fmt.Sprintf("Value should be equal to %s", err.Param())
		case "nefield", "ne":
			errorMsg = fmt.Sprintf("Value should not be equal to %s", err.Param())
		case "oneof":
			errorMsg = fmt.Sprintf("Value should be one of %s", err.Param())
		case "nooneof":
			errorMsg = fmt.Sprintf("Value should not be one of %s", err.Param())
		case "len":
			errorMsg = fmt.Sprintf("Length should be equal to %s", err.Param())
		case "unique":
			errorMsg = "Value must not contain duplicate values"
		case "url":
			errorMsg = "Value must be a valid URL"
		case "email":
			errorMsg = "Value must be a valid email address"
		case "alphanum":
			errorMsg = "Value must be alphanumeric"
		case "sortbymoviefield":
			errorMsg = "Value must be a name of one of the movie fields (e.g. +title, -year, etc...)"
		default:
			errorMsg = "This field is invalid"
		}
	}
	return
}

// CUSTOM VALIDATORS

func ValidateSortByMovieField(fl govalidator.FieldLevel) bool {
	sort := fl.Field().String()
	t := reflect.TypeOf(models.Movie{})
	sort = strings.TrimPrefix(sort, "-")
	fieldName := strings.ToUpper(string(sort[0])) + sort[1:]
	fmt.Println(fieldName)
	if _, ok := t.FieldByNameFunc(func(s string) bool { return strings.EqualFold(fieldName, s) }); !ok {
		return false
	}
	return true
}