<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { apiPost } from '$lib/api';
	import { authStore } from '$lib/stores/auth';

	let callId = $state('');
	let user = $state<any>(null);
	let displayName = $state('');
	let email = $state('');
	let useHeadphones = $state(true);
	let videoEnabled = $state(true);

	let audioDevices = $state<MediaDeviceInfo[]>([]);
	let videoDevices = $state<MediaDeviceInfo[]>([]);
	let outputDevices = $state<MediaDeviceInfo[]>([]);
	let selectedAudioDevice = $state<string>('');
	let selectedVideoDevice = $state<string>('');
	let selectedOutputDevice = $state<string>('');

	let localStream: MediaStream | null = null;
	let videoElement: HTMLVideoElement | null = null;
	let audioLevel = $state(0);
	let videoResolution = $state('');
	let audioContext: AudioContext | null = null;
	let analyser: AnalyserNode | null = null;
	let animationFrameId: number | null = null;

	$effect(() => {
		callId = $page.params.callId ?? '';
	});

	authStore.subscribe((value) => {
		user = value;
		if (value?.displayName) {
			displayName = value.displayName;
		} else if (value?.contact) {
			displayName = value.contact.split('@')[0];
		}
	});

	let isGuest = $derived(!user);

	onMount(async () => {
		try {
			await navigator.mediaDevices.getUserMedia({ audio: true, video: true });
			const devices = await navigator.mediaDevices.enumerateDevices();
			audioDevices = devices.filter((d) => d.kind === 'audioinput');
			videoDevices = devices.filter((d) => d.kind === 'videoinput');
			outputDevices = devices.filter((d) => d.kind === 'audiooutput');

			if (audioDevices.length > 0) selectedAudioDevice = audioDevices[0].deviceId;
			if (videoDevices.length > 0) selectedVideoDevice = videoDevices[0].deviceId;
			if (outputDevices.length > 0) {
				selectedOutputDevice =
					localStorage.getItem('awwdio-output-device') || outputDevices[0].deviceId;
			}

			await startMediaStream();
		} catch (error) {
			console.error('Error accessing media devices:', error);
		}
	});

	onDestroy(() => {
		stopMediaStream();
	});

	async function startMediaStream() {
		try {
			const echoCancellation = !useHeadphones;
			const constraints: MediaStreamConstraints = {
				audio: selectedAudioDevice
					? { deviceId: { exact: selectedAudioDevice }, echoCancellation }
					: { echoCancellation },
				video:
					videoEnabled && selectedVideoDevice
						? { deviceId: { exact: selectedVideoDevice } }
						: videoEnabled
							? true
							: false
			};

			localStream = await navigator.mediaDevices.getUserMedia(constraints);

			if (videoElement && localStream) {
				videoElement.srcObject = localStream;
				const videoTrack = localStream.getVideoTracks()[0];
				if (videoTrack) {
					videoElement.onloadedmetadata = () => {
						const settings = videoTrack.getSettings();
						const fps = Math.round(settings.frameRate ?? 0);
						const height = settings.height ?? 0;
						videoResolution = height > 0 ? `${height}p / ${fps}fps` : '';
					};
				}
			}

			setupAudioMonitoring();
		} catch (error) {
			console.error('Error accessing media:', error);
		}
	}

	function stopMediaStream() {
		if (localStream) {
			localStream.getTracks().forEach((track) => track.stop());
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

	async function changeAudioDevice() {
		stopMediaStream();
		await startMediaStream();
	}

	async function changeVideoDevice() {
		stopMediaStream();
		await startMediaStream();
	}

	function changeOutputDevice() {
		localStorage.setItem('awwdio-output-device', selectedOutputDevice);
	}

	async function setHeadphones(value: boolean) {
		useHeadphones = value;
		stopMediaStream();
		await startMediaStream();
	}

	function joinCall() {
		if (displayName && user && !user.displayName) {
			authStore.updateDisplayName(displayName);
		}
		stopMediaStream();
		goto(`/call/${callId}`);
	}
</script>

<svelte:head>
	<title>Setup Call - Awwdio</title>
</svelte:head>

<div class="min-h-screen bg-twilio-gray-90 flex items-center justify-center p-8">
	<div class="w-full max-w-5xl flex gap-12 items-center">

		<!-- Left: Form panel -->
		<div class="flex-shrink-0 w-80">
			<!-- Subtitle + heading -->
			<p class="text-twilio-gray-40 text-sm mb-2">
				You're about to join the studio
			</p>
			<h1 class="text-white text-3xl font-semibold mb-8 leading-tight">
				Let's check your cam and mic
			</h1>

			<!-- Name field -->
			<div class="mb-3">
				<div class="relative">
					<input
						type="text"
						bind:value={displayName}
						placeholder="Your name"
						class="w-full bg-twilio-gray-80 text-white placeholder-twilio-gray-50 rounded-lg px-4 py-3 pr-20 border border-twilio-gray-70 focus:outline-none focus:border-twilio-purple-60 transition-colors text-sm"
					/>
					{#if isGuest}
						<span class="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-twilio-gray-40 bg-twilio-gray-70 px-2 py-0.5 rounded">
							Guest
						</span>
					{/if}
				</div>
			</div>

			<!-- Email field -->
			<div class="mb-4">
				<input
					type="email"
					bind:value={email}
					placeholder="Email (optional)"
					class="w-full bg-twilio-gray-80 text-white placeholder-twilio-gray-50 rounded-lg px-4 py-3 border border-twilio-gray-70 focus:outline-none focus:border-twilio-purple-60 transition-colors text-sm"
				/>
			</div>

			<!-- Headphones toggle -->
			<div class="mb-6 flex gap-2">
				<button
					onclick={() => setHeadphones(false)}
					class="flex-1 py-2.5 px-3 rounded-lg text-sm font-medium transition-all {!useHeadphones
						? 'bg-twilio-purple-60 text-white'
						: 'bg-twilio-gray-80 text-twilio-gray-30 border border-twilio-gray-70 hover:border-twilio-gray-50'}"
				>
					No headphones
				</button>
				<button
					onclick={() => setHeadphones(true)}
					class="flex-1 py-2.5 px-3 rounded-lg text-sm font-medium transition-all {useHeadphones
						? 'bg-twilio-purple-60 text-white'
						: 'bg-twilio-gray-80 text-twilio-gray-30 border border-twilio-gray-70 hover:border-twilio-gray-50'}"
				>
					Using headphones
				</button>
			</div>

			<!-- Join button -->
			<button
				onclick={joinCall}
				class="w-full py-3.5 rounded-lg font-semibold text-white text-sm transition-opacity hover:opacity-90"
				style="background: linear-gradient(135deg, #8957CF 0%, #6A3FB2 100%);"
			>
				Join studio
			</button>

			<!-- Guest footer -->
			{#if isGuest}
				<p class="mt-4 text-twilio-gray-50 text-xs text-center">
					You are joining as a guest, are you the host?
					<a href="/login" class="text-white font-medium hover:text-twilio-purple-30 transition-colors">
						Log in
					</a>
				</p>
			{/if}
		</div>

		<!-- Right: Video preview + device selectors -->
		<div class="flex-1 min-w-0">
			<!-- Video preview -->
			<div class="relative rounded-xl overflow-hidden bg-twilio-gray-100 mb-3" style="aspect-ratio: 16/9;">
				{#if videoEnabled}
					<video
						bind:this={videoElement}
						autoplay
						muted
						playsinline
						class="w-full h-full object-cover"
					></video>
				{:else}
					<div class="w-full h-full flex items-center justify-center">
						<div class="w-20 h-20 rounded-full bg-twilio-gray-80 flex items-center justify-center">
							<svg xmlns="http://www.w3.org/2000/svg" class="w-9 h-9 text-twilio-gray-50" viewBox="0 0 24 24" fill="currentColor">
								<path d="M12 12c2.7 0 4.8-2.1 4.8-4.8S14.7 2.4 12 2.4 7.2 4.5 7.2 7.2 9.3 12 12 12zm0 2.4c-3.2 0-9.6 1.6-9.6 4.8v2.4h19.2v-2.4c0-3.2-6.4-4.8-9.6-4.8z"/>
							</svg>
						</div>
					</div>
				{/if}

				<!-- Resolution badge -->
				{#if videoResolution}
					<div class="absolute top-3 left-3 bg-black/50 backdrop-blur-sm text-white text-xs px-2.5 py-1 rounded-md font-medium">
						{videoResolution}
					</div>
				{/if}

				<!-- Bottom overlay: audio meter + track icons -->
				<div class="absolute bottom-3 right-3 flex items-center gap-2">
					<!-- Audio level indicator rings around mic icon -->
					<div class="relative flex items-center justify-center">
						<!-- Audio ring: scales with audioLevel -->
						<div
							class="absolute rounded-full bg-white/20 transition-all duration-75"
							style="width: {28 + audioLevel * 20}px; height: {28 + audioLevel * 20}px;"
						></div>
						<div class="relative z-10 w-7 h-7 rounded-full bg-black/50 backdrop-blur-sm flex items-center justify-center">
							<svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5 text-white" viewBox="0 0 24 24" fill="currentColor">
								<path d="M12 1a4 4 0 014 4v6a4 4 0 01-8 0V5a4 4 0 014-4zm-7 10a7 7 0 0014 0h2a9 9 0 01-8 8.94V22h-2v-2.06A9 9 0 013 11H5z"/>
							</svg>
						</div>
					</div>

					<!-- Camera icon -->
					<div class="w-7 h-7 rounded-full bg-black/50 backdrop-blur-sm flex items-center justify-center">
						<svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5 text-white" viewBox="0 0 24 24" fill="currentColor">
							<path d="M17 10.5V7a1 1 0 00-1-1H4a1 1 0 00-1 1v10a1 1 0 001 1h12a1 1 0 001-1v-3.5l4 4v-11l-4 4z"/>
						</svg>
					</div>
				</div>
			</div>

			<!-- Device selectors -->
			<div class="space-y-2">
				<!-- Camera -->
				<div class="flex items-center gap-3 bg-twilio-gray-80 rounded-lg px-4 py-3 border border-twilio-gray-70">
					<svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-twilio-gray-40 flex-shrink-0" viewBox="0 0 24 24" fill="currentColor">
						<path d="M17 10.5V7a1 1 0 00-1-1H4a1 1 0 00-1 1v10a1 1 0 001 1h12a1 1 0 001-1v-3.5l4 4v-11l-4 4z"/>
					</svg>
					<select
						bind:value={selectedVideoDevice}
						onchange={changeVideoDevice}
						class="flex-1 bg-transparent text-white text-sm focus:outline-none cursor-pointer appearance-none"
					>
						{#each videoDevices as device}
							<option value={device.deviceId} class="bg-twilio-gray-80">
								{device.label || `Camera ${device.deviceId.slice(0, 8)}`}
							</option>
						{/each}
					</select>
					<svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-twilio-gray-40 flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7"/>
					</svg>
				</div>

				<!-- Microphone -->
				<div class="flex items-center gap-3 bg-twilio-gray-80 rounded-lg px-4 py-3 border border-twilio-gray-70">
					<svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-twilio-gray-40 flex-shrink-0" viewBox="0 0 24 24" fill="currentColor">
						<path d="M12 1a4 4 0 014 4v6a4 4 0 01-8 0V5a4 4 0 014-4zm-7 10a7 7 0 0014 0h2a9 9 0 01-8 8.94V22h-2v-2.06A9 9 0 013 11H5z"/>
					</svg>
					<select
						bind:value={selectedAudioDevice}
						onchange={changeAudioDevice}
						class="flex-1 bg-transparent text-white text-sm focus:outline-none cursor-pointer appearance-none"
					>
						{#each audioDevices as device}
							<option value={device.deviceId} class="bg-twilio-gray-80">
								{device.label || `Microphone ${device.deviceId.slice(0, 8)}`}
							</option>
						{/each}
					</select>
					<svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-twilio-gray-40 flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7"/>
					</svg>
				</div>

				<!-- Speaker / Output -->
				{#if outputDevices.length > 0}
					<div class="flex items-center gap-3 bg-twilio-gray-80 rounded-lg px-4 py-3 border border-twilio-gray-70">
						<svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-twilio-gray-40 flex-shrink-0" viewBox="0 0 24 24" fill="currentColor">
							<path d="M3 9v6h4l5 5V4L7 9H3zm13.5 3A4.5 4.5 0 0014 7.97v8.05c1.48-.73 2.5-2.25 2.5-4.02zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z"/>
						</svg>
						<select
							bind:value={selectedOutputDevice}
							onchange={changeOutputDevice}
							class="flex-1 bg-transparent text-white text-sm focus:outline-none cursor-pointer appearance-none"
						>
							{#each outputDevices as device}
								<option value={device.deviceId} class="bg-twilio-gray-80">
									{device.label || `Speaker ${device.deviceId.slice(0, 8)}`}
								</option>
							{/each}
						</select>
						<svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-twilio-gray-40 flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7"/>
						</svg>
					</div>
				{/if}
			</div>
		</div>

	</div>
</div>
