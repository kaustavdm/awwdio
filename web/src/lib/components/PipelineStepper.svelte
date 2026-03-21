<script lang="ts">
  import type { PipelineStatus } from '$lib/types';

  let { status = 'unknown' }: { status?: PipelineStatus } = $props();

  const steps = [
    { key: 'recording', label: 'Record' },
    { key: 'composing', label: 'Compose' },
    { key: 'transcribing', label: 'Analyze' },
    { key: 'ready', label: 'Ready' }
  ];

  const statusOrder: Record<string, number> = {
    unknown: -1,
    recording: 0,
    composing: 1,
    transcribing: 2,
    ready: 3,
    failed: -2
  };

  let currentIndex = $derived(statusOrder[status] ?? -1);
  let isFailed = $derived(status === 'failed');
</script>

<div class="flex flex-col md:flex-row items-start md:items-center gap-2 md:gap-0">
  {#each steps as step, i (step.key)}
    <div class="flex items-center gap-2 md:gap-0">
      <!-- Step circle -->
      <div class="flex items-center justify-center w-8 h-8 rounded-full text-sm font-semibold
        {i < currentIndex ? 'bg-twilio-green-60 text-white' :
         i === currentIndex && !isFailed ? 'bg-twilio-blue-60 text-white ring-2 ring-twilio-blue-60 ring-offset-2 dark:ring-offset-twilio-gray-90' :
         i === currentIndex && isFailed ? 'bg-twilio-red-60 text-white' :
         'bg-twilio-gray-20 dark:bg-twilio-gray-70 text-twilio-gray-60 dark:text-twilio-gray-40'}">
        {#if i < currentIndex}
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"/></svg>
        {:else if i === currentIndex && isFailed}
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M6 18L18 6M6 6l12 12"/></svg>
        {:else}
          {i + 1}
        {/if}
      </div>

      <!-- Step label -->
      <span class="text-sm md:ml-1 md:mr-3
        {i <= currentIndex && !isFailed ? 'text-twilio-gray-100 dark:text-twilio-gray-0 font-medium' :
         i === currentIndex && isFailed ? 'text-twilio-red-60 dark:text-twilio-red-30 font-medium' :
         'text-twilio-gray-50 dark:text-twilio-gray-50'}">
        {step.label}
      </span>

      <!-- Connector line (not after last step) -->
      {#if i < steps.length - 1}
        <div class="hidden md:block w-8 h-0.5
          {i < currentIndex ? 'bg-twilio-green-60' : 'bg-twilio-gray-20 dark:bg-twilio-gray-70'}"></div>
      {/if}
    </div>
  {/each}
</div>
