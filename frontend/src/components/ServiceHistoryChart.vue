<template>
  <div class="space-y-2">
    <h4 v-if="title" class="text-sm font-semibold text-foreground">{{ title }}</h4>
    <div
      v-if="loading"
      class="space-y-2"
      role="status"
      aria-live="polite"
      aria-label="Loading history"
    >
      <Skeleton height="1rem" width="40%" />
      <Skeleton height="220px" />
    </div>
    <EmptyState
      v-else-if="!hasData"
      title="No history yet"
      description="Snapshots are written every 30 seconds. Check back in a minute, or pick a different time range."
      class="!border-0 !shadow-none !p-2"
    />
    <v-chart
      v-else
      :option="option"
      :update-options="{ notMerge: true }"
      :autoresize="true"
      style="height: 260px; width: 100%"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import Skeleton from "@/components/ui/Skeleton.vue";
import EmptyState from "@/components/ui/EmptyState.vue";
import { useChartTheme } from "@/composables";
import { ServiceSnapshot } from "@/types/system";
import { withAlpha } from "@/lib/chart-theme";

interface ChartSeries {
  points: ServiceSnapshot[];
  field: "cpu" | "memory" | "disk" | "netSent" | "netRecv";
  label: string;
  colorVar?: "critical" | "info" | "success" | "warning";
  yAxisFormatter?: (v: number) => string;
}

const props = withDefaults(
  defineProps<{
    points?: ServiceSnapshot[];
    loading?: boolean;
    field?: "cpu" | "memory" | "disk" | "netSent" | "netRecv";
    metricLabel?: string;
    colorVar?: "critical" | "info" | "success" | "warning";
    yAxisSuffix?: string;
    yAxisFormatter?: (v: number) => string;
    showHighWatermark?: boolean;
    series?: ChartSeries[];
    title?: string;
  }>(),
  { colorVar: "info", showHighWatermark: false },
);

const chartTheme = useChartTheme();

const colorFor = (colorVar?: ChartSeries["colorVar"]) =>
  chartTheme.value[colorVar ?? props.colorVar ?? "info"] ??
  chartTheme.value.info;

const normalizedSeries = computed<ChartSeries[]>(() => {
  if (props.series && props.series.length > 0) {
    return props.series;
  }
  if (!props.field || !props.metricLabel) return [];
  return [
    {
      points: props.points ?? [],
      field: props.field,
      label: props.metricLabel,
      colorVar: props.colorVar,
      yAxisFormatter: props.yAxisFormatter,
    },
  ];
});

const allPoints = computed(() => {
  const map = new Map<string, ServiceSnapshot>();
  for (const s of normalizedSeries.value) {
    for (const p of s.points) {
      const existing = map.get(p.ts);
      if (existing) {
        map.set(p.ts, { ...existing, ...p });
      } else {
        map.set(p.ts, { ...p });
      }
    }
  }
  return Array.from(map.values()).sort(
    (a, b) => new Date(a.ts).getTime() - new Date(b.ts).getTime(),
  );
});

const xAxis = computed(() =>
  allPoints.value.map((p) => {
    const d = new Date(p.ts);
    return d.toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
    });
  }),
);

const statuses = computed(() => allPoints.value.map((p) => p.status ?? ""));

const hasData = computed(() =>
  normalizedSeries.value.some((s) =>
    s.points.some((p) => {
      const v = p[s.field];
      return typeof v === "number" && !Number.isNaN(v);
    }),
  ),
);

function formatValue(
  v: number,
  formatter?: (value: number) => string,
): string {
  const fn = formatter ?? props.yAxisFormatter;
  if (fn) return fn(v);
  return `${v.toFixed(1)}${props.yAxisSuffix ?? ""}`;
}

const option = computed(() => {
  const markers: Record<number, string> = {};
  statuses.value.forEach((status, i) => {
    if (status && status !== "online" && status !== "running") {
      markers[i] = status;
    }
  });

  const seriesList = normalizedSeries.value.map((s) => {
    const data = allPoints.value.map((p) => {
      const sourcePoint = s.points.find((sp) => sp.ts === p.ts);
      if (!sourcePoint) return Number.NaN;
      const v = sourcePoint[s.field];
      return typeof v === "number" && !Number.isNaN(v) ? v : Number.NaN;
    });
    const color = colorFor(s.colorVar);
    return {
      name: s.label,
      type: "line",
      data,
      smooth: true,
      symbol: "circle",
      symbolSize: 5,
      showSymbol: data.some((v) => !Number.isNaN(v) && v > 0),
      connectNulls: true,
      lineStyle: { color, width: 2.5 },
      itemStyle: { color, borderColor: chartTheme.value.popover, borderWidth: 2 },
      areaStyle: {
        color: {
          type: "linear",
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: withAlpha(color, 0.33) },
            { offset: 1, color: withAlpha(color, 0) },
          ],
        },
      },
    };
  });

  const isMulti = seriesList.length > 1;

  return {
    animation: false,
    tooltip: {
      trigger: "axis",
      backgroundColor: chartTheme.value.popover,
      borderColor: chartTheme.value.border,
      borderWidth: 1,
      textStyle: { color: chartTheme.value.text },
      extraCssText:
        "box-shadow: 0 4px 12px rgba(0,0,0,0.25); border-radius: 6px;",
      formatter: (params: { axisValue: string; seriesName: string; data: number; dataIndex: number }[]) => {
        const idx = params[0]?.dataIndex ?? -1;
        const status = idx >= 0 ? markers[idx] : undefined;
        const rows = params
          .map((p) => {
            const series = normalizedSeries.value.find((s) => s.label === p.seriesName);
            const value =
              typeof p.data !== "number" || Number.isNaN(p.data)
                ? "—"
                : formatValue(p.data, series?.yAxisFormatter);
            return `${p.seriesName}: <strong>${value}</strong>`;
          })
          .join("<br/>");
        return `${params[0]?.axisValue ?? ""}<br/>${rows}${
          status ? `<br/><span style="opacity:.75">status: ${status}</span>` : ""
        }`;
      },
    },
    legend: isMulti
      ? {
          bottom: 0,
          textStyle: { color: chartTheme.value.textMuted, fontSize: 11 },
        }
      : undefined,
    grid: { left: 56, right: 16, top: 16, bottom: isMulti ? 40 : 28 },
    xAxis: {
      type: "category",
      data: xAxis.value,
      axisLine: { lineStyle: { color: chartTheme.value.border } },
      axisLabel: { color: chartTheme.value.textMuted, fontSize: 11 },
      axisTick: { lineStyle: { color: chartTheme.value.border } },
    },
    yAxis: {
      type: "value",
      splitLine: { lineStyle: { color: chartTheme.value.border, type: "dashed" } },
      axisLabel: {
        color: chartTheme.value.textMuted,
        fontSize: 11,
        formatter: (v: number) =>
          props.yAxisFormatter
            ? props.yAxisFormatter(v)
            : `${v}${props.yAxisSuffix ?? ""}`,
      },
    },
    series: seriesList.map((s) => ({
      ...s,
      markLine: props.showHighWatermark
        ? {
            silent: true,
            symbol: "none",
            lineStyle: { color: chartTheme.value.warning, type: "dashed", width: 1 },
            data: [{ yAxis: 80 }],
            label: {
              color: chartTheme.value.warning,
              fontSize: 10,
              formatter: "80%",
              position: "insideEndTop",
            },
          }
        : undefined,
    })),
  };
});
</script>