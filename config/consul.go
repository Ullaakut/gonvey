package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	consul "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
)

// Consul wraps a consul configuration source
type Consul struct {
	kv *consul.KV
}

// NewConsul creates a new config source from consul
func NewConsul() (*Consul, error) {
	consulConfig := consul.DefaultConfig()
	consulConfig.Address = "config:8500"

	consulClient, err := consul.NewClient(consulConfig)
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize consul connection")
	}

	csl := &Consul{
		kv: consulClient.KV(),
	}

	return csl, nil
}

// Watch binds a callback to watched configuration values
func (c *Consul) Watch(cfg interface{}, callback func(bool, error)) error {
	p := reflect.ValueOf(cfg)
	if p.Kind() != reflect.Ptr {
		return errors.New("need to have a pointer to struct")
	}

	// check that it points to a struct
	v := p.Elem()
	if v.Kind() != reflect.Struct {
		return errors.New("need to have a pointer to struct")
	}
	t := v.Type()

	// loop over all config struct fields
	n := v.NumField()
	for i := 0; i < n; i++ {
		info := t.Field(i)
		tag := info.Tag
		key, ok := tag.Lookup("config")
		if !ok {
			continue
		}

		if strings.Contains(key, ",watch") {
			requiresRestart := strings.Contains(key, ",restart")

			// Remove tags to get only key
			key = strings.Replace(key, ",watch", "", 1)
			key = strings.Replace(key, ",restart", "", 1)

			go func() {
				var queryOptions *consul.QueryOptions
				for {
					kv, _, err := c.kv.Get(key, queryOptions)

					if err != nil || kv == nil {
						if err == nil {
							err = errors.Errorf("consul key '%s' cannot be found", key)
						}
						callback(false, err)
						queryOptions = nil
						// connection lost with consul, sleep before retrying
						time.Sleep(5 * time.Second)
						continue
					}

					lastIndex := kv.ModifyIndex
					if queryOptions != nil && queryOptions.WaitIndex == lastIndex {
						// value didn't change, we just had a timeout
						continue
					}

					queryOptions = &consul.QueryOptions{
						WaitIndex: lastIndex,
					}
					c.Get(cfg)
					callback(requiresRestart, nil)
				}
			}()
		}
	}

	return nil
}

// Set writes the data from the config into consul
func (c *Consul) Set(cfg interface{}) error {
	// check that we have a pointer
	p := reflect.ValueOf(cfg)
	if p.Kind() != reflect.Ptr {
		return errors.New("need to have a pointer to struct")
	}

	// check that it points to a struct
	v := p.Elem()
	if v.Kind() != reflect.Struct {
		return errors.New("need to have a pointer to struct")
	}
	t := v.Type()

	// loop over all config struct fields
	n := v.NumField()
	for i := 0; i < n; i++ {
		info := t.Field(i)
		tag := info.Tag
		key, ok := tag.Lookup("config")
		if !ok {
			continue
		}

		key = strings.Replace(key, ",watch", "", 1)
		key = strings.Replace(key, ",restart", "", 1)

		fmt.Println("Inserting", fmt.Sprint(v.Field(i)), "for key", key)

		_, err := c.kv.Put(&consul.KVPair{
			Key:   key,
			Value: []byte(fmt.Sprint(v.Field(i))),
		}, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// Get reads the data fron consul and populates the config structure
func (c *Consul) Get(cfg interface{}) error {

	// check that we have a pointer
	p := reflect.ValueOf(cfg)
	if p.Kind() != reflect.Ptr {
		return errors.New("need to have a pointer to struct")
	}

	// check that it points to a struct
	v := p.Elem()
	if v.Kind() != reflect.Struct {
		return errors.New("need to have a pointer to struct")
	}
	t := v.Type()

	// loop over all config struct fields
	n := v.NumField()
	for i := 0; i < n; i++ {

		// get the annotation info
		info := t.Field(i)
		tag := info.Tag
		key, ok := tag.Lookup("config")
		if !ok {
			continue
		}

		key = strings.Replace(key, ",watch", "", 1)
		key = strings.Replace(key, ",restart", "", 1)

		// set the field to the desired value
		field := v.Field(i)
		switch field.Kind() {
		case reflect.Bool:
			strVal, err := c.getStringVal(key)
			if err != nil {
				return errors.Wrapf(err, "could not read string value for %v", key)
			}
			boolVal, err := strconv.ParseBool(strVal)
			if err != nil {
				return errors.Wrapf(err, "could not read %v as bool", key)
			}
			field.SetBool(boolVal)
		case reflect.Int:
			strVal, err := c.getStringVal(key)
			if err != nil {
				return errors.Wrapf(err, "could not read string value for %v", key)
			}
			intVal, err := strconv.ParseInt(strVal, 10, 64)
			if err != nil {
				return errors.Wrapf(err, "could not read %v as int", key)
			}
			field.SetInt(intVal)
		case reflect.String:
			strVal, err := c.getStringVal(key)
			if err != nil {
				return errors.Wrapf(err, "could not read string value for %v", key)
			}
			field.SetString(strVal)
		case reflect.Map:
			mapVal, err := c.getMapVal(key)
			if err != nil {
				return errors.Wrapf(err, "could not read map value for %v", key)
			}
			mapField := reflect.MakeMap(field.Type())
			field.Set(mapField)
			for key, value := range mapVal {
				keyVal := reflect.ValueOf(key)
				valVal := reflect.ValueOf(value)
				field.SetMapIndex(keyVal, valVal)
			}
		case reflect.Int64:
			strVal, err := c.getStringVal(key)
			if err != nil {
				return errors.Wrapf(err, "could not read string value for %v", key)
			}
			duration, err := time.ParseDuration(strVal)
			if err != nil {
				return errors.Wrapf(err, "could not read %v as duration", key)
			}
			field.Set(reflect.ValueOf(duration))
		default:
			return errors.Errorf("invalid field type for key %v: %v", key, field)
		}
	}

	return nil
}

func (c *Consul) getStringVal(key string) (string, error) {
	pair, _, err := c.kv.Get(key, nil)
	if err != nil {
		return "", errors.Wrap(err, "could not get key")
	}
	if pair == nil {
		return "", errors.New("key not found")
	}
	value := string(pair.Value)
	return value, nil
}

func (c *Consul) getMapVal(key string) (map[string]string, error) {
	list, _, err := c.kv.List(key, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not list key")
	}
	values := make(map[string]string, len(list))
	for _, pair := range list {
		values[pair.Key] = string(pair.Value)
	}
	return values, nil
}
