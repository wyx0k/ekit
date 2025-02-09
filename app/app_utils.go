package app

import (
	"errors"
)

var ErrComponentNotFound = errors.New("the component not found")
var ErrComponentTypeMissMatch = errors.New("found component but type do not match")
var ErrComponentMetaNotFound = errors.New("the component meta info not found")

func GetComponentById[T Component](app *AppContext, id string) (T, error) {
	var t T
	c := app.GetComponentById(id)
	if c == nil {
		return t, ErrComponentNotFound
	}
	if tc, ok := c.(T); ok {
		return tc, nil
	} else {
		return t, ErrComponentTypeMissMatch
	}
}

func GetSingletonComponentByType[T Component](app *AppContext, componentType ComponentType) (T, error) {
	var t T
	c := app.GetSingletonComponent(string(componentType))
	if c == nil {
		return t, ErrComponentNotFound
	}
	if tc, ok := c.(T); ok {
		return tc, nil
	} else {
		return t, ErrComponentTypeMissMatch
	}
}

func GetComponentId(app *AppContext, c Component) (string, error) {
	meta := app.Meta(c)
	if meta == nil {
		return "", ErrComponentMetaNotFound
	}
	return meta.ID(), nil
}

func GetComponentMetaInfo(app *AppContext, c Component) (*ComponentMeta[Component], error) {
	meta := app.Meta(c)
	if meta == nil {
		return nil, ErrComponentMetaNotFound
	}
	return meta, nil
}
func GetComponentMetaInfoById(app *AppContext, id string) (*ComponentMeta[Component], error) {
	c, err := GetComponentById[Component](app, id)
	if err != nil {
		return nil, err
	}
	meta := app.Meta(c)
	if meta == nil {
		return nil, ErrComponentMetaNotFound
	}
	return meta, nil
}
