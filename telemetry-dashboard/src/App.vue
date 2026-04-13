<script setup lang="ts">
import ConnectionBadge from "@/components/ConnectionBadge.vue";
import MetricsCards from "@/components/MetricsCards.vue";
import TelemetryCharts from "@/components/TelemetryCharts.vue";
import { useTelemetryStore } from "@/store/telemetryStore";
import { useTelemetryStream } from "@/composables/useTelemetryStream";

useTelemetryStream();
const store = useTelemetryStore();
</script>

<template>
  <main class="container">
    <header class="header">
      <h1>FinBud Telemetry Dashboard</h1>
      <ConnectionBadge />
    </header>

    <MetricsCards />
    <TelemetryCharts />

    <p v-if="store.lastError" class="error">
      {{ store.lastError }}
    </p>
  </main>
</template>

<style scoped>
.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 16px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.header h1 {
  margin: 0;
  font-size: 24px;
}

.error {
  color: #b00020;
}
</style>

