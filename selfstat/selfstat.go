package selfstat

import (
	"hash/fnv"
	"sort"
	"sync"
)

var (
	registry *Registry
)

type Stat interface {
	Name() string
	FieldName() string
	Tags() map[string]string
	Key() uint64
	Incr(v int64)
	Set(v int64)
	Get() int64
}

func Register(measurement, field string, tags map[string]string) Stat {
	return registry.register(&stat{
		measurement: "internal_" + measurement,
		field:       field,
		tags:        tags,
	})
}

type Registry struct {
	stats map[uint64]map[string]Stat
	mu    sync.Mutex
}

func (r *Registry) register(s Stat) Stat {
	r.mu.Lock()
	defer r.mu.Unlock()
	if stats, ok := r.stats[s.Key()]; ok {
		if stat, ok := stats[s.FieldName()]; ok {
			return stat
		}
		r.stats[s.Key()][s.FieldName()] = s
		return s
	} else {
		r.stats[s.Key()] = map[string]Stat{s.FieldName(): s}
		return s
	}
}

func key(measurement string, tags map[string]string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(measurement))

	tmp := make([]string, len(tags))
	i := 0
	for k, v := range tags {
		tmp[i] = k + v
		i++
	}
	sort.Strings(tmp)
	for _, s := range tmp {
		h.Write([]byte(s))
	}
	return h.Sum64()
}

func init() {
	registry = &Registry{
		stats: make(map[uint64]map[string]Stat),
	}
}
