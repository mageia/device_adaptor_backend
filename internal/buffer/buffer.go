package buffer

import (
	"device_adaptor"
	"sync"
)

type Buffer struct {
	sync.Mutex
	buf   []device_agent.Metric
	first int
	last  int
	size  int
	empty bool
}

func NewBuffer(size int) *Buffer {
	return &Buffer{
		buf:   make([]device_agent.Metric, size),
		first: 0,
		last:  0,
		size:  size,
		empty: true,
	}
}

func (b *Buffer) IsEmpty() bool {
	return b.empty
}

func (b *Buffer) Len() int {
	if b.empty {
		return 0
	} else if b.first < b.last {
		return b.last - b.first + 1
	}
	return b.size - (b.first - b.last - 1)
}

func (b *Buffer) push(m device_agent.Metric) {
	if b.empty {
		b.last = b.first
		b.buf[b.last] = m
		b.empty = false
		return
	}

	b.last++
	b.last %= b.size

	if b.first == b.last {
		b.first = (b.first + 1) % b.size
	}
	b.buf[b.last] = m
}

func (b *Buffer) Add(metrics ...device_agent.Metric) {
	b.Lock()
	defer b.Unlock()
	for i := range metrics {
		b.push(metrics[i])
	}
}

func (b *Buffer) Batch(batchSize int) []device_agent.Metric {
	b.Lock()
	defer b.Unlock()
	outLen := min(b.Len(), batchSize)
	out := make([]device_agent.Metric, outLen)
	if outLen == 0 {
		return out
	}
	rightInd := min(b.size, b.first+outLen) - 1
	copyCount := copy(out, b.buf[b.first:rightInd+1])
	if rightInd == b.last {
		b.empty = true
	}
	b.first = rightInd + 1
	b.first %= b.size

	if copyCount < outLen {
		right := min(b.last, outLen-copyCount)
		copy(out[copyCount:], b.buf[b.first:right+1])
		if right == b.last {
			b.empty = true
		}
		b.first = right + 1
		b.first %= b.size
	}
	return out
}

func min(a, b int) int {
	if b < a {
		return b
	}
	return a
}
