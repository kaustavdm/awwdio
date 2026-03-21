<script lang="ts">
  import { onDestroy, untrack } from 'svelte';
  import type { AudioAnalyzer } from '$lib/utils/audio';
  import { downsampleToBars } from '$lib/utils/audio';

  let {
    analyzer = null,
    barCount = 12,
    barColor = '#0263E0',
    height = 32
  }: {
    analyzer?: AudioAnalyzer | null;
    barCount?: number;
    barColor?: string;
    height?: number;
  } = $props();

  let bars = $state<number[]>(new Array(barCount).fill(0));
  let frameId: number | null = null;

  $effect(() => {
    // Track analyzer prop as the dependency
    const currentAnalyzer = analyzer;

    // Clean up previous animation loop (untracked -- frameId is internal state)
    untrack(() => {
      if (frameId) cancelAnimationFrame(frameId);
      frameId = null;
    });

    if (currentAnalyzer && currentAnalyzer.isActive()) {
      function tick() {
        if (!currentAnalyzer || !currentAnalyzer.isActive()) {
          bars = new Array(barCount).fill(0);
          return;
        }
        const data = currentAnalyzer.getFrequencyData();
        bars = downsampleToBars(data, barCount);
        frameId = requestAnimationFrame(tick);
      }
      tick();
    } else {
      bars = new Array(barCount).fill(0);
    }

    return () => {
      if (frameId) cancelAnimationFrame(frameId);
    };
  });
</script>

<div class="flex items-end gap-[2px]" style="height: {height}px;">
  {#each bars as level, i (i)}
    <div
      class="flex-1 rounded-sm transition-[height] duration-75"
      style="height: {Math.max(2, level * height)}px; background-color: {barColor}; min-width: 3px;"
    ></div>
  {/each}
</div>
