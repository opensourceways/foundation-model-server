package dp

import "errors"

var modelNameValidator ModelNameValidator

type ModelNameValidator interface {
	IsValidModelName(string) bool
}

type ModelName interface {
	ModelName() string
}

func NewModelName(v string) (ModelName, error) {
	if v == "" || !modelNameValidator.IsValidModelName(v) {
		return nil, errors.New("invalid modelName")
	}

	return nil, nil
}

type modelName string

func (v modelName) ModelName() string {
	return string(v)
}
