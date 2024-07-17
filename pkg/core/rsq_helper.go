package core

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rbicker/go-rsql"
	log "github.com/sirupsen/logrus"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

var OperationMap = map[string]string{
	"==":   "=",
	"$ne":  "!=",
	"$gt":  ">",
	"$gte": ">=",
	"$lt":  "<",
	"$lte": "<=",
	"$in":  "in",
	"$nin": "not in",
}

type FieldQuery struct {
	AllowSort   bool
	AllowFilter bool
	Kind        reflect.Kind
}

type RSQLHelper struct {
	FilterSQL   string
	Arguments   []interface{}
	SortSQL     string
	FieldsQuery map[string]*FieldQuery
}

func (r *RSQLHelper) ParseFilter(queryType any, c *gin.Context) error {
	err := r.parseFieldsQuery(queryType)
	if err != nil {
		return err
	}

	filter, sort, err := r.getRSQLQuery(c.Request.URL.RawQuery)
	if err != nil {
		return err
	}

	err = r.parseFilter(filter)
	if err != nil {
		return err
	}

	err = r.parseSort(sort)
	return err
}

func (r *RSQLHelper) parseFilter(filter string) error {
	if len(filter) == 0 {
		r.FilterSQL = ""
		r.Arguments = make([]interface{}, 0)

		return nil
	}
	parser, err := rsql.NewParser(rsql.Mongo())
	if err != nil {
		log.Errorf("error while creating parser: %s", err)
		return fmt.Errorf("failed to parse filter: %s", err)
	}

	res, err := parser.Process(filter)
	if err != nil {
		log.Errorf("error while parsing: %s", err)
		return err
	}
	var jsonQuery map[string]interface{}

	err = json.Unmarshal([]byte(res), &jsonQuery)
	if err != nil {
		log.Errorf("error while parsing: %s", err)
		return err
	}

	sql, arguments, err := r.parseRSQL(jsonQuery)
	if err != nil {
		log.Errorf("error while parsing: %s", err)
		return err
	}
	r.FilterSQL = sql
	r.Arguments = arguments
	return nil
}

func (r *RSQLHelper) parseRSQL(query map[string]interface{}) (sql string, args []interface{}, err error) {

	sqls, args, err := r._parseRSQL(query)
	if err != nil {
		return "", nil, err
	} else {
		q := strings.Join(sqls, " ")

		return q[1 : len(q)-1], args, nil
	}
}

func (r *RSQLHelper) _parseRSQL(query map[string]interface{}) ([]string, []interface{}, error) {

	keys := reflect.ValueOf(query).MapKeys()
	if len(keys) == 0 || len(keys) > 1 {
		return nil, nil, fmt.Errorf("query must have a single key")
	}

	key := keys[0].String()
	q := query[key]

	if slices.Contains([]string{"$or", "$and"}, key) {
		op := " or "
		if key == "$and" {
			op = " and "
		}

		var subQuery []interface{}
		switch t := q.(type) {
		case []interface{}:
			for x := 0; x < len(t); x++ {
				subQuery = append(subQuery, t[x])
			}
		default:
			return nil, nil, fmt.Errorf("query must contain an array, not %v", t)
		}
		var sqls []string
		var args []interface{}
		for _, subQ := range subQuery {
			switch st := subQ.(type) {
			case map[string]interface{}:
				subSqls, subArgs, err := r._parseRSQL(st)
				if err != nil {
					return nil, nil, err
				}
				sql := fmt.Sprintf("%s", strings.Join(subSqls, " "))
				sqls = append(sqls, sql)
				args = append(args, subArgs...)
			default:
				return nil, nil, fmt.Errorf("query must contain an array of maps for key %s, not %v", key, st)
			}

		}
		return []string{fmt.Sprintf("(%s)", strings.Join(sqls, op))}, args, nil

	} else {
		sqls := []string{key}
		if !r.allowFilterField(key) {
			return nil, nil, fmt.Errorf("field %s is not allowed to filter", key)
		}

		fieldDef := r.FieldsQuery[key]
		switch v := q.(type) {
		case map[string]interface{}:
			opVal := reflect.ValueOf(v).MapKeys()
			if len(opVal) == 0 || len(opVal) > 1 {
				return nil, nil, fmt.Errorf("query %s value must be a map", key)
			}

			op := opVal[0].String()
			sqlOp, ok := OperationMap[op]
			if !ok {
				return nil, nil, fmt.Errorf("unknown operator %s", op)
			}
			var opArg, err = r.convertVal(v[op], fieldDef)

			args := []interface{}{opArg}
			sqls = append(sqls, sqlOp, "?")
			return sqls, args, err
		case []interface{}:
			return nil, nil, fmt.Errorf("query must be a scalar value or a map, not %v", v)
		default:
			sqls = append(sqls, "=", "?")
			var opArg, err = r.convertVal(v, fieldDef)

			args := []interface{}{opArg}
			return sqls, args, err
		}
	}
}

func (r *RSQLHelper) parseSort(sort string) error {
	if len(sort) == 0 {
		r.SortSQL = ""
		return nil
	}
	cols := strings.Split(sort, ";")
	var orders []string
	for _, v := range cols {
		v = strings.TrimSpace(v)
		if len(v) == 0 {
			continue
		}
		fieldDirection := strings.Split(v, ",")

		if len(fieldDirection) <= 0 || len(fieldDirection) > 2 {
			return fmt.Errorf("rsql: invalid sort %s", v)
		}
		field := strings.TrimSpace(fieldDirection[0])
		if len(field) == 0 {
			return fmt.Errorf("rsql: invalid sort %s", v)
		}

		if !r.allowSortField(field) {
			return fmt.Errorf("rsql: invalid sort field %s, it's not allowed to sort", v)
		}

		if len(fieldDirection) == 2 {
			dir := strings.ToLower(strings.TrimSpace(fieldDirection[1]))
			if dir != "asc" && dir != "desc" {
				return fmt.Errorf("rsql: invalid sort %s", v)
			}
			orders = append(orders, fmt.Sprintf("%s %s", field, dir))
		} else {
			orders = append(orders, field)
		}
	}
	if len(orders) == 0 {
		r.SortSQL = ""
	} else {
		r.SortSQL = strings.Join(orders, ",")
	}

	return nil
}

func (r *RSQLHelper) parseFieldsQuery(queryType any) error {
	typ := reflect.TypeOf(queryType)
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("%s is not a struct", typ)
	}
	for i := 0; i < typ.NumField(); i++ {
		fld := typ.Field(i)
		if qName := fld.Tag.Get("rsql"); qName != "" {
			tags := strings.Split(qName, ",")
			name := strings.TrimSpace(tags[0])
			r.FieldsQuery[name] = &FieldQuery{
				AllowFilter: false,
				AllowSort:   false,
				Kind:        fld.Type.Kind(),
			}

			for j := 1; j < len(tags); j++ {
				tag := strings.ToLower(strings.TrimSpace(tags[j]))
				if tag == "filter" {
					r.FieldsQuery[name].AllowFilter = true
				} else if tag == "sort" {
					r.FieldsQuery[name].AllowSort = true
				} else {
					log.Errorf("unknown tag %s", tag)
				}

			}
		}
	}

	return nil
}

func (r *RSQLHelper) allowSortField(field string) bool {
	if val, ok := r.FieldsQuery[field]; ok {
		return val.AllowSort
	}
	return false
}

func (r *RSQLHelper) allowFilterField(field string) bool {
	if val, ok := r.FieldsQuery[field]; ok {
		return val.AllowFilter
	}
	return false
}

func (r *RSQLHelper) getRSQLQuery(query string) (filter string, sort string, err error) {
	count := 0
	for query != "" {
		if count >= 2 {
			return
		}
		var key string
		key, query, _ = strings.Cut(query, "&")
		if key == "" {
			continue
		}

		if !strings.Contains(key, "=") {
			continue
		}

		index := strings.Index(key, "=")

		name := strings.TrimSpace(key[:index])
		if name == "filter" {
			count++

			var value string
			value, err = url.QueryUnescape(strings.TrimSpace(key[index+1:]))

			if err != nil {
				return
			}
			filter = value
		} else if name == "sort" {
			count++
			var value string
			value, err = url.QueryUnescape(strings.TrimSpace(key[index+1:]))
			if err != nil {
				return
			}
			sort = value
		}

	}
	return
}

func (r *RSQLHelper) convertVal(i interface{}, def *FieldQuery) (interface{}, error) {
	switch i.(type) {
	case int, int8, int16, int32, int64:
		return r.convertInt(i.(int64), def)
	case uint, uint8, uint16, uint32, uint64:
		return r.convertUInt(i.(uint64), def)
	case float32, float64:
		return r.convertFloat(i.(float64), def)
	case bool:
		return r.convertBool(i.(bool), def)
	case string:
		return r.convertStr(i.(string), def)

	default:
		return nil, fmt.Errorf("rsql: invalid type %T", i)
	}
}

func (r *RSQLHelper) convertInt(i int64, def *FieldQuery) (val interface{}, err error) {
	switch def.Kind {
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		val = i
	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
		val = uint64(i)
	case reflect.Float32, reflect.Float64:
		val = float64(i)
	case reflect.Bool:
		err = fmt.Errorf("rsql: invalid type %T, required type %d", i, def.Kind)
	case reflect.String:
		val = fmt.Sprintf("%d", i)
	default:
		err = fmt.Errorf("rsql: invalid type %T, required type %d", i, def.Kind)
	}
	return
}

func (r *RSQLHelper) convertUInt(u uint64, def *FieldQuery) (val interface{}, err error) {
	switch def.Kind {
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		val = int64(u)
	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
		val = u
	case reflect.Float32, reflect.Float64:
		val = float64(u)
	case reflect.Bool:
		err = fmt.Errorf("rsql: invalid type %T, required type %d", u, def.Kind)

	case reflect.String:
		val = fmt.Sprintf("%d", u)
	default:
		err = fmt.Errorf("rsql: invalid type %T, required type %d", u, def.Kind)
	}
	return
}

func (r *RSQLHelper) convertFloat(f float64, def *FieldQuery) (val interface{}, err error) {
	switch def.Kind {
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		val = int64(f)
	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
		val = uint64(f)
	case reflect.Float32, reflect.Float64:
		val = f
	case reflect.Bool:
		err = fmt.Errorf("rsql: invalid type %T, required type %d", f, def.Kind)

	case reflect.String:
		val = fmt.Sprintf("%f", f)
	default:
		err = fmt.Errorf("rsql: invalid type %T, required type %d", f, def.Kind)
	}
	return
}

func (r *RSQLHelper) convertBool(b bool, def *FieldQuery) (val interface{}, err error) {
	switch def.Kind {
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		err = fmt.Errorf("rsql: invalid type %T, required type %d", b, def.Kind)

	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
		err = fmt.Errorf("rsql: invalid type %T, required type %d", b, def.Kind)

	case reflect.Float32, reflect.Float64:
		err = fmt.Errorf("rsql: invalid type %T, required type %d", b, def.Kind)

	case reflect.Bool:
		err = fmt.Errorf("rsql: invalid type %T, required type %d", b, def.Kind)

	case reflect.String:
		val = strconv.FormatBool(b)
	default:
		err = fmt.Errorf("rsql: invalid type %T, required type %d", b, def.Kind)
	}
	return
}

func (r *RSQLHelper) convertStr(s string, def *FieldQuery) (val interface{}, err error) {
	switch def.Kind {
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		val, err = strconv.ParseInt(s, 10, 64)

	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
		val, err = strconv.ParseUint(s, 10, 64)
	case reflect.Float32, reflect.Float64:
		val, err = strconv.ParseFloat(s, 64)

	case reflect.Bool:
		val, err = strconv.ParseBool(s)

	case reflect.String:
		val = s
	default:
		err = fmt.Errorf("rsql: invalid type %T, required type %d", s, def.Kind)
	}
	return
}

func NewRSQLHelper() *RSQLHelper {
	return &RSQLHelper{
		FieldsQuery: make(map[string]*FieldQuery),
	}
}
