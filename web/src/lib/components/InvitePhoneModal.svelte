<script lang="ts">
  import { apiPost } from '$lib/api';

  let {
    roomName,
    open = false,
    onclose,
    onsuccess
  }: {
    roomName: string;
    open?: boolean;
    onclose?: () => void;
    onsuccess?: () => void;
  } = $props();

  let phoneNumber = $state('');
  let error = $state('');
  let loading = $state(false);
  let success = $state(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = '';
    loading = true;

    try {
      const resp = await apiPost<{ success: boolean; callSid: string }>('/api/voice/dial', {
        roomName,
        phoneNumber
      });

      if (resp.error || !resp.data?.success) {
        throw new Error(resp.error || 'Failed to place call');
      }

      success = true;
      setTimeout(() => {
        onsuccess?.();
        handleClose();
      }, 2000);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to place call';
    } finally {
      loading = false;
    }
  }

  function handleClose() {
    phoneNumber = '';
    error = '';
    loading = false;
    success = false;
    onclose?.();
  }
</script>

{#if open}
  <!-- Overlay -->
  <div
    class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
    onclick={(e) => { if (e.target === e.currentTarget) handleClose(); }}
    role="dialog"
    aria-modal="true"
  >
    <div class="bg-twilio-gray-0 dark:bg-twilio-gray-90 rounded-xl shadow-2xl max-w-md w-full mx-4 p-6">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold">Invite via Phone</h2>
        <button onclick={handleClose} class="text-twilio-gray-60 hover:text-twilio-gray-80 dark:hover:text-twilio-gray-20" aria-label="Close">
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>

      {#if success}
        <div class="text-center py-6">
          <div class="w-12 h-12 rounded-full bg-twilio-green-60 flex items-center justify-center mx-auto mb-3">
            <svg class="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
            </svg>
          </div>
          <p class="text-twilio-gray-80 dark:text-twilio-gray-20">Calling {phoneNumber}...</p>
        </div>
      {:else}
        <p class="text-sm text-twilio-gray-60 dark:text-twilio-gray-40 mb-4">
          Enter a phone number to dial into this call.
        </p>

        {#if error}
          <div class="mb-4 p-3 bg-twilio-red-10 dark:bg-twilio-red-100 text-twilio-red-70 dark:text-twilio-red-30 rounded-lg text-sm">
            {error}
          </div>
        {/if}

        <form onsubmit={handleSubmit}>
          <input
            type="tel"
            bind:value={phoneNumber}
            placeholder="+1 (555) 000-0000"
            class="w-full px-4 py-3 rounded-lg border border-twilio-gray-30 dark:border-twilio-gray-70 bg-white dark:bg-twilio-gray-80 mb-4"
            required
            disabled={loading}
          />
          <div class="flex gap-3">
            <button
              type="button"
              onclick={handleClose}
              class="px-4 py-2 rounded-lg border border-twilio-gray-30 dark:border-twilio-gray-70 hover:bg-twilio-gray-10 dark:hover:bg-twilio-gray-80"
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              class="flex-1 py-2 bg-twilio-green-60 hover:bg-twilio-green-70 text-white font-semibold rounded-lg disabled:opacity-50"
              disabled={loading || !phoneNumber}
            >
              {loading ? 'Calling...' : 'Call'}
            </button>
          </div>
        </form>
      {/if}
    </div>
  </div>
{/if}
