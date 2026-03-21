<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { onMount, onDestroy } from 'svelte';
  import { apiGet } from '$lib/api';
  import { createSyncStore } from '$lib/stores/sync';
  import PipelineStepper from '$lib/components/PipelineStepper.svelte';
  import SyncStatusBadge from '$lib/components/SyncStatusBadge.svelte';
  import TranscriptViewer from '$lib/components/TranscriptViewer.svelte';
  import InsightsPanel from '$lib/components/InsightsPanel.svelte';
  import type { TranscriptSentence, OperatorResult, PipelineStatus } from '$lib/types';

  let callId = $derived($page.params.callId ?? '');
  const syncStore = createSyncStore();

  let sentences = $state<TranscriptSentence[]>([]);
  let results = $state<OperatorResult[]>([]);
  let dataLoaded = $state(false);
  let dataError = $state('');
  let loadingData = $state(false);

  // Subscribe to the sync store
  let syncState = $state<any>(null);
  const unsubscribe = syncStore.subscribe((value) => {
    syncState = value;
  });

  // Derive pipeline status from sync document data
  let pipelineStatus = $derived<PipelineStatus>(
    syncState?.documentData?.status || 'unknown'
  );

  // When status becomes 'ready', fetch intelligence data
  $effect(() => {
    if (pipelineStatus === 'ready' && !dataLoaded && !loadingData) {
      fetchIntelligenceData();
    }
  });

  onMount(() => {
    syncStore.init(callId);
  });

  onDestroy(() => {
    syncStore.destroy();
    unsubscribe();
  });

  async function fetchIntelligenceData() {
    loadingData = true;
    dataError = '';

    try {
      const [sentencesResp, resultsResp] = await Promise.all([
        apiGet<TranscriptSentence[]>(`/api/intelligence/sentences/${encodeURIComponent(callId)}`),
        apiGet<OperatorResult[]>(`/api/intelligence/results/${encodeURIComponent(callId)}`)
      ]);

      if (sentencesResp.error) {
        console.error('Failed to fetch sentences:', sentencesResp.error);
      } else if (sentencesResp.data) {
        sentences = sentencesResp.data;
      }

      if (resultsResp.error) {
        console.error('Failed to fetch results:', resultsResp.error);
      } else if (resultsResp.data) {
        results = resultsResp.data;
      }

      dataLoaded = true;
    } catch (e) {
      dataError = e instanceof Error ? e.message : 'Failed to load results';
    } finally {
      loadingData = false;
    }
  }
</script>

<svelte:head>
  <title>Call Summary - Awwdio</title>
</svelte:head>

<div class="min-h-screen flex flex-col bg-twilio-gray-10 dark:bg-twilio-gray-100">
  <!-- Header -->
  <div class="bg-twilio-gray-0 dark:bg-twilio-gray-90 shadow-sm px-4 py-3 flex items-center justify-between">
    <div class="flex items-center gap-4">
      <button
        onclick={() => goto('/')}
        class="text-twilio-gray-60 hover:text-twilio-gray-80 dark:hover:text-twilio-gray-20"
        aria-label="Back to home"
      >
        <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"/>
        </svg>
      </button>
      <h1 class="text-lg font-semibold">Call Summary</h1>
    </div>
    <SyncStatusBadge status={pipelineStatus} />
  </div>

  <!-- Pipeline Stepper -->
  <div class="px-4 py-4 bg-twilio-gray-0 dark:bg-twilio-gray-90 border-b border-twilio-gray-20 dark:border-twilio-gray-80">
    <PipelineStepper status={pipelineStatus} />
  </div>

  <!-- Main Content -->
  {#if syncState?.error}
    <!-- Sync Error -->
    <div class="flex-1 flex items-center justify-center p-4">
      <div class="text-center max-w-md">
        <div class="w-16 h-16 rounded-full bg-twilio-red-10 dark:bg-twilio-red-100 flex items-center justify-center mx-auto mb-4">
          <svg class="w-8 h-8 text-twilio-red-60" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"/>
          </svg>
        </div>
        <p class="text-twilio-gray-70 dark:text-twilio-gray-30">{syncState.error}</p>
        <button
          onclick={() => goto('/')}
          class="mt-4 px-4 py-2 bg-twilio-blue-60 hover:bg-twilio-blue-70 text-white rounded-lg"
        >
          Back to Home
        </button>
      </div>
    </div>

  {:else if pipelineStatus === 'failed'}
    <!-- Pipeline Failed -->
    <div class="flex-1 flex items-center justify-center p-4">
      <div class="text-center max-w-md">
        <div class="w-16 h-16 rounded-full bg-twilio-red-10 dark:bg-twilio-red-100 flex items-center justify-center mx-auto mb-4">
          <svg class="w-8 h-8 text-twilio-red-60" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </div>
        <h2 class="text-lg font-semibold mb-2">Analysis Failed</h2>
        <p class="text-twilio-gray-60 dark:text-twilio-gray-40">
          {syncState?.documentData?.error || 'The conversation analysis could not be completed.'}
        </p>
        <button
          onclick={() => goto('/')}
          class="mt-4 px-4 py-2 bg-twilio-blue-60 hover:bg-twilio-blue-70 text-white rounded-lg"
        >
          Back to Home
        </button>
      </div>
    </div>

  {:else if pipelineStatus !== 'ready'}
    <!-- Pipeline In Progress -->
    <div class="flex-1 flex items-center justify-center p-4">
      <div class="text-center">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-twilio-blue-60 mx-auto mb-4"></div>
        <p class="text-lg text-twilio-gray-70 dark:text-twilio-gray-30">
          {#if pipelineStatus === 'composing'}
            Composing audio from recording...
          {:else if pipelineStatus === 'transcribing'}
            Analyzing conversation...
          {:else}
            Waiting for call data...
          {/if}
        </p>
        <p class="text-sm text-twilio-gray-50 mt-2">This may take a few minutes</p>
      </div>
    </div>

  {:else}
    <!-- Results Ready -->
    {#if dataError}
      <div class="m-4 p-4 bg-twilio-red-10 dark:bg-twilio-red-100 text-twilio-red-70 dark:text-twilio-red-30 rounded-lg">
        {dataError}
        <button onclick={fetchIntelligenceData} class="ml-2 underline">Retry</button>
      </div>
    {/if}

    {#if loadingData}
      <div class="flex-1 flex items-center justify-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-twilio-blue-60"></div>
      </div>
    {:else}
      <!-- Split View: Transcript (left) | Insights (right) -->
      <div class="flex-1 px-4 py-4 flex flex-col lg:flex-row gap-4 overflow-hidden min-h-0">
        <!-- Transcript Pane -->
        <div class="lg:w-3/5 bg-twilio-gray-0 dark:bg-twilio-gray-90 rounded-lg p-4 overflow-y-auto lg:max-h-[calc(100vh-14rem)]">
          <TranscriptViewer {sentences} />
        </div>

        <!-- Insights Pane -->
        <div class="lg:w-2/5 bg-twilio-gray-0 dark:bg-twilio-gray-90 rounded-lg p-4 overflow-y-auto lg:max-h-[calc(100vh-14rem)]">
          <InsightsPanel {results} />
        </div>
      </div>
    {/if}
  {/if}
</div>
