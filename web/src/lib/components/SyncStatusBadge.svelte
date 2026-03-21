<script lang="ts">
  import type { PipelineStatus } from '$lib/types';

  let { status = 'unknown' }: { status?: PipelineStatus } = $props();

  const config: Record<PipelineStatus, { label: string; dotClass: string; pulse: boolean }> = {
    recording: { label: 'Recording', dotClass: 'bg-twilio-blue-60', pulse: true },
    composing: { label: 'Composing audio...', dotClass: 'bg-twilio-orange-60', pulse: true },
    transcribing: { label: 'Analyzing...', dotClass: 'bg-twilio-purple-60', pulse: true },
    ready: { label: 'Results ready', dotClass: 'bg-twilio-green-60', pulse: false },
    failed: { label: 'Analysis failed', dotClass: 'bg-twilio-red-60', pulse: false },
    unknown: { label: 'Waiting...', dotClass: 'bg-twilio-gray-40', pulse: true }
  };

  let current = $derived(config[status] || config.unknown);
</script>

<div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-twilio-gray-10 dark:bg-twilio-gray-80 text-sm">
  <span class="relative flex h-2.5 w-2.5">
    {#if current.pulse}
      <span class="animate-ping absolute inline-flex h-full w-full rounded-full {current.dotClass} opacity-75"></span>
    {/if}
    <span class="relative inline-flex rounded-full h-2.5 w-2.5 {current.dotClass}"></span>
  </span>
  <span class="text-twilio-gray-80 dark:text-twilio-gray-20">{current.label}</span>
</div>
