package gen

import (
	"bytes"
	"errors"
	"time"
)

type metadata struct {
	Title string
	Date  time.Time
	Href  string
}

var ErrInvalidMetadata = errors.New("invalid metadata")

const datefmt = "2006-01-02"

func (m *metadata) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	var (
		buf bytes.Buffer
		key string
	)

	set := func(k, v string) error {
		if v == "" {
			return ErrInvalidMetadata
		}

		switch key {
		case "date":
			d, err := time.Parse(datefmt, buf.String())
			if err != nil {
				return ErrInvalidMetadata
			}
			m.Date = d
		case "title":
			m.Title = buf.String()
		default:
			return ErrInvalidMetadata
		}

		return nil
	}

	for i := 0; i < len(b); i++ {
		switch b[i] {
		case ' ':
			// skip start spaces
			if buf.Len() == 0 {
				continue
			}
			buf.WriteByte(b[i])
		case '\n':
			if err := set(key, buf.String()); err != nil {
				return err
			}
			key = ""
			buf.Reset()
		case ':':
			if key != "" {
				// char : is valid in a metadata value
				buf.WriteByte(b[i])
				continue
			}
			key = buf.String()
			buf.Reset()
		default:
			buf.WriteByte(b[i])
		}
	}

	if buf.Len() > 0 || key != "" {
		return set(key, buf.String())
	}

	return nil
}

func (m metadata) MarshalText() (text []byte, err error) {
	var buf bytes.Buffer
	if m.Title != "" {
		buf.WriteString("title: ")
		buf.WriteString(m.Title)
	}
	if !m.Date.IsZero() {
		if buf.Len() > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString("date: ")
		buf.WriteString(m.Date.Format(datefmt))
	}

	return buf.Bytes(), nil
}

func (m metadata) String() string {
	b, err := m.MarshalText()
	if err != nil {
		panic(err)
	}

	return string(b)
}

func (m metadata) IsPost() bool {
	return !m.Date.IsZero()
}
