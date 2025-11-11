package httpx

import (
	"chatx-01/pkg/errjon"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strconv"
)

func BindRequest[R interface{ Validate() error }](r *http.Request) (R, error) {
	const op = "BindRequest"
	var req R

	// Get the value and type of the request struct
	reqVal := reflect.ValueOf(&req).Elem()
	reqType := reqVal.Type()

	// Bind path and query parameters
	for i := 0; i < reqType.NumField(); i++ {
		field := reqType.Field(i)
		fieldVal := reqVal.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		// Bind path parameter
		if pathTag := field.Tag.Get("path"); pathTag != "" {
			pathValue := r.PathValue(pathTag)
			if pathValue != "" {
				if err := setFieldValue(fieldVal, pathValue); err != nil {
					return req, errjon.Wrap(op, errjon.AddFieldError(nil, pathTag, err.Error()))
				}
			}
		}

		// Bind query parameter
		if queryTag := field.Tag.Get("query"); queryTag != "" {
			queryValue := r.URL.Query().Get(queryTag)
			if queryValue != "" {
				if err := setFieldValue(fieldVal, queryValue); err != nil {
					return req, errjon.Wrap(op, errjon.AddFieldError(nil, queryTag, err.Error()))
				}
			}
		}
	}

	// Bind JSON body if content type is application/json
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return req, errjon.Wrap(op, err)
		}
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		return req, errjon.Wrap(op, err)
	}

	return req, nil
}

// setFieldValue sets a field value from a string based on its type.
func setFieldValue(field reflect.Value, value string) error {
	const op = "setFieldValue"
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return errjon.Wrap(op, err)
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return errjon.Wrap(op, err)
		}
		field.SetUint(uintVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return errjon.Wrap(op, err)
		}
		field.SetBool(boolVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return errjon.Wrap(op, err)
		}
		field.SetFloat(floatVal)
	default:
		// Unsupported type
		return errjon.Wrap(op, errors.New("unsupported type"))
	}
	return nil
}
