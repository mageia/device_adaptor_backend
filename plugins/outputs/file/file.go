package file

import (
	"deviceAdaptor"
	"deviceAdaptor/plugins/outputs"
	"deviceAdaptor/plugins/serializers"
	"fmt"
	"io"
	"os"
)

type File struct {
	Files      []string
	writers    []io.Writer
	closers    []io.Closer
	serializer serializers.Serializer
}

func (f *File) SetSerializer(serializer serializers.Serializer) {
	f.serializer = serializer
}

func (f *File) Connect() error {
	if len(f.Files) == 0 {
		f.Files = []string{"stdout"}
	}
	for _, file := range f.Files {
		if file == "stdout" {
			f.writers = append(f.writers, os.Stdout)
		} else {
			var of *os.File
			var err error
			if _, err := os.Stat(file); os.IsNotExist(err) {
				of, err = os.Create(file)
			} else {
				of, err = os.OpenFile(file, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			}
			if err != nil {
				return nil
			}
			f.writers = append(f.writers, of)
			f.closers = append(f.closers, of)
		}
	}
	return nil
}

func (f *File) Close() error {
	var errS string
	for _, c := range f.closers {
		if err := c.Close(); err != nil {
			errS += err.Error() + "\n"
		}
	}
	if errS != "" {
		return fmt.Errorf(errS)
	}
	return nil
}

func (f *File) SampleConfig() string {
	return "SampleConfig of output.file"
}

func (f *File) Description() string {
	return "Description of output.file"
}
func (f *File) Write(metrics []deviceAgent.Metric) error {
	var writeErr error = nil
	for _, metric := range metrics {
		b, err := f.serializer.Serialize(metric)
		if err != nil {
			return fmt.Errorf("failed to serialize message: %s", err)
		}
		for _, writer := range f.writers {
			_, err = writer.Write(b)
			if err != nil && writer != os.Stdout {
				writeErr = fmt.Errorf("failed to write message: %s, %s", b, err)
			}
		}
	}
	return writeErr
}

func init() {
	outputs.Add("file", func() deviceAgent.Output {
		return &File{}
	})
}
