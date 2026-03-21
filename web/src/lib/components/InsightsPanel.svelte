<script lang="ts">
  import type { OperatorResult } from '$lib/types';

  let { results = [] }: { results?: OperatorResult[] } = $props();

  function isSummary(r: OperatorResult): boolean {
    return r.name?.toLowerCase().includes('summary') || !!r.text_generation_results?.result;
  }

  function isSentiment(r: OperatorResult): boolean {
    return r.name?.toLowerCase().includes('sentiment') || !!r.predicted_label;
  }

  function getSentimentColor(label: string): string {
    const l = label.toLowerCase();
    if (l.includes('positive')) return 'text-twilio-green-60';
    if (l.includes('negative')) return 'text-twilio-red-60';
    if (l.includes('mixed')) return 'text-twilio-orange-60';
    return 'text-twilio-gray-60';
  }

  function getSentimentBgColor(label: string): string {
    const l = label.toLowerCase();
    if (l.includes('positive')) return 'bg-twilio-green-60';
    if (l.includes('negative')) return 'bg-twilio-red-60';
    if (l.includes('mixed')) return 'bg-twilio-orange-60';
    return 'bg-twilio-gray-40';
  }
</script>

<div class="flex flex-col h-full">
  <h3 class="text-sm font-semibold uppercase tracking-wider text-twilio-gray-60 dark:text-twilio-gray-40 mb-3 px-1">
    AI Insights
  </h3>

  {#if results.length === 0}
    <div class="flex-1 flex items-center justify-center text-twilio-gray-50">
      <p>No insights available</p>
    </div>
  {:else}
    <div class="flex-1 overflow-y-auto space-y-4">
      {#each results as result (result.operator_sid)}
        <div class="bg-twilio-gray-0 dark:bg-twilio-gray-80 rounded-lg p-4 border border-twilio-gray-20 dark:border-twilio-gray-70">
          {#if isSummary(result)}
            <!-- Summary Card -->
            <div class="flex items-center gap-2 mb-2">
              <svg class="w-4 h-4 text-twilio-blue-60" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
              </svg>
              <h4 class="text-sm font-semibold text-twilio-gray-90 dark:text-twilio-gray-10">Summary</h4>
            </div>
            <p class="text-sm text-twilio-gray-70 dark:text-twilio-gray-30 leading-relaxed">
              {result.text_generation_results?.result || 'No summary available'}
            </p>

          {:else if isSentiment(result)}
            <!-- Sentiment Card -->
            <div class="flex items-center gap-2 mb-2">
              <svg class="w-4 h-4 text-twilio-purple-60" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
              <h4 class="text-sm font-semibold text-twilio-gray-90 dark:text-twilio-gray-10">Sentiment</h4>
            </div>
            {#if result.predicted_label}
              <div class="flex items-center gap-2 mb-2">
                <span class="text-sm font-medium capitalize {getSentimentColor(result.predicted_label)}">
                  {result.predicted_label}
                </span>
              </div>
            {/if}
            {#if result.label_probabilities}
              <div class="space-y-1.5">
                {#each result.label_probabilities as lp}
                  <div class="flex items-center gap-2">
                    <span class="text-xs w-16 text-twilio-gray-60 dark:text-twilio-gray-40 capitalize">{lp.label}</span>
                    <div class="flex-1 h-2 bg-twilio-gray-20 dark:bg-twilio-gray-70 rounded-full overflow-hidden">
                      <div class="h-full rounded-full {getSentimentBgColor(lp.label)}" style="width: {Math.round(lp.probability * 100)}%"></div>
                    </div>
                    <span class="text-xs text-twilio-gray-50 w-8 text-right">{Math.round(lp.probability * 100)}%</span>
                  </div>
                {/each}
              </div>
            {/if}

          {:else}
            <!-- Generic Operator Card -->
            <div class="flex items-center gap-2 mb-2">
              <svg class="w-4 h-4 text-twilio-gray-60" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
              </svg>
              <h4 class="text-sm font-semibold text-twilio-gray-90 dark:text-twilio-gray-10">{result.name || 'Operator Result'}</h4>
            </div>
            <pre class="text-xs text-twilio-gray-70 dark:text-twilio-gray-30 bg-twilio-gray-10 dark:bg-twilio-gray-90 p-2 rounded overflow-x-auto">{JSON.stringify(result.normalized_result, null, 2)}</pre>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>
