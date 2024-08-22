package decoder

// import (
// 	"errors"
// 	"greenlight/proj/internal/utils"
// 	"reflect"
// 	"strings"
// )


// type URLDecoder struct {
// 	ignoreUnknownKeys bool
// }

// func New() *URLDecoder {
// 	return &URLDecoder{}
// }

// func (d *URLDecoder) IgnoreUnknownKeys(i bool) {
// 	d.ignoreUnknownKeys = i
// }

// func (d *URLDecoder) Decode(dst any, src map[string][]string) error {
// 	ref := reflect.ValueOf(dst)
// 	if ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
// 		return errors.New("url decoder: interface must be a pointer to struct")
// 	}
// 	v := ref.Elem()
// 	t := v.Type()
// 	fieldsToSrc := make(map[string][2]string)
// 	for i := 0; i < v.NumField(); i++ {
// 		f := t.Field(i)
// 		// first element is source field name, second is default value for that field
// 		tag := strings.Split(f.Tag.Get("decode"), ",")
// 		if tag[0] == "-" {
// 			continue
// 		}
// 		srcFieldName := tag[0]
// 		if srcFieldName == "" {
// 			srcFieldName = utils.CamelToSnake(f.Name)
// 		}
// 		fieldsToSrc[srcFieldName] = f.Name
// 	}
// 	for key, value := range src {
// 		dstFieldName, ok := fieldsToSrc[key]
// 		if !ok {
// 			if d.ignoreUnknownKeys {
// 				continue
// 			} else {
// 				return errors.New("url decoder: unknown field " + key)
// 			}
// 		}
// 		f := v.FieldByName(dstFieldName)
// 		if !f.CanSet() {
// 			return errors.New("url decoder: cannot set field " + key)
// 		}
// 		if err := d.decode(f, value); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (d *URLDecoder) decode(f reflect.Value, value []string, defaultValue any) error {
// 	if len(value) == 0 {
// 		return nil
// 	}
// 	switch f.Kind() {
// 	case reflect.String:
// 		f.SetString(value[0])
// 	case reflect.Int:
// 		intValue, err := strconv.Atoi(value[0])
// 		if err != nil {
// 			return err
// 		}
// 		f.SetInt(int64(intValue))
// 	case reflect.Bool:
// 		boolValue, err := strconv.ParseBool(value[0])
// 		if err != nil {
// 			return err
// 		}
// 		f.SetBool(boolValue)
// 	}
// 	return nil
// }