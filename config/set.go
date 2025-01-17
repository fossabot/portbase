package config

import (
	"errors"
	"fmt"
	"sync"

	"github.com/tevino/abool"
)

var (
	// ErrInvalidJSON is returned by SetConfig and SetDefaultConfig if they receive invalid json.
	ErrInvalidJSON = errors.New("json string invalid")

	// ErrInvalidOptionType is returned by SetConfigOption and SetDefaultConfigOption if given an unsupported option type.
	ErrInvalidOptionType = errors.New("invalid option value type")

	validityFlag     = abool.NewBool(true)
	validityFlagLock sync.RWMutex

	changedSignal     = make(chan struct{})
	changedSignalLock sync.Mutex
)

func getValidityFlag() *abool.AtomicBool {
	validityFlagLock.RLock()
	defer validityFlagLock.RUnlock()
	return validityFlag
}

// Changed signals if any config option was changed.
func Changed() <-chan struct{} {
	changedSignalLock.Lock()
	defer changedSignalLock.Unlock()
	return changedSignal
}

func signalChanges() {
	// refetch and save release level and expertise level
	updateReleaseLevel()
	updateExpertiseLevel()

	// reset validity flag
	validityFlagLock.Lock()
	validityFlag.SetTo(false)
	validityFlag = abool.NewBool(true)
	validityFlagLock.Unlock()

	// trigger change signal: signal listeners that a config option was changed.
	changedSignalLock.Lock()
	close(changedSignal)
	changedSignal = make(chan struct{})
	changedSignalLock.Unlock()
}

// setConfig sets the (prioritized) user defined config.
func setConfig(newValues map[string]interface{}) error {
	optionsLock.Lock()
	for key, option := range options {
		newValue, ok := newValues[key]
		option.Lock()
		if ok {
			option.activeValue = newValue
		} else {
			option.activeValue = nil
		}
		option.Unlock()
	}
	optionsLock.Unlock()

	signalChanges()
	go pushFullUpdate()
	return nil
}

// SetDefaultConfig sets the (fallback) default config.
func SetDefaultConfig(newValues map[string]interface{}) error {
	optionsLock.Lock()
	for key, option := range options {
		newValue, ok := newValues[key]
		option.Lock()
		if ok {
			option.activeDefaultValue = newValue
		} else {
			option.activeDefaultValue = nil
		}
		option.Unlock()
	}
	optionsLock.Unlock()

	signalChanges()
	go pushFullUpdate()
	return nil
}

func validateValue(option *Option, value interface{}) error {
	switch v := value.(type) {
	case string:
		if option.OptType != OptTypeString {
			return fmt.Errorf("expected type %s for option %s, got type %T", getTypeName(option.OptType), option.Key, v)
		}
		if option.compiledRegex != nil {
			if !option.compiledRegex.MatchString(v) {
				return fmt.Errorf("validation failed: string \"%s\" did not match regex for option %s", v, option.Key)
			}
		}
		return nil
	case []string:
		if option.OptType != OptTypeStringArray {
			return fmt.Errorf("expected type %s for option %s, got type %T", getTypeName(option.OptType), option.Key, v)
		}
		if option.compiledRegex != nil {
			for pos, entry := range v {
				if !option.compiledRegex.MatchString(entry) {
					return fmt.Errorf("validation failed: string \"%s\" at index %d did not match regex for option %s", entry, pos, option.Key)
				}
			}
		}
		return nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		if option.OptType != OptTypeInt {
			return fmt.Errorf("expected type %s for option %s, got type %T", getTypeName(option.OptType), option.Key, v)
		}
		if option.compiledRegex != nil {
			if !option.compiledRegex.MatchString(fmt.Sprintf("%d", v)) {
				return fmt.Errorf("validation failed: number \"%d\" did not match regex for option %s", v, option.Key)
			}
		}
		return nil
	case bool:
		if option.OptType != OptTypeBool {
			return fmt.Errorf("expected type %s for option %s, got type %T", getTypeName(option.OptType), option.Key, v)
		}
		return nil
	default:
		return fmt.Errorf("invalid option value type: %T", value)
	}
}

// SetConfigOption sets a single value in the (prioritized) user defined config.
func SetConfigOption(key string, value interface{}) error {
	return setConfigOption(key, value, true)
}

func setConfigOption(key string, value interface{}, push bool) (err error) {
	optionsLock.Lock()
	option, ok := options[key]
	optionsLock.Unlock()
	if !ok {
		return fmt.Errorf("config option %s does not exist", key)
	}

	option.Lock()
	if value == nil {
		option.activeValue = nil
	} else {
		err = validateValue(option, value)
		if err == nil {
			option.activeValue = value
		}
	}
	option.Unlock()
	if err != nil {
		return err
	}

	// finalize change, activate triggers
	signalChanges()
	if push {
		go pushUpdate(option)
	}
	return saveConfig()
}

// SetDefaultConfigOption sets a single value in the (fallback) default config.
func SetDefaultConfigOption(key string, value interface{}) error {
	return setDefaultConfigOption(key, value, true)
}

func setDefaultConfigOption(key string, value interface{}, push bool) (err error) {
	optionsLock.Lock()
	option, ok := options[key]
	optionsLock.Unlock()
	if !ok {
		return fmt.Errorf("config option %s does not exist", key)
	}

	option.Lock()
	if value == nil {
		option.activeDefaultValue = nil
	} else {
		err = validateValue(option, value)
		if err == nil {
			option.activeDefaultValue = value
		}
	}
	option.Unlock()
	if err != nil {
		return err
	}

	// finalize change, activate triggers
	signalChanges()
	if push {
		go pushUpdate(option)
	}
	return saveConfig()
}
