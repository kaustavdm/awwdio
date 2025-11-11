<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';

	let callId = $state('');
	let joinMethod = $state<'web' | 'phone' | null>(null);
	let videoEnabled = $state(false);
	let phoneNumber = $state('');

	let audioDevices = $state<MediaDeviceInfo[]>([]);
	let videoDevices = $state<MediaDeviceInfo[]>([]);
	let selectedAudioDevice = $state<string>('');
	let selectedVideoDevice = $state<string>('');

	let localStream: MediaStream | null = null;
	let videoElement: HTMLVideoElement | null = null;
	let audioLevel = $state(0);
	let audioContext: AudioContext | null = null;
	let analyser: AnalyserNode | null = null;
	let animationFrameId: number | null = null;

	$effect(() => {
		callId = $page.params.callId;
	});

	onMount(async () => {
		// Enumerate devices
		try {
			await navigator.mediaDevices.getUserMedia({ audio: true, video: false });
			const devices = await navigator.mediaDevices.enumerateDevices();
			audioDevices = devices.filter(d => d.kind === 'audioinput');
			videoDevices = devices.filter(d => d.kind === 'videoinput');

			if (audioDevices.length > 0) {
				selectedAudioDevice = audioDevices[0].deviceId;
			}
			if (videoDevices.length > 0) {
				selectedVideoDevice = videoDevices[0].deviceId;
			}
		} catch (error) {
			console.error('Error accessing media devices:', error);
		}
	});

	onDestroy(() => {
		stopMediaStream();
	});

	async function startMediaStream() {
		try {
			const constraints: MediaStreamConstraints = {
				audio: selectedAudioDevice ? { deviceId: selectedAudioDevice } : true,
				video: videoEnabled && selectedVideoDevice ? { deviceId: selectedVideoDevice } : false
			};

			localStream = await navigator.mediaDevices.getUserMedia(constraints);

			if (videoEnabled && videoElement) {
				videoElement.srcObject = localStream;
			}

			// Setup audio level monitoring
			setupAudioMonitoring();
		} catch (error) {
			console.error('Error accessing media:', error);
		}
	}

	function stopMediaStream() {
		if (localStream) {
			localStream.getTracks().forEach(track => track.stop());
			localStream = null;
		}
		if (audioContext) {
			audioContext.close();
			audioContext = null;
		}
		if (animationFrameId) {
			cancelAnimationFrame(animationFrameId);
			animationFrameId = null;
		}
	}

	function setupAudioMonitoring() {
		if (!localStream) return;

		audioContext = new AudioContext();
		analyser = audioContext.createAnalyser();
		const source = audioContext.createMediaStreamSource(localStream);
		source.connect(analyser);
		analyser.fftSize = 256;

		const bufferLength = analyser.frequencyBinCount;
		const dataArray = new Uint8Array(bufferLength);

		function updateAudioLevel() {
			if (!analyser) return;

			analyser.getByteFrequencyData(dataArray);
			const average = dataArray.reduce((a, b) => a + b) / bufferLength;
			audioLevel = average / 255;

			animationFrameId = requestAnimationFrame(updateAudioLevel);
		}

		updateAudioLevel();
	}

	async function selectJoinMethod(method: 'web' | 'phone') {
		joinMethod = method;
		if (method === 'web') {
			await startMediaStream();
		}
	}

	async function toggleVideo() {
		videoEnabled = !videoEnabled;
		stopMediaStream();
		await startMediaStream();
	}

	async function changeAudioDevice() {
		stopMediaStream();
		await startMediaStream();
	}

	async function changeVideoDevice() {
		if (videoEnabled) {
			stopMediaStream();
			await startMediaStream();
		}
	}

	function joinCall() {
		stopMediaStream();
		goto(`/call/${callId}`);
	}

	async function submitPhoneNumber() {
		// TODO: Call API to setup phone call
		console.log('Setting up phone call for:', phoneNumber);
		goto(`/call/${callId}`);
	}
</script>

<svelte:head>
	<title>Setup Call - Awwdio</title>
</svelte:head>

<div class="min-h-screen p-4 flex items-center justify-center">
	<div class="w-full max-w-2xl">
		<h1 class="text-3xl font-bold text-center mb-8">Setup Your Call</h1>

		{#if !joinMethod}
			<div class="grid md:grid-cols-2 gap-6">
				<!-- Join on Web -->
				<button
					onclick={() => selectJoinMethod('web')}
					class="p-8 bg-twilio-gray-0 dark:bg-twilio-gray-90 rounded-lg shadow-lg hover:shadow-xl transition-all border-2 border-transparent hover:border-twilio-blue-60"
				>
					<div class="flex flex-col items-center">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mb-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
						</svg>
						<h2 class="text-xl font-semibold mb-2">Join on Web</h2>
						<p class="text-sm text-twilio-gray-60 dark:text-twilio-gray-40 text-center">
							Use your browser to join with audio and video
						</p>
					</div>
				</button>

				<!-- Join by Phone -->
				<button
					onclick={() => selectJoinMethod('phone')}
					class="p-8 bg-twilio-gray-0 dark:bg-twilio-gray-90 rounded-lg shadow-lg hover:shadow-xl transition-all border-2 border-transparent hover:border-twilio-blue-60"
				>
					<div class="flex flex-col items-center">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mb-4 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
						</svg>
						<h2 class="text-xl font-semibold mb-2">Join by Phone</h2>
						<p class="text-sm text-twilio-gray-60 dark:text-twilio-gray-40 text-center">
							Dial in with your phone for audio only
						</p>
					</div>
				</button>
			</div>
		{:else if joinMethod === 'web'}
			<div class="bg-twilio-gray-0 dark:bg-twilio-gray-90 rounded-lg shadow-xl p-8">
				<!-- Video preview -->
				{#if videoEnabled}
					<div class="mb-6">
						<video
							bind:this={videoElement}
							autoplay
							muted
							playsinline
							class="w-full rounded-lg bg-gray-900"
						></video>
					</div>
				{/if}

				<!-- Audio level meter -->
				<div class="mb-6">
					<label class="block text-sm font-medium mb-2">Audio Level</label>
					<div class="w-full h-4 bg-twilio-gray-20 dark:bg-twilio-gray-80 rounded-full overflow-hidden">
						<div
							class="h-full bg-gradient-to-r from-green-500 to-green-600 transition-all duration-75"
							style="width: {audioLevel * 100}%"
						></div>
					</div>
				</div>

				<!-- Audio device selection -->
				<div class="mb-4">
					<label class="block text-sm font-medium mb-2" for="audioDevice">Microphone</label>
					<select
						id="audioDevice"
						bind:value={selectedAudioDevice}
						onchange={changeAudioDevice}
						class="w-full px-4 py-2 rounded-lg border border-twilio-gray-30 dark:border-twilio-gray-70 bg-white dark:bg-gray-700"
					>
						{#each audioDevices as device}
							<option value={device.deviceId}>{device.label || `Microphone ${device.deviceId.slice(0, 8)}`}</option>
						{/each}
					</select>
				</div>

				<!-- Video device selection -->
				<div class="mb-4">
					<label class="block text-sm font-medium mb-2" for="videoDevice">Camera</label>
					<select
						id="videoDevice"
						bind:value={selectedVideoDevice}
						onchange={changeVideoDevice}
						disabled={!videoEnabled}
						class="w-full px-4 py-2 rounded-lg border border-twilio-gray-30 dark:border-twilio-gray-70 bg-white dark:bg-gray-700 disabled:opacity-50"
					>
						{#each videoDevices as device}
							<option value={device.deviceId}>{device.label || `Camera ${device.deviceId.slice(0, 8)}`}</option>
						{/each}
					</select>
				</div>

				<!-- Video toggle -->
				<div class="mb-6">
					<button
						onclick={toggleVideo}
						class="flex items-center gap-2 px-4 py-2 rounded-lg {videoEnabled ? 'bg-twilio-blue-60 text-white' : 'bg-twilio-gray-20 dark:bg-twilio-gray-80'}"
					>
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
						</svg>
						{videoEnabled ? 'Video On' : 'Video Off (Audio Only)'}
					</button>
				</div>

				<!-- Join button -->
				<div class="flex gap-4">
					<button
						onclick={() => { joinMethod = null; stopMediaStream(); }}
						class="px-6 py-3 rounded-lg border border-twilio-gray-30 dark:border-twilio-gray-70 hover:bg-twilio-gray-10 dark:hover:bg-twilio-gray-80"
					>
						Back
					</button>
					<button
						onclick={joinCall}
						class="flex-1 py-3 bg-twilio-blue-60 hover:bg-twilio-blue-70 text-white font-semibold rounded-lg"
					>
						Join Call
					</button>
				</div>
			</div>
		{:else if joinMethod === 'phone'}
			<div class="bg-twilio-gray-0 dark:bg-twilio-gray-90 rounded-lg shadow-xl p-8">
				<h2 class="text-xl font-semibold mb-4">Enter Your Phone Number</h2>
				<p class="text-sm text-twilio-gray-60 dark:text-twilio-gray-40 mb-6">
					We'll call you at this number to connect you to the call.
				</p>

				<form onsubmit={(e) => { e.preventDefault(); submitPhoneNumber(); }}>
					<input
						type="tel"
						bind:value={phoneNumber}
						placeholder="+1 (555) 000-0000"
						class="w-full px-4 py-3 rounded-lg border border-twilio-gray-30 dark:border-twilio-gray-70 bg-white dark:bg-gray-700 mb-6"
						required
					/>

					<div class="flex gap-4">
						<button
							type="button"
							onclick={() => { joinMethod = null; }}
							class="px-6 py-3 rounded-lg border border-twilio-gray-30 dark:border-twilio-gray-70 hover:bg-twilio-gray-10 dark:hover:bg-twilio-gray-80"
						>
							Back
						</button>
						<button
							type="submit"
							class="flex-1 py-3 bg-twilio-green-60 hover:bg-twilio-green-70 text-white font-semibold rounded-lg"
						>
							Call Me
						</button>
					</div>
				</form>
			</div>
		{/if}
	</div>
</div>
