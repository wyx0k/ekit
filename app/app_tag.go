package app

import (
	"errors"
	"strings"
)

const (
	TagEkit          = "ekit"
	TagEkitSep       = ";"
	TagEkitKVSep     = ":"
	TagEkitValuesSep = ","

	TagComponent = "component"
	TagRequired  = "required"
)

type EkitTagStr string

func (e EkitTagStr) Parse() (EkitTags, error) {
	tags := strings.Split(string(e), TagEkitSep)
	var result EkitTags
	duplicate := map[string]struct{}{}
	for _, tag := range tags {
		et := EkitTag{}
		idx := strings.Index(tag, TagEkitKVSep)
		if idx == -1 {
			et.Key = tag
		} else {
			et.Key = tag[0:idx]
			values := strings.Split(tag[idx+1:], TagEkitValuesSep)
			et.Values = values
		}
		if _, ok := duplicate[et.Key]; ok {
			return result, errors.New("标签重复：" + et.Key)
		}
		result = append(result, et)
		duplicate[et.Key] = struct{}{}
	}
	return result, nil
}

type EkitTags []EkitTag

func (e EkitTags) TagCount() int {
	return len(e)
}

type EkitTag struct {
	Key    string
	Values []string
}

func (e EkitTag) ValueCount() int {
	return len(e.Values)
}

func FindTag(tags EkitTags, target string) (EkitTag, bool) {
	for _, tag := range tags {
		if tag.Key == target {
			return tag, true
		}
	}
	return EkitTag{}, false
}
