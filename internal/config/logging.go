package config

import "github.com/sirupsen/logrus"

// PrefixedTextFormatter defines a text formatter for logging
type PrefixedTextFormatter struct {
	Prefix string
}

// Format formats the log entry
func (f *PrefixedTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	formatter := &logrus.TextFormatter{
		DisableTimestamp:       true,
		PadLevelText:           false,
		DisableLevelTruncation: true,
	}
	b, err := formatter.Format(entry)
	if err == nil {
		if len(b) > 0 {
			prefix := []byte(f.Prefix)
			buf := append(prefix[:], b[:]...)
			return buf, nil
		}
		return b, nil
	}
	return nil, err
}
