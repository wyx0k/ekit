package app

import (
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type FileConfig struct {
	SourceName string
	Filename   string
	Filepath   string
	v          *viper.Viper
	updater    *ConfigUpdater
}

func DefaultFileConfigLoader() *FileConfig {
	return NewFileConfigLoader("./config.yaml")
}
func NewFileConfigLoader(path string) *FileConfig {
	_, f := filepath.Split(path)
	fc := &FileConfig{
		Filename: f,
		Filepath: path,
	}
	return fc
}

func (f *FileConfig) Load(updater *ConfigUpdater) error {
	v := viper.New()
	v.SetConfigFile(f.Filepath)
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	v.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("[config] %s changed:%s", f.Filename, e.Name)
		c := Conf{}
		err := v.Unmarshal(&c)
		if err != nil {
			log.Error(err)
		}
		updater.UpdateConfig(&c)
	})
	v.WatchConfig()
	c := Conf{}
	err := v.Unmarshal(&c)
	if err != nil {
		log.Error(err)
	}
	f.v = v
	f.updater = updater
	updater.UpdateConfig(&c)
	return nil
}

func (f *FileConfig) Stop() error {
	return nil
}
