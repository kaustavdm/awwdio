<script lang="ts">
  import { onDestroy, untrack } from 'svelte';
  import AudioWaveform from './AudioWaveform.svelte';
  import { createAudioAnalyzer, type AudioAnalyzer } from '$lib/utils/audio';
  import { obfuscatePhone, isPhoneIdentity } from '$lib/utils/phone';

  let {
    identity,
    type = 'browser',
    isLocal = false,
    audioTrack = null,
    videoTrack = null,
    audioEnabled = true,
    videoEnabled = false,
    displayName
  }: {
    identity: string;
    type?: 'browser' | 'pstn';
    isLocal?: boolean;
    audioTrack?: MediaStreamTrack | null;
    videoTrack?: MediaStreamTrack | null;
    audioEnabled?: boolean;
    videoEnabled?: boolean;
    displayName?: string;
  } = $props();

  let analyzer: AudioAnalyzer | null = $state(null);
  let videoEl: HTMLVideoElement | undefined = $state();

  // Display name: use provided name, obfuscate phone, or use identity
  let name = $derived(
    displayName || (type === 'pstn' || isPhoneIdentity(identity) ? obfuscatePhone(identity) : identity)
  );

  let initial = $derived(
    type === 'pstn' || isPhoneIdentity(identity) ? '#' : (name?.[0] || '?').toUpperCase()
  );

  let barColor = $derived(
    videoEnabled && videoTrack ? 'rgba(255,255,255,0.8)' :
    type === 'pstn' ? '#14B053' : '#0263E0'
  );

  // Create/destroy AudioAnalyzer when audioTrack changes.
  // Use untrack for analyzer reads to avoid circular reactivity --
  // this effect should only re-run when audioTrack changes, not when
  // analyzer itself is written.
  $effect(() => {
    const track = audioTrack;
    const prev = untrack(() => analyzer);
    if (prev) {
      prev.destroy();
    }
    if (track && track.readyState === 'live') {
      analyzer = createAudioAnalyzer(track);
    } else {
      analyzer = null;
    }
    return () => {
      const current = untrack(() => analyzer);
      if (current) {
        current.destroy();
      }
    };
  });

  // Attach video track to video element
  $effect(() => {
    if (videoEl && videoTrack) {
      const stream = new MediaStream([videoTrack]);
      videoEl.srcObject = stream;
    } else if (videoEl) {
      videoEl.srcObject = null;
    }
  });
</script>

<div class="bg-twilio-gray-90 rounded-lg overflow-hidden relative aspect-video flex flex-col">
  <!-- Video or Avatar area -->
  <div class="flex-1 flex items-center justify-center relative">
    {#if videoEnabled && videoTrack}
      <video
        bind:this={videoEl}
        autoplay
        muted={isLocal}
        playsinline
        class="w-full h-full object-cover"
      ></video>
      <!-- Waveform overlay on video -->
      <div class="absolute bottom-1 left-2 right-2">
        <AudioWaveform {analyzer} barColor={barColor} height={24} />
      </div>
    {:else}
      <div class="flex flex-col items-center">
        <div class="w-16 h-16 rounded-full flex items-center justify-center text-white text-xl font-bold mb-3
          {type === 'pstn' ? 'bg-twilio-green-60' : 'bg-twilio-blue-60'}">
          {#if type === 'pstn'}
            <svg class="w-8 h-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"/>
            </svg>
          {:else}
            {initial}
          {/if}
        </div>
        <!-- Waveform below avatar -->
        <div class="w-32">
          <AudioWaveform {analyzer} barColor={barColor} height={28} />
        </div>
      </div>
    {/if}
  </div>

  <!-- Name bar -->
  <div class="absolute bottom-0 left-0 right-0 bg-black bg-opacity-50 px-3 py-1.5 flex items-center justify-between">
    <span class="text-white text-sm truncate">
      {name}{#if isLocal} (You){/if}
    </span>
    <div class="flex gap-1">
      {#if !audioEnabled}
        <span class="px-1.5 py-0.5 bg-twilio-red-60 text-white text-xs rounded">Muted</span>
      {/if}
      <span class="px-1.5 py-0.5 text-white text-xs rounded
        {type === 'pstn' ? 'bg-twilio-green-60' : 'bg-twilio-blue-60'}">
        {type === 'pstn' ? 'Phone' : 'Web'}
      </span>
    </div>
  </div>
</div>
