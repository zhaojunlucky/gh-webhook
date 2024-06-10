package core

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rbicker/go-rsql"
	log "github.com/sirupsen/logrus"
	"reflect"
	"slices"
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
	filter := c.Query("filter")
	sort := c.Query("sort")

	err = r.parseFilter(filter)
	if err != nil {
		return nil
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
		log.Fatalf("error while parsing: %s", err)
		return err
	}
	var jsonQuery map[string]interface{}

	err = json.Unmarshal([]byte(res), &jsonQuery)
	if err != nil {
		log.Fatalf("error while parsing: %s", err)
		return err
	}

	sql, arguments, err := r.parseRSQL(jsonQuery)
	if err != nil {
		log.Fatalf("error while parsing: %s", err)
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
			args := []interface{}{v[op]}
			sqls = append(sqls, sqlOp, "?")
			return sqls, args, nil
		case []interface{}:
			return nil, nil, fmt.Errorf("query must be a scalar value or a map, not %v", v)
		default:
			sqls = append(sqls, "=", "?")
			args := []interface{}{v}
			return sqls, args, nil
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

func NewRSQLHelper() *RSQLHelper {
	return &RSQLHelper{
		FieldsQuery: make(map[string]*FieldQuery),
	}
}
