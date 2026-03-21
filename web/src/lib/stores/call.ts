import { writable } from 'svelte/store';
import type { CallParticipant } from '$lib/types';

function createCallStore() {
  const { subscribe, set, update } = writable<CallParticipant[]>([]);

  return {
    subscribe,
    addParticipant(participant: CallParticipant) {
      update(list => {
        if (list.some(p => p.sid === participant.sid)) return list;
        return [...list, participant];
      });
    },
    removeParticipant(sid: string) {
      update(list => list.filter(p => p.sid !== sid));
    },
    updateParticipant(sid: string, changes: Partial<CallParticipant>) {
      update(list =>
        list.map(p => p.sid === sid ? { ...p, ...changes } : p)
      );
    },
    reset() {
      set([]);
    }
  };
}

export const callStore = createCallStore();
