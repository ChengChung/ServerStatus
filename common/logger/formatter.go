package logger

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/chengchung/ServerStatus/common/go_utils"
	"github.com/sirupsen/logrus"
)

const (
	CustomTimeFormat = "0102 15:04:05.000000"

	FieldGOID    = "goid"
	FieldPID     = "pid"
	FieldKeyMsg  = logrus.FieldKeyMsg
	FieldKeyFunc = logrus.FieldKeyFunc
	FieldKeyFile = logrus.FieldKeyFile
)

type CustomFormatter struct {
	// Force quoting of all values
	ForceQuote bool

	// DisableQuote disables quoting for all values.
	// DisableQuote will have a lower priority than ForceQuote.
	// If both of them are set to true, quote will be forced on all values.
	DisableQuote bool

	// QuoteEmptyFields will wrap empty fields in quotes if true
	QuoteEmptyFields bool

	DisableSorting bool
	SortingFunc    func([]string)

	EnableGoRoutineId bool
}

func pid() int {
	return os.Getpid()
}

func MarshalLevel(level logrus.Level) []byte {
	switch level {
	case logrus.TraceLevel:
		return []byte("T")
	case logrus.DebugLevel:
		return []byte("D")
	case logrus.InfoLevel:
		return []byte("I")
	case logrus.WarnLevel:
		return []byte("W")
	case logrus.ErrorLevel:
		return []byte("E")
	case logrus.FatalLevel:
		return []byte("F")
	case logrus.PanicLevel:
		return []byte("P")
	}

	return []byte("UNKNOWN")
}

// Format renders a single log entry
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields)
	for k, v := range entry.Data {
		data[k] = v
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	if !f.DisableSorting {
		if f.SortingFunc == nil {
			sort.Strings(keys)
		} else {
			f.SortingFunc(keys)
		}
	}

	//  {LEVEL}{DATE} {TIME} {PID} {GOROUTINEID} {CODELINE} [{Funct}] LOGINFO OTHERKVMAP
	fixedKeys := make([]string, 0, 3)
	fixedKVMap := make(map[string]interface{}, 5)

	fixedKeys = append(fixedKeys, FieldPID)
	fixedKVMap[FieldPID] = pid()

	if f.EnableGoRoutineId {
		fixedKeys = append(fixedKeys, FieldGOID)
		fixedKVMap[FieldGOID] = go_utils.Goid()
	}

	var funcVal, fileVal string
	if entry.HasCaller() {
		funcVal = fmt.Sprintf("[%s]", filepath.Ext(entry.Caller.Function)[1:])
		fileVal = fmt.Sprintf("%s:%d", filepath.Base(entry.Caller.File), entry.Caller.Line)

		if fileVal != "" {
			fixedKVMap[FieldKeyFile] = fileVal
			fixedKeys = append(fixedKeys, FieldKeyFile)
		}
		if funcVal != "" {
			fixedKVMap[FieldKeyFunc] = funcVal
			fixedKeys = append(fixedKeys, FieldKeyFunc)
		}
	}

	if entry.Message != "" {
		fixedKeys = append(fixedKeys, FieldKeyMsg)
		fixedKVMap[FieldKeyMsg] = entry.Message
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	f.appendFieldValue(b, fmt.Sprintf("%s%s", MarshalLevel(entry.Level), entry.Time.Format(CustomTimeFormat)))

	for _, key := range fixedKeys {
		f.appendFieldValue(b, fixedKVMap[key])
	}

	for _, key := range keys {
		f.appendKeyValue(b, key, data[key])
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *CustomFormatter) appendFieldValue(b *bytes.Buffer, value interface{}) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}
	b.WriteString(stringVal)
}

func (f *CustomFormatter) needsQuoting(text string) bool {
	if f.ForceQuote {
		return true
	}
	if f.QuoteEmptyFields && len(text) == 0 {
		return true
	}
	if f.DisableQuote {
		return false
	}
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.' || ch == '_' || ch == '/' || ch == '@' || ch == '^' || ch == '+') {
			return true
		}
	}
	return false
}

func (f *CustomFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	b.WriteByte('=')
	f.appendValue(b, value)
}

func (f *CustomFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}

	if !f.needsQuoting(stringVal) {
		b.WriteString(stringVal)
	} else {
		b.WriteString(fmt.Sprintf("%q", stringVal))
	}
}
