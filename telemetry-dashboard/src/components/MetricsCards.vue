<script setup lang="ts">
import { computed } from "vue";
import { useTelemetryStore } from "@/store/telemetryStore";

const store = useTelemetryStore();
const oneMinute = computed(() => store.snapshot?.windows["1m"]);
</script>

<template>
  <div class="cards">
    <div class="card">
      <h3>Throughput (1m)</h3>
      <p>{{ oneMinute?.throughput ?? 0 }} ev/s</p>
    </div>
    <div class="card">
      <h3>Error Rate (1m)</h3>
      <p>{{ oneMinute?.error_rate ?? 0 }}%</p>
    </div>
    <div class="card">
      <h3>Latency p95 (1m)</h3>
      <p>{{ oneMinute?.latency_ms.p95 ?? 0 }} ms</p>
    </div>
    <div class="card">
      <h3>Consumer Lag</h3>
      <p>{{ store.snapshot?.consumer_lag ?? 0 }}</p>
    </div>
  </div>
</template>

<style scoped>
.cards {
  display: grid;
  grid-template-columns: repeat(4, minmax(180px, 1fr));
  gap: 12px;
}

.card {
  background: #fff;
  border: 1px solid #efefef;
  border-radius: 10px;
  padding: 14px;
}

.card h3 {
  margin: 0;
  font-size: 13px;
  color: #666;
}

.card p {
  margin: 10px 0 0;
  font-size: 22px;
  font-weight: 600;
}
</style>

