package record

import (
	"errors"
	"fmt"
	"sync"

	"github.com/safing/portbase/container"
	"github.com/safing/portbase/database/accessor"
	"github.com/safing/portbase/formats/dsd"
	"github.com/safing/portbase/formats/varint"
)

// Wrapper wraps raw data and implements the Record interface.
type Wrapper struct {
	Base
	sync.Mutex

	Format uint8
	Data   []byte
}

// NewRawWrapper returns a record wrapper for the given data, including metadata. This is normally only used by storage backends when loading records.
func NewRawWrapper(database, key string, data []byte) (*Wrapper, error) {
	version, offset, err := varint.Unpack8(data)
	if err != nil {
		return nil, err
	}
	if version != 1 {
		return nil, fmt.Errorf("incompatible record version: %d", version)
	}

	metaSection, n, err := varint.GetNextBlock(data[offset:])
	if err != nil {
		return nil, fmt.Errorf("could not get meta section: %s", err)
	}
	offset += n

	newMeta := &Meta{}
	if len(metaSection) == 34 && metaSection[4] == 0 {
		// TODO: remove in 2020
		// backward compatibility:
		// format would byte shift and populate metaSection[4] with value > 0 (would naturally populate >0 at 07.02.2106 07:28:15)
		// this must be gencode without format
		_, err = newMeta.GenCodeUnmarshal(metaSection)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal meta section: %s", err)
		}
	} else {
		_, err = dsd.Load(metaSection, newMeta)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal meta section: %s", err)
		}
	}

	format, n, err := varint.Unpack8(data[offset:])
	if err != nil {
		return nil, fmt.Errorf("could not get dsd format: %s", err)
	}
	offset += n

	return &Wrapper{
		Base{
			database,
			key,
			newMeta,
		},
		sync.Mutex{},
		format,
		data[offset:],
	}, nil
}

// NewWrapper returns a new record wrapper for the given data.
func NewWrapper(key string, meta *Meta, format uint8, data []byte) (*Wrapper, error) {
	dbName, dbKey := ParseKey(key)

	return &Wrapper{
		Base{
			dbName: dbName,
			dbKey:  dbKey,
			meta:   meta,
		},
		sync.Mutex{},
		format,
		data,
	}, nil
}

// Marshal marshals the object, without the database key or metadata
func (w *Wrapper) Marshal(r Record, format uint8) ([]byte, error) {
	if w.Meta() == nil {
		return nil, errors.New("missing meta")
	}

	if w.Meta().Deleted > 0 {
		return nil, nil
	}

	if format != AUTO && format != w.Format {
		return nil, errors.New("could not dump model, wrapped object format mismatch")
	}

	data := make([]byte, len(w.Data)+1)
	data[0] = w.Format
	copy(data[1:], w.Data)

	return data, nil
}

// MarshalRecord packs the object, including metadata, into a byte array for saving in a database.
func (w *Wrapper) MarshalRecord(r Record) ([]byte, error) {
	// Duplication necessary, as the version from Base would call Base.Marshal instead of Wrapper.Marshal

	if w.Meta() == nil {
		return nil, errors.New("missing meta")
	}

	// version
	c := container.New([]byte{1})

	// meta
	metaSection, err := dsd.Dump(w.meta, GenCode)
	if err != nil {
		return nil, err
	}
	c.AppendAsBlock(metaSection)

	// data
	dataSection, err := w.Marshal(r, JSON)
	if err != nil {
		return nil, err
	}
	c.Append(dataSection)

	return c.CompileData(), nil
}

// IsWrapped returns whether the record is a Wrapper.
func (w *Wrapper) IsWrapped() bool {
	return true
}

// Unwrap unwraps data into a record.
func Unwrap(wrapped, new Record) error {
	wrapper, ok := wrapped.(*Wrapper)
	if !ok {
		return fmt.Errorf("cannot unwrap %T", wrapped)
	}

	_, err := dsd.LoadAsFormat(wrapper.Data, wrapper.Format, new)
	if err != nil {
		return fmt.Errorf("failed to unwrap %T: %s", new, err)
	}

	new.SetKey(wrapped.Key())
	new.SetMeta(wrapped.Meta())

	return nil
}

// GetAccessor returns an accessor for this record, if available.
func (w *Wrapper) GetAccessor(self Record) accessor.Accessor {
	if w.Format == JSON && len(w.Data) > 0 {
		return accessor.NewJSONBytesAccessor(&w.Data)
	}
	return nil
}
