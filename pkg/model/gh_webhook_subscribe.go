package model

import (
	"fmt"
	"github.com/PaesslerAG/jsonpath"
	"github.com/expr-lang/expr"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"reflect"
	"regexp"
	"strconv"
)

type GHWebhookField struct {
	PositiveMatches []string // regex, null or empty will map all
	NegativeMatches []string // regex, null or empty will skip
	Child           map[string]GHWebhookField
	Expr            string // https://expr-lang.org/docs/configuration, empty will match all
}

type GHWebhookFieldVal struct {
	Value interface{}
	Type  reflect.Type
}

func (v *GHWebhookFieldVal) GetAsString() string {
	switch v.Type.Kind() {
	case reflect.Int64, reflect.Uint64, reflect.Int32, reflect.Uint32, reflect.Int16, reflect.Uint16,
		reflect.Int8, reflect.Uint8, reflect.Int, reflect.Uint:
		return fmt.Sprintf("%d", v.Value)
	case reflect.Float64:
		return fmt.Sprintf("%f", v.Value)
	case reflect.Bool:
		return strconv.FormatBool(v.Value.(bool))
	case reflect.String:
		return v.Value.(string)
	default:
		return ""
	}
}

func (v *GHWebhookFieldVal) GetAsMap() map[string]interface{} {
	return v.Value.(map[string]interface{})
}

func (v *GHWebhookFieldVal) IsMap() bool {
	return v.Type.Kind() == reflect.Map && v.Type.Key().Kind() == reflect.String
}

func (v *GHWebhookFieldVal) IsString() bool {
	return v.Type.Kind() == reflect.String
}

func (v *GHWebhookFieldVal) IsNumeric() bool {
	return v.Type.Kind() == reflect.Float64 ||
		v.Type.Kind() == reflect.Int64 || v.Type.Kind() == reflect.Uint64 ||
		v.Type.Kind() == reflect.Int || v.Type.Kind() == reflect.Uint ||
		v.Type.Kind() == reflect.Int32 || v.Type.Kind() == reflect.Uint32 ||
		v.Type.Kind() == reflect.Int16 || v.Type.Kind() == reflect.Uint16 ||
		v.Type.Kind() == reflect.Int8 || v.Type.Kind() == reflect.Uint8
}

func (v *GHWebhookFieldVal) IsBool() bool {
	return v.Type.Kind() == reflect.Bool
}

func (f *GHWebhookField) Matches(payload map[string]interface{}, ghEvent GHWebhookEvent, key string) error {
	curObj, err := jsonpath.Get(key, payload)

	if err != nil {
		return err
	}

	if curObj == nil {
		return fmt.Errorf("key %s not found in payload", key)
	}

	fieldVal := GHWebhookFieldVal{Value: curObj, Type: reflect.TypeOf(curObj)}

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

	if len(f.Expr) > 0 {
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

		switch reflect.TypeOf(output).Kind() {
		case reflect.Bool:
			if !output.(bool) {
				log.Infof("event[%d] failed to match expr", ghEvent.ID)
				return fmt.Errorf("event[%d] failed to match expr", ghEvent.ID)
			}
		default:
			log.Errorf("invalid return type %v for expr %s", reflect.TypeOf(output), f.Expr)
			return fmt.Errorf("invalid return type %v for expr %s", reflect.TypeOf(output), f.Expr)
		}

	}

	if len(f.Child) > 0 {
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

func (f *GHWebhookField) IsValid() error {
	if len(f.PositiveMatches) == 0 && len(f.NegativeMatches) == 0 && len(f.Child) == 0 && len(f.Expr) == 0 {
		return fmt.Errorf("no filter")
	}

	if len(f.Expr) > 0 {
		_, err := expr.Compile(f.Expr, expr.AsBool())
		if err != nil {
			log.Errorf("failed to compile expr %s: %v", f.Expr, err)
			return err
		}
	}
	return nil

}

type GHWebHookSubscribe struct {
	gorm.Model
	GHWebhookReceiverID uint
	GHWebhookReceiver   GHWebhookReceiver
	Event               string // mandatory

	Filters map[string]GHWebhookField `gorm:"serializer:json"`
}

func (s *GHWebHookSubscribe) Matches(payload map[string]interface{}, ghEvent GHWebhookEvent) error {
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
			return nil
		}
	}
	return fmt.Errorf("event[%d] %s doesn't match %s", ghEvent.ID, ghEvent.Event, s.Event)
}

func (s *GHWebHookSubscribe) IsValid() error {
	if len(s.Event) == 0 {
		return fmt.Errorf("event is required")
	}

	for k, v := range s.Filters {
		if err := v.IsValid(); err != nil {
			return fmt.Errorf("invalid filter %s: %v", k, err)
		}
	}
	return nil
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
