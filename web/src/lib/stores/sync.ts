import { writable, get } from 'svelte/store';
import { browser } from '$app/environment';
import { apiGet } from '$lib/api';
import type { SyncDocumentData } from '$lib/types';

export interface SyncState {
  connected: boolean;
  loading: boolean;
  documentData: SyncDocumentData | null;
  error: string | null;
}

const initialState: SyncState = {
  connected: false,
  loading: false,
  documentData: null,
  error: null
};

export function createSyncStore() {
  const { subscribe, set, update } = writable<SyncState>(initialState);

  let syncClient: any = null;
  let document: any = null;
  let retryTimer: ReturnType<typeof setInterval> | null = null;

  async function fetchToken(): Promise<string | null> {
    const resp = await apiGet<{ token: string }>('/api/sync/token');
    if (resp.error || !resp.data) return null;
    return resp.data.token;
  }

  return {
    subscribe,

    async init(roomName: string) {
      if (!browser) return;

      update(s => ({ ...s, loading: true, error: null }));

      try {
        const token = await fetchToken();
        if (!token) {
          update(s => ({ ...s, loading: false, error: 'Failed to get Sync token' }));
          return;
        }

        const { SyncClient } = await import('twilio-sync');
        syncClient = new SyncClient(token);

        syncClient.on('tokenAboutToExpire', async () => {
          const newToken = await fetchToken();
          if (newToken) syncClient.updateToken(newToken);
        });

        update(s => ({ ...s, connected: true }));

        // Try to open the document; it may not exist yet if the pipeline hasn't started
        await subscribeToDocument(roomName);
      } catch (err) {
        update(s => ({
          ...s,
          loading: false,
          error: err instanceof Error ? err.message : 'Failed to connect to Sync'
        }));
      }
    },

    destroy() {
      if (retryTimer) {
        clearInterval(retryTimer);
        retryTimer = null;
      }
      if (document) {
        document.removeAllListeners?.();
        document = null;
      }
      if (syncClient) {
        syncClient.shutdown?.();
        syncClient = null;
      }
      set(initialState);
    }
  };

  async function subscribeToDocument(roomName: string) {
    try {
      document = await syncClient.document(roomName);
      update(s => ({
        ...s,
        loading: false,
        documentData: document.data as SyncDocumentData
      }));

      document.on('updated', (event: { data: SyncDocumentData }) => {
        update(s => ({ ...s, documentData: event.data }));
      });
    } catch {
      // Document may not exist yet — retry every 5 seconds, up to 60 seconds
      update(s => ({ ...s, loading: false }));

      let retries = 0;
      retryTimer = setInterval(async () => {
        retries++;
        if (retries > 12) {
          if (retryTimer) clearInterval(retryTimer);
          retryTimer = null;
          update(s => ({
            ...s,
            error: 'Call data not available. The call may not have been recorded.'
          }));
          return;
        }
        try {
          document = await syncClient.document(roomName);
          if (retryTimer) clearInterval(retryTimer);
          retryTimer = null;
          update(s => ({
            ...s,
            documentData: document.data as SyncDocumentData
          }));
          document.on('updated', (event: { data: SyncDocumentData }) => {
            update(s => ({ ...s, documentData: event.data }));
          });
        } catch {
          // Still not ready, will retry
        }
      }, 5000);
    }
  }
}
