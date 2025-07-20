import React, {useEffect, useState} from 'react';
import {useFetching} from "../hooks/useFetching";
import AdaptiveLoading from "../components/UI/AdaptiveLoading";
import ErrorField from "../components/UI/ErrorField";
import ChartAvailability from "../components/UI/metrics/ChartAvailability";
import MetricsService from "../API/MetricsService";
import Speedometer from "../components/UI/metrics/Speedometer";
import DefaultMetric from "../components/UI/metrics/DefaultMetric";
import {formatUptime} from "../utils/utils";
import "../styles/metrics.css"
import useWebSocket from "../hooks/useWebSocket";
import ChartApiWeightChanges from "../components/UI/metrics/ChartApiWeightChanges";
import ChartSystemLoad from "../components/UI/metrics/ChartSystemLoad";

const Logs = () => {
    const [dynamicMetrics, setDynamicMetrics] = useState(null)
    const [averageLoadMetrics, setAverageLoadMetrics] = useState(null)
    const [staticMetrics, setStaticMetrics] = useState(null);
    const [logs, setLogs] = useState(null);
    const [fetchLogs, isLogsLoading, LogsError] = useFetching(async () => {
        return await MetricsService.getAvailabilityLogs()
    }, 1000, 1000)
    const [fetchStaticMetrics, isStaticMetricsLoading, ] = useFetching(async () => {
        return await MetricsService.getStaticMetrics()
    }, 1000, 1000)
    const [metrics] = useWebSocket('/metrics/ws');

     useEffect(() => {
        const loadData = async () => {
            if (!isLogsLoading && !LogsError) {
                const data = await fetchStaticMetrics();
                if (data) {
                    setStaticMetrics(data.data);
                }
            }
        }
        void loadData();
    }, [isStaticMetricsLoading])

    useEffect(() => {
        if (metrics && metrics.metrics?.mem_alloc_mb && logs?.length > 0) {
            metrics?.load_avg_60 && setAverageLoadMetrics(metrics.load_avg_60)
            setDynamicMetrics(metrics.metrics);
        }
    }, [metrics]);

    useEffect(() => {
        const loadData = async () => {
            if (!isLogsLoading && logs === null){
                const data = await fetchLogs();
                if (data) {
                    setLogs(data.data);
                }
            }
        }
        void loadData();
    }, [isLogsLoading])

    return (
        <div className={"section__main"}>
            {isLogsLoading ? (
                <div className={"loading__center"}>
                    <AdaptiveLoading/>
                </div>
            ) : LogsError ? (
                <ErrorField/>
            ) : logs?.length > 0 ? (
                <div className={"area__metrics"}>
                    {staticMetrics && (
                        <div className={"metrics__list__1x1"}>
                            <DefaultMetric header={"CPU Total"} value={staticMetrics.total_cpu}/>
                            <DefaultMetric header={"RAM Total"} value={staticMetrics?.total_memory_mb}/>
                            <DefaultMetric header={"uptime"} value={formatUptime(staticMetrics.start_time)}/>

                            <DefaultMetric header={"CPU Used"} value={dynamicMetrics?.cpu_used_percent.toFixed(3)}/>
                            <DefaultMetric header={"RAM Used"} value={dynamicMetrics?.ram_used_mb}/>
                            <DefaultMetric header={"Binance Overload"} value={dynamicMetrics?.binance_overload}/>
                        </div>
                    )}
                    <Speedometer header={"CPU Usage"} percentage={dynamicMetrics?.cpu_usage_percent.toFixed(3)}/>
                    <Speedometer header={"Mem Usage"} percentage={dynamicMetrics?.memory_usage_percent.toFixed(3)}/>
                    <Speedometer header={"CPU Allocation"} percentage={dynamicMetrics?.cpu_allocation.toFixed(3)}/>
                    <Speedometer header={"Mem Allocation"} percentage={dynamicMetrics?.mem_alloc_mb}/>
                    <ChartApiWeightChanges/>
                    {staticMetrics.is_load_metrics_on === true && (
                        <ChartSystemLoad data={averageLoadMetrics}/>
                    )}
                    <ChartAvailability data={logs}/>
                </div>
            ) : (isLogsLoading === false && logs.length === 0) && (
                <ErrorField message={"logs are empty, this is non-standard situation"}/>
            )}
        </div>
    );
};

export default Logs;