export interface AudioAnalyzer {
  getFrequencyData(): Uint8Array;
  readonly binCount: number;
  isActive(): boolean;
  destroy(): void;
}

/**
 * Create an AudioAnalyzer from a MediaStreamTrack.
 * This connects the track to a Web Audio API AnalyserNode for FFT data.
 * Returns a no-op analyzer if AudioContext is not available.
 */
export function createAudioAnalyzer(
  mediaStreamTrack: MediaStreamTrack,
  options?: { fftSize?: number; smoothing?: number }
): AudioAnalyzer {
  const fftSize = options?.fftSize ?? 1024;
  const smoothing = options?.smoothing ?? 0.5;
  const emptyData = new Uint8Array(fftSize / 2);

  try {
    const AudioCtx = window.AudioContext || (window as any).webkitAudioContext;
    if (!AudioCtx) throw new Error('No AudioContext');

    const audioContext = new AudioCtx();
    const analyser = audioContext.createAnalyser();
    analyser.fftSize = fftSize;
    analyser.smoothingTimeConstant = smoothing;

    const stream = new MediaStream([mediaStreamTrack]);
    const source = audioContext.createMediaStreamSource(stream);
    source.connect(analyser);

    const data = new Uint8Array(analyser.frequencyBinCount);

    return {
      getFrequencyData() {
        if (mediaStreamTrack.readyState === 'live') {
          analyser.getByteFrequencyData(data);
        }
        return data;
      },
      get binCount() {
        return analyser.frequencyBinCount;
      },
      isActive() {
        return mediaStreamTrack.readyState === 'live';
      },
      destroy() {
        try {
          source.disconnect();
          audioContext.close();
        } catch {
          // Already closed
        }
      }
    };
  } catch {
    return {
      getFrequencyData: () => emptyData,
      get binCount() { return fftSize / 2; },
      isActive: () => false,
      destroy: () => {}
    };
  }
}

/**
 * Downsample full FFT frequency data to a fixed number of bars.
 * Returns an array of normalized values (0-1) for equalizer display.
 */
export function downsampleToBars(data: Uint8Array, barCount: number): number[] {
  const bars: number[] = [];
  const step = Math.max(1, Math.floor(data.length / barCount));
  for (let i = 0; i < barCount; i++) {
    let sum = 0;
    const start = i * step;
    const end = Math.min(start + step, data.length);
    for (let j = start; j < end; j++) {
      sum += data[j];
    }
    bars.push(sum / (end - start) / 255);
  }
  return bars;
}
