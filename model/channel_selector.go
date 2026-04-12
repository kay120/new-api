package model

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// ChannelMetrics stores performance metrics for a channel
type ChannelMetrics struct {
	ChannelId    int
	SuccessCount int64
	FailureCount int64
	TotalLatency int64 // in milliseconds
	AvgLatency   float64
	SuccessRate  float64
	LastUsed     time.Time
	Score        float64
	mu           sync.RWMutex
}

// ChannelSelector provides intelligent channel selection based on performance metrics
type ChannelSelector struct {
	metrics     map[int]*ChannelMetrics
	metricsLock sync.RWMutex
	ttl         time.Duration
}

var (
	selector     *ChannelSelector
	selectorOnce sync.Once
)

// GetChannelSelector returns the singleton ChannelSelector instance
func GetChannelSelector() *ChannelSelector {
	selectorOnce.Do(func() {
		selector = &ChannelSelector{
			metrics: make(map[int]*ChannelMetrics),
			ttl:     5 * time.Minute, // Metrics TTL
		}
		// Start cleanup goroutine
		go selector.cleanupLoop()
	})
	return selector
}

// RecordResult records the result of a channel usage
func (cs *ChannelSelector) RecordResult(channelId int, success bool, latencyMs int) {
	cs.metricsLock.Lock()
	defer cs.metricsLock.Unlock()

	m, exists := cs.metrics[channelId]
	if !exists {
		m = &ChannelMetrics{ChannelId: channelId}
		cs.metrics[channelId] = m
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if success {
		m.SuccessCount++
		m.TotalLatency += int64(latencyMs)
	} else {
		m.FailureCount++
	}
	m.LastUsed = time.Now()

	// Update calculated metrics
	cs.updateMetrics(m)
}

// updateMetrics recalculates derived metrics
func (cs *ChannelSelector) updateMetrics(m *ChannelMetrics) {
	total := m.SuccessCount + m.FailureCount
	if total > 0 {
		m.SuccessRate = float64(m.SuccessCount) / float64(total)
	}
	if m.SuccessCount > 0 {
		m.AvgLatency = float64(m.TotalLatency) / float64(m.SuccessCount)
	}

	// Calculate score (higher is better)
	// Score formula: success_rate * 100 - latency_penalty - failure_penalty
	latencyPenalty := math.Min(m.AvgLatency/100.0, 50.0) // Cap at 50
	failurePenalty := float64(m.FailureCount) * 10.0
	m.Score = m.SuccessRate*100 - latencyPenalty - failurePenalty

	// Ensure minimum score
	if m.Score < 1 {
		m.Score = 1
	}
}

// GetChannelScore returns the current score for a channel
func (cs *ChannelSelector) GetChannelScore(channelId int) float64 {
	cs.metricsLock.RLock()
	defer cs.metricsLock.RUnlock()

	if m, exists := cs.metrics[channelId]; exists {
		m.mu.RLock()
		defer m.mu.RUnlock()
		return m.Score
	}
	return 50.0 // Default score for new channels
}

// SelectChannel selects the best channel from available options using weighted random selection
// based on channel scores (dynamic performance-based weights)
func (cs *ChannelSelector) SelectChannel(channels []*Channel, useDynamicWeight bool) *Channel {
	if len(channels) == 0 {
		return nil
	}
	if len(channels) == 1 {
		return channels[0]
	}

	if !useDynamicWeight {
		// Fall back to original weight-based selection
		return cs.selectByStaticWeight(channels)
	}

	// Calculate dynamic weights based on scores
	type channelWeight struct {
		channel *Channel
		weight  float64
	}

	var weightedChannels []channelWeight
	var totalWeight float64

	for _, ch := range channels {
		score := cs.GetChannelScore(ch.Id)
		// Combine static weight with dynamic score
		staticWeight := float64(ch.GetWeight())
		if staticWeight == 0 {
			staticWeight = 100 // Default weight
		}

		// Dynamic weight = static_weight * (score / 50)
		// This means channels with score > 50 get boosted, < 50 get penalized
		dynamicWeight := staticWeight * (score / 50.0)

		weightedChannels = append(weightedChannels, channelWeight{
			channel: ch,
			weight:  dynamicWeight,
		})
		totalWeight += dynamicWeight
	}

	// Weighted random selection
	randomValue := rand.Float64() * totalWeight
	var cumulativeWeight float64

	for _, cw := range weightedChannels {
		cumulativeWeight += cw.weight
		if randomValue <= cumulativeWeight {
			return cw.channel
		}
	}

	// Fallback to last channel
	return weightedChannels[len(weightedChannels)-1].channel
}

// selectByStaticWeight uses the original static weight-based selection
func (cs *ChannelSelector) selectByStaticWeight(channels []*Channel) *Channel {
	var totalWeight int
	for _, ch := range channels {
		totalWeight += ch.GetWeight()
	}

	if totalWeight == 0 {
		// Equal weight for all channels
		return channels[rand.Intn(len(channels))]
	}

	randomWeight := rand.Intn(totalWeight)
	var cumulativeWeight int

	for _, ch := range channels {
		cumulativeWeight += ch.GetWeight()
		if randomWeight < cumulativeWeight {
			return ch
		}
	}

	return channels[len(channels)-1]
}

// GetChannelStats returns statistics for all channels
func (cs *ChannelSelector) GetChannelStats() map[int]map[string]interface{} {
	cs.metricsLock.RLock()
	defer cs.metricsLock.RUnlock()

	stats := make(map[int]map[string]interface{})
	for id, m := range cs.metrics {
		m.mu.RLock()
		stats[id] = map[string]interface{}{
			"success_count": m.SuccessCount,
			"failure_count": m.FailureCount,
			"avg_latency":   m.AvgLatency,
			"success_rate":  m.SuccessRate,
			"score":         m.Score,
			"last_used":     m.LastUsed,
		}
		m.mu.RUnlock()
	}
	return stats
}

// ResetChannelStats resets metrics for a specific channel
func (cs *ChannelSelector) ResetChannelStats(channelId int) {
	cs.metricsLock.Lock()
	defer cs.metricsLock.Unlock()

	delete(cs.metrics, channelId)
}

// cleanupLoop periodically removes stale metrics
func (cs *ChannelSelector) cleanupLoop() {
	ticker := time.NewTicker(cs.ttl)
	defer ticker.Stop()

	for range ticker.C {
		cs.cleanup()
	}
}

// cleanup removes metrics older than TTL
func (cs *ChannelSelector) cleanup() {
	cs.metricsLock.Lock()
	defer cs.metricsLock.Unlock()

	cutoff := time.Now().Add(-cs.ttl)
	for id, m := range cs.metrics {
		m.mu.RLock()
		lastUsed := m.LastUsed
		m.mu.RUnlock()

		if lastUsed.Before(cutoff) {
			delete(cs.metrics, id)
		}
	}
}

// RecordChannelResult is a convenience function to record channel usage result
func RecordChannelResult(channelId int, success bool, latencyMs int) {
	if common.MemoryCacheEnabled {
		GetChannelSelector().RecordResult(channelId, success, latencyMs)
	}
}

// GetChannelPerformanceStats returns performance statistics for API exposure
func GetChannelPerformanceStats() map[int]map[string]interface{} {
	return GetChannelSelector().GetChannelStats()
}
