package selfstat

import (
	"sync/atomic"
)

type stat struct {
	v           int64
	measurement string
	field       string
	tags        map[string]string
	key         uint64
}

func (s *stat) Incr(v int64) {
	atomic.AddInt64(&s.v, v)
}

func (s *stat) Set(v int64) {
	atomic.StoreInt64(&s.v, v)
}
func (s *stat) Get() int64 {
	return atomic.LoadInt64(&s.v)
}
func (s *stat) Name() string {
	return s.measurement
}
func (s *stat) FieldName() string {
	return s.field
}
func (s *stat) Tags() map[string]string {
	m := make(map[string]string, len(s.tags))
	for k, v := range s.tags {
		m[k] = v
	}
	return m
}
func (s *stat) Key() uint64 {
	if s.key == 0 {
		s.key = key(s.measurement, s.tags)
	}
	return s.key
}
