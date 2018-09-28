package inputs

import (
	"deviceAdaptor"
	"github.com/stretchr/testify/mock"
)

type MockPlugin struct {
	mock.Mock
}

func (m *MockPlugin) Gather(_a0 deviceAgent.Accumulator) error {
	ret := m.Called(_a0)
	return ret.Error(0)
}