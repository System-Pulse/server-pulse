package test

import (
	"testing"
)

func TestUpdateCPUInfo(t *testing.T) {
	t.Parallel()

	t.Run("Skip CPU tests - requires system access", func(t *testing.T) {
		t.Skip("CPU tests require system access - skipping integration tests")
	})
}

func TestUpdateMemoryInfo(t *testing.T) {
	t.Parallel()

	t.Run("Skip memory tests - requires system access", func(t *testing.T) {
		t.Skip("Memory tests require system access - skipping integration tests")
	})
}

func TestUpdateDiskInfo(t *testing.T) {
	t.Parallel()

	t.Run("Skip disk tests - requires system access", func(t *testing.T) {
		t.Skip("Disk tests require system access - skipping integration tests")
	})
}

func TestUpdateNetworkInfo(t *testing.T) {
	t.Parallel()

	t.Run("Skip network tests - requires system access", func(t *testing.T) {
		t.Skip("Network tests require system access - skipping integration tests")
	})
}

func TestLoadAvg(t *testing.T) {
	t.Parallel()

	t.Run("Skip load average tests - requires system access", func(t *testing.T) {
		t.Skip("Load average tests require system access - skipping integration tests")
	})
}
