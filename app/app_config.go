package app

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"dario.cat/mergo"
	"github.com/spf13/cast"
)

// source -> conf
type ConfigLoader interface {
	Load(updater *ConfigUpdater) error
}
type CloseableConfigLoader interface {
	ConfigLoader
	Stop() error
}

type ConfigUpdater struct {
	lock       sync.RWMutex
	target     *Conf
	stopUpdate bool
}

func (u *ConfigUpdater) stop() {
	u.stopUpdate = true
}

func (u *ConfigUpdater) UpdateConfig(confs ...*Conf) error {
	if u.stopUpdate {
		return nil
	}
	u.lock.Lock()
	defer u.lock.Unlock()
	u.target.store(confs...)
	return nil
}

type ConfContext struct {
	loaders          []ConfigLoader
	mustCloseLoaders []CloseableConfigLoader
	config           *Conf
	configUpdater    *ConfigUpdater
}

func NewConfContext(loaders ...ConfigLoader) *ConfContext {
	config := &Conf{}
	configUpdater := &ConfigUpdater{target: config}
	return &ConfContext{
		loaders:       loaders,
		config:        config,
		configUpdater: configUpdater,
	}
}

func (c *ConfContext) Close() error {
	if c != nil && c.configUpdater != nil {
		c.configUpdater.stop()
		var errs []error
		for _, l := range c.mustCloseLoaders {
			err := l.Stop()
			errs = append(errs, err)
		}
		return errors.Join(errs...)
	}
	return nil
}

func (c *ConfContext) initConf() error {
	for _, loader := range c.loaders {
		err := loader.Load(c.configUpdater)
		if err != nil {
			return err
		}
		if cl, ok := loader.(CloseableConfigLoader); ok {
			c.mustCloseLoaders = append(c.mustCloseLoaders, cl)

		}
	}
	return nil
}

func (c *ConfContext) Value(key string) ConfValue {
	c.configUpdater.lock.RLock()
	defer c.configUpdater.lock.RUnlock()
	return c.config.Value(key)
}

type Conf map[string]any

func (c *Conf) Value(fullKey string) ConfValue {
	v, found := getValue(*c, fullKey)
	if !found {
		return ConfValue{notfound: true}
	}
	return v
}

func getValue(target map[string]any, key string) (ConfValue, bool) {
	ks := strings.Split(key, ".")
	lastKey := len(ks) - 1
	nextLevel := target
	for idx, k := range ks {
		if v, ok := nextLevel[k]; ok {
			if idx == lastKey {
				cv := ConfValue{
					value: v,
				}
				return cv, true
			} else {
				if tmp, ok := v.(map[string]any); ok {
					nextLevel = tmp
					continue
				} else {
					return ConfValue{}, false
				}
			}
		} else {
			return ConfValue{}, false
		}
	}
	return ConfValue{}, false
}

func (c *Conf) store(data ...*Conf) error {
	n := Conf{}
	mergo.Map(&n, *c, mergo.WithOverride)
	for _, d := range data {
		mergo.Map(&n, d, mergo.WithOverride)
	}
	*c = n
	return nil
}

var ErrConfigNotFound = errors.New("config key not found")

type ConfValue struct {
	notfound bool
	value    any
}

func (c ConfValue) Bool() bool {
	b, err := c.MustBool()
	if err != nil {
		return false
	}
	return b
}

func (c ConfValue) Int() int {
	b, err := c.MustInt()
	if err != nil {
		return 0
	}
	return b
}

func (c ConfValue) Int64() int64 {
	b, err := c.MustInt64()
	if err != nil {
		return 0
	}
	return b
}

func (c ConfValue) Uint64() uint64 {
	b, err := c.MustUint64()
	if err != nil {
		return 0
	}
	return b
}

func (c ConfValue) Float64() float64 {
	b, err := c.MustFloat64()
	if err != nil {
		return 0.0
	}
	return b
}

func (c ConfValue) String() string {
	s, err := c.MustString()
	if err != nil {
		return ""
	}
	return s
}

func (c ConfValue) Duration() time.Duration {
	d, err := c.MustDuration()
	if err != nil {
		return 0
	}
	return d
}

func (c ConfValue) Slice() []ConfValue {
	s, err := c.MustSlice()
	if err != nil {
		return nil
	}
	return s
}

func (c ConfValue) Map() map[string]ConfValue {
	s, err := c.MustMap()
	if err != nil {
		return nil
	}
	return s
}

func (c ConfValue) Scan(data any) error {
	j, err := json.Marshal(c.value)
	if err != nil {
		return err
	}
	return json.Unmarshal(j, data)
}

func (c ConfValue) MustBool() (bool, error) {
	if c.notfound {
		return false, ErrConfigNotFound
	}
	return cast.ToBoolE(c.value)
}
func (c ConfValue) MustInt() (int, error) {
	if c.notfound {
		return 0, ErrConfigNotFound
	}
	return cast.ToIntE(c.value)
}
func (c ConfValue) MustInt64() (int64, error) {
	if c.notfound {
		return 0, ErrConfigNotFound
	}
	return cast.ToInt64E(c.value)
}
func (c ConfValue) MustUint64() (uint64, error) {
	if c.notfound {
		return 0, ErrConfigNotFound
	}
	return cast.ToUint64E(c.value)
}
func (c ConfValue) MustFloat64() (float64, error) {
	if c.notfound {
		return 0.0, ErrConfigNotFound
	}
	return cast.ToFloat64E(c.value)
}

func (c ConfValue) MustString() (string, error) {
	if c.notfound {
		return "", ErrConfigNotFound
	}
	return cast.ToStringE(c.value)
}

func (c ConfValue) MustDuration() (time.Duration, error) {
	if c.notfound {
		return 0, ErrConfigNotFound
	}
	return cast.ToDurationE(c.value)
}

func (c ConfValue) MustSlice() ([]ConfValue, error) {
	var cvs []ConfValue
	if c.notfound {
		return cvs, ErrConfigNotFound
	}
	lst, err := cast.ToSliceE(c.value)
	if err != nil {
		return cvs, err
	}

	for _, e := range lst {
		cvs = append(cvs, ConfValue{value: e})
	}
	return cvs, nil
}

func (c ConfValue) MustMap() (map[string]ConfValue, error) {
	cvs := map[string]ConfValue{}
	if c.notfound {
		return cvs, ErrConfigNotFound
	}
	m, err := cast.ToStringMapE(c.value)
	if err != nil {
		return cvs, err
	}
	for k, v := range m {
		cvs[k] = ConfValue{value: v}
	}
	return cvs, nil
}
