package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/Safing/portbase/log"
)

var (
	configFilePath string
)

func loadConfig() error {
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	m, err := JSONToMap(data)
	if err != nil {
		return err
	}

	return setConfig(m)
}

func saveConfig() (err error) {
	data, err := MapToJSON(userConfig)
	if err == nil {
		err = ioutil.WriteFile(configFilePath, data, 0600)
	}

	if err != nil {
		log.Errorf("config: failed to save config: %s", err)
	}

	return err
}

// JSONToMap parses and flattens a hierarchical json object.
func JSONToMap(jsonData []byte) (map[string]interface{}, error) {
	loaded := make(map[string]interface{})
	err := json.Unmarshal(jsonData, &loaded)
	if err != nil {
		return nil, err
	}

	flatten(loaded, loaded, "")
	return loaded, nil
}

func flatten(rootMap, subMap map[string]interface{}, subKey string) {
	for key, entry := range subMap {

		// get next level key
		subbedKey := key
		if subKey != "" {
			subbedKey = fmt.Sprintf("%s/%s", subKey, key)
		}

		// check for next subMap
		nextSub, ok := entry.(map[string]interface{})
		if ok {
			flatten(rootMap, nextSub, subbedKey)
			delete(rootMap, key)
		} else if subKey != "" {
			// only set if not on root level
			rootMap[subbedKey] = entry
		}
	}
}

// MapToJSON expands a flattened map and returns it as json.
func MapToJSON(mapData map[string]interface{}) ([]byte, error) {
	configLock.RLock()
	defer configLock.RUnlock()

	new := make(map[string]interface{})
	for key, value := range mapData {
		new[key] = value
	}
	
	expand(new)
	return json.MarshalIndent(new, "", "  ")
}

// expand expands a flattened map.
func expand(mapData map[string]interface{}) {
	var newMaps []map[string]interface{}
	for key, entry := range mapData {
		if strings.Contains(key, "/") {
			parts := strings.SplitN(key, "/", 2)
			if len(parts) == 2 {

				// get subMap
				var subMap map[string]interface{}
				v, ok := mapData[parts[0]]
				if ok {
					subMap, ok = v.(map[string]interface{})
					if !ok {
						subMap = make(map[string]interface{})
						newMaps = append(newMaps, subMap)
						mapData[parts[0]] = subMap
					}
				} else {
					subMap = make(map[string]interface{})
					newMaps = append(newMaps, subMap)
					mapData[parts[0]] = subMap
				}

				// set entry
				subMap[parts[1]] = entry
				// delete entry from
				delete(mapData, key)

			}
		}
	}
	for _, entry := range newMaps {
		expand(entry)
	}
}