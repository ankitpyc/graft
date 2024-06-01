package wal

type Options func(*Manager)

func WithLogReplication() Options {
	return func(m *Manager) {
		m.LogReplicator.WALManager = m
	}
}
