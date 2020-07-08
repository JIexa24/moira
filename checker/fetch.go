package checker

import (
	"fmt"

	"github.com/moira-alert/moira/checker/metrics/conversion"
	metricSource "github.com/moira-alert/moira/metric_source"
)

func (triggerChecker *TriggerChecker) fetchTriggerMetrics() (map[string][]metricSource.MetricData, error) {
	triggerMetricsData, metrics, err := triggerChecker.fetch()
	if err != nil {
		return triggerMetricsData, err
	}
	triggerChecker.cleanupMetricsValues(metrics, triggerChecker.until)

	if len(triggerChecker.lastCheck.Metrics) == 0 {
		if hasEmptyTargets, emptyTargets := conversion.HasEmptyTargets(triggerMetricsData); hasEmptyTargets {
			return nil, ErrTriggerHasEmptyTargets{targets: emptyTargets}
		}
		if conversion.HasOnlyWildcards(triggerMetricsData) {
			return triggerMetricsData, ErrTriggerHasOnlyWildcards{}
		}
	}

	return triggerMetricsData, nil
}

func (triggerChecker *TriggerChecker) fetch() (map[string][]metricSource.MetricData, []string, error) {
	triggerMetricsData := make(map[string][]metricSource.MetricData)
	metricsArr := make([]string, 0)

	isSimpleTrigger := triggerChecker.trigger.IsSimple()
	for targetIndex, target := range triggerChecker.trigger.Targets {
		targetIndex++ // increasing target index to have target names started from 1 instead of 0
		fetchResult, err := triggerChecker.source.Fetch(target, triggerChecker.from, triggerChecker.until, isSimpleTrigger)
		if err != nil {
			id := ""
			if triggerChecker.trigger != nil {
				id = triggerChecker.trigger.ID
			}
			triggerChecker.logger.Warningf("NOVARIABLES triggerChecker.source.Fetch ID: %s, ERROR: %v, ",
				id, err)
			return nil, nil, err
		}
		metricsData := fetchResult.GetMetricsData()

		metricsFetchResult, metricsErr := fetchResult.GetPatternMetrics()
		if metricsErr != nil {
			id := ""
			if triggerChecker.trigger != nil {
				id = triggerChecker.trigger.ID
			}
			triggerChecker.logger.Warningf("NOVARIABLES GetPatternMetrics ID: %s, ERROR: %v, ",
				id, metricsErr)
		}

		if metricsErr == nil {
			metricsArr = append(metricsArr, metricsFetchResult...)
		}

		targetName := fmt.Sprintf("t%d", targetIndex)
		triggerMetricsData[targetName] = metricsData
	}

	if triggerChecker.trigger.ID == "265cb2bf-e029-4df2-9836-b628c64a8373" {
		triggerChecker.logger.Warningf("NOVARIABLES triggerMetricsData: %#v, ",triggerMetricsData)
	}

	return triggerMetricsData, metricsArr, nil
}

func (triggerChecker *TriggerChecker) cleanupMetricsValues(metrics []string, until int64) {
	if len(metrics) > 0 {
		if err := triggerChecker.database.RemoveMetricsValues(metrics, until-triggerChecker.database.GetMetricsTTLSeconds()); err != nil {
			triggerChecker.logger.Error(err.Error())
		}
	}
}
