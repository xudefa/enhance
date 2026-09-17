package metrics

// collectHistograms 收集所有直方图的快照。
func (r *simpleRegistry) collectHistograms(now int64) []Metric {
	metrics := make([]Metric, 0)
	r.histograms.Range(func(key, value any) bool {
		histogramKey := key.(string)
		histogram := value.(*simpleHistogram)
		count := histogram.Count()
		var avg float64
		if count > 0 {
			avg = histogram.Sum() / float64(count)
		}
		metric := Metric{
			Name:      metricNameFromKey(histogramKey),
			Value:     avg,
			Type:      "histogram",
			Timestamp: now,
			Count:     count,
			Sum:       histogram.Sum(),
		}
		if tags := histogram.tagsSnapshot(); len(tags) > 0 {
			metric.Tags = tags
		}
		metrics = append(metrics, metric)
		return true
	})
	return metrics
}
