// Participant in a call
export interface CallParticipant {
  sid: string;
  identity: string;
  type: 'browser' | 'pstn';
  audioEnabled: boolean;
  videoEnabled: boolean;
}

// Sync Document data shape (matches backend webhooks that update Sync)
export interface SyncDocumentData {
  roomSid?: string;
  roomName?: string;
  status?: 'composing' | 'transcribing' | 'ready' | 'failed';
  compositionSid?: string;
  transcriptSid?: string;
  error?: string | null;
}

// Pipeline status for the stepper UI on summary page
export type PipelineStatus = 'recording' | 'composing' | 'transcribing' | 'ready' | 'failed' | 'unknown';

// From GET /api/intelligence/sentences/{roomName}
// These match the Twilio Intelligence V2 sentence response shape
export interface TranscriptSentence {
  media_channel: number;
  sentence_index: number;
  start_time: number;
  end_time: number;
  transcript: string;
  confidence: number;
  words?: Array<{
    word: string;
    start_time: number;
    end_time: number;
    confidence: number;
  }>;
}

// From GET /api/intelligence/results/{roomName}
// These match the Twilio Intelligence V2 operator result response shape
export interface OperatorResult {
  operator_sid: string;
  name: string;
  operator_type: string;
  extract_match: boolean;
  match_probability: number;
  normalized_result: Record<string, unknown>;
  text_generation_results?: { result: string };
  label_probabilities?: Array<{ label: string; probability: number }>;
  utterance_results?: Array<Record<string, unknown>>;
  predicted_label?: string;
}
