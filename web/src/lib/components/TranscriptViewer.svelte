<script lang="ts">
  import type { TranscriptSentence } from '$lib/types';
  import { isPhoneIdentity, obfuscatePhone } from '$lib/utils/phone';

  let { sentences = [] }: { sentences?: TranscriptSentence[] } = $props();

  // Colors for different speakers/channels
  const channelColors = [
    'text-twilio-blue-60',
    'text-twilio-purple-60',
    'text-twilio-green-60',
    'text-twilio-orange-60'
  ];

  const channelBgColors = [
    'bg-twilio-blue-60',
    'bg-twilio-purple-60',
    'bg-twilio-green-60',
    'bg-twilio-orange-60'
  ];

  function formatTime(seconds: number): string {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  }

  function speakerLabel(channel: number): string {
    return `Speaker ${channel + 1}`;
  }
</script>

<div class="flex flex-col h-full">
  <h3 class="text-sm font-semibold uppercase tracking-wider text-twilio-gray-60 dark:text-twilio-gray-40 mb-3 px-1">
    Transcript
  </h3>

  {#if sentences.length === 0}
    <div class="flex-1 flex items-center justify-center text-twilio-gray-50">
      <p>No transcript available</p>
    </div>
  {:else}
    <div class="flex-1 overflow-y-auto space-y-3">
      {#each sentences as sentence, i (i)}
        <div class="group">
          <div class="flex items-baseline gap-2 mb-0.5">
            <span class="text-xs font-medium {channelColors[sentence.media_channel % channelColors.length]}">
              {speakerLabel(sentence.media_channel)}
            </span>
            <span class="text-xs text-twilio-gray-50">
              {formatTime(sentence.start_time)}
            </span>
          </div>
          <div class="pl-0 relative">
            <div class="absolute left-0 top-0 bottom-0 w-0.5 rounded {channelBgColors[sentence.media_channel % channelBgColors.length]} opacity-30"></div>
            <p class="pl-3 text-sm text-twilio-gray-90 dark:text-twilio-gray-10 leading-relaxed">
              {sentence.transcript}
            </p>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>
