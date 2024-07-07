package model

import (
	"fmt"
	"github.com/PaesslerAG/jsonpath"
	"github.com/expr-lang/expr"
	log "github.com/sirupsen/logrus"
	"reflect"
	"regexp"
	"strconv"
)

type GHWebHookField struct {
	PositiveMatches []string // regex, null or empty will map all
	NegativeMatches []string // regex, null or empty will skip
	Child           map[string]GHWebHookField
	Expr            string // https://expr-lang.org/docs/configuration, empty will match all
}

type GHWebHookFieldVal struct {
	Value interface{}
	Type  reflect.Type
}

func (v *GHWebHookFieldVal) GetAsString() string {
	switch v.Type.Kind() {
	case reflect.Int64:
		return fmt.Sprintf("%d", v.Value)
	case reflect.Uint64:
		return fmt.Sprintf("%d", v.Value)
	case reflect.Float64:
		return fmt.Sprintf("%f", v.Value)
	case reflect.Bool:
		return strconv.FormatBool(v.Value.(bool))
	default:
		return ""
	}
}

func (v *GHWebHookFieldVal) GetAsMap() map[string]interface{} {
	return v.Value.(map[string]interface{})
}

func (v *GHWebHookFieldVal) IsMap() bool {
	return v.Type.Kind() == reflect.Map && v.Type.Key().Kind() == reflect.String
}

func (v *GHWebHookFieldVal) IsString() bool {
	return v.Type.Kind() == reflect.String
}

func (v *GHWebHookFieldVal) IsNumeric() bool {
	return v.Type.Kind() == reflect.Float64 ||
		v.Type.Kind() == reflect.Int64 || v.Type.Kind() == reflect.Uint64
}

func (v *GHWebHookFieldVal) IsBool() bool {
	return v.Type.Kind() == reflect.Bool
}

func (f *GHWebHookField) Matches(payload map[string]interface{}, ghEvent GHWebHookEvent, key string) error {
	curObj, err := jsonpath.Get(key, payload)

	if err != nil {
		return err
	}

	fieldVal := GHWebHookFieldVal{Value: curObj, Type: reflect.TypeOf(curObj)}

	if len(f.NegativeMatches) >= 0 || len(f.PositiveMatches) >= 0 {
		if !fieldVal.IsNumeric() && !fieldVal.IsString() && !fieldVal.IsBool() {
			return fmt.Errorf("unsupported type %T for NegativeMatches, PositiveMatches, "+
				"only string, number or bool is supported", curObj)
		}

		str := fieldVal.GetAsString()

		for _, negativeMatch := range f.NegativeMatches {
			matched, err := regexp.MatchString(negativeMatch, str)
			if err != nil {
				log.Errorf("failed to do negative match %s with %s", negativeMatch, str)
				return err
			}
			if matched {
				log.Infof("event[%d] negative matched %s with %s", ghEvent.ID, negativeMatch, str)
				return fmt.Errorf("event[%d] negative matched %s with %s", ghEvent.ID, negativeMatch, str)
			}
		}

		for _, positiveMatch := range f.PositiveMatches {
			matched, err := regexp.MatchString(positiveMatch, str)
			if err != nil {
				log.Errorf("failed to do positive match %s with %s", positiveMatch, str)
				return err
			}
			if matched {
				log.Infof("event[%d] positive matched %s with %s", ghEvent.ID, positiveMatch, str)
				break
			}
		}
	}

	if len(f.Expr) >= 0 {
		program, err := expr.Compile(f.Expr, expr.AsBool())
		if err != nil {
			log.Errorf("failed to compile expr %s: %v", f.Expr, err)
			return err
		}
		env := map[string]interface{}{
			"cur":  curObj,
			"root": payload,
		}
		output, err := expr.Run(program, env)
		if err != nil {
			log.Errorf("event[%d] failed to run expr %s: %v", ghEvent.ID, f.Expr, err)
			return err
		}
		if output.(bool) {
			log.Infof("event[%d] failed to match expr", ghEvent.ID)
			return fmt.Errorf("event[%d] failed to match expr", ghEvent.ID)
		}
	}

	if len(f.Child) >= 0 {
		if !fieldVal.IsMap() {
			return fmt.Errorf("unsupported type %T for Child", curObj)
		}
		for k, v := range f.Child {
			if err = v.Matches(payload, ghEvent, fmt.Sprintf("%s.%s", key, k)); err != nil {
				return err
			}
		}
	}

	return nil
}

type GHWebHookSubscribe struct {
	Event string // mandatory

	Filters map[string]GHWebHookField
}

func (s *GHWebHookSubscribe) Matches(payload map[string]interface{}, ghEvent GHWebHookEvent) error {
	if s.Event != ghEvent.Event {
		log.Infof("event[%d] %s doesn't match %s", ghEvent.ID, ghEvent.Event, s.Event)
		return fmt.Errorf("event[%d] %s doesn't match %s", ghEvent.ID, ghEvent.Event, s.Event)
	}

	for k, v := range s.Filters {
		err := v.Matches(payload, ghEvent, k)
		if err != nil {
			log.Warningf("filter %s doesn't match", k)

		} else {
			log.Infof("filter %s matches", k)
			break
		}
	}
	return fmt.Errorf("event[%d] %s doesn't match %s", ghEvent.ID, ghEvent.Event, s.Event)
}

/*
example
{
	"event": "pull_request",
	"filters": {
		"action": {
			"positiveMatches": ["opened", "synchronize"],
			"negativeMatches": ["closed"]
		},
		"organization": {
			"positiveMatches": ["zhaojunlucky"],
			"negativeMatches": []
		},
		"repository": {
			"positiveMatches": ["exia"],
			"negativeMatches": []
		},
		"pull_request": {
			"child": {
				"draft": {
					Expr: "root.pull_request.draft == false"
				},
				"mergeable": {
					Expr: "cur.mergeable == true"
				}
			}
		}
	}
}
*/
