<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth';
	import type { Room, LocalParticipant, RemoteParticipant, RemoteTrack } from 'twilio-video';

	let callId = $state('');
	let user = $state<any>(null);
	let room: Room | null = null;
	let localParticipant: LocalParticipant | null = null;
	let remoteParticipants = $state<RemoteParticipant[]>([]);
	let error = $state('');
	let connecting = $state(false);
	let audioEnabled = $state(true);
	let videoEnabled = $state(false);

	authStore.subscribe((value) => {
		user = value;
	});

	$effect(() => {
		callId = $page.params.callId;
	});

	onMount(async () => {
		await connectToRoom();
	});

	onDestroy(() => {
		if (room) {
			room.disconnect();
		}
	});

	async function connectToRoom() {
		connecting = true;
		error = '';

		try {
			// Get access token from API
			const response = await fetch('/api/video/token', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					room: callId,
					identity: user?.displayName || user?.email || 'Anonymous'
				})
			});

			if (!response.ok) {
				throw new Error('Failed to get access token');
			}

			const data = await response.json();

			// Import Twilio Video dynamically
			const Video = await import('twilio-video');

			// Connect to the room
			room = await Video.connect(data.token, {
				name: callId,
				audio: true,
				video: false
			});

			localParticipant = room.localParticipant;

			// Handle existing participants
			room.participants.forEach(handleParticipantConnected);

			// Handle new participants
			room.on('participantConnected', handleParticipantConnected);
			room.on('participantDisconnected', handleParticipantDisconnected);

			room.on('disconnected', () => {
				remoteParticipants = [];
			});

		} catch (e) {
			console.error('Failed to connect to room:', e);
			error = 'Failed to connect to call. Please try again.';
		} finally {
			connecting = false;
		}
	}

	function handleParticipantConnected(participant: RemoteParticipant) {
		remoteParticipants = [...remoteParticipants, participant];

		participant.tracks.forEach((publication) => {
			if (publication.track) {
				handleTrackPublished(publication.track, participant);
			}
		});

		participant.on('trackSubscribed', (track) => handleTrackPublished(track, participant));
	}

	function handleParticipantDisconnected(participant: RemoteParticipant) {
		remoteParticipants = remoteParticipants.filter(p => p.sid !== participant.sid);
	}

	function handleTrackPublished(track: RemoteTrack, participant: RemoteParticipant) {
		const participantDiv = document.getElementById(`participant-${participant.sid}`);
		if (participantDiv) {
			const mediaContainer = participantDiv.querySelector('.media-container');
			if (mediaContainer) {
				const element = track.attach();
				mediaContainer.appendChild(element);
			}
		}
	}

	async function toggleAudio() {
		if (localParticipant) {
			localParticipant.audioTracks.forEach((publication) => {
				if (audioEnabled) {
					publication.track.disable();
				} else {
					publication.track.enable();
				}
			});
			audioEnabled = !audioEnabled;
		}
	}

	async function toggleVideo() {
		if (!localParticipant) return;

		if (videoEnabled) {
			// Disable video
			localParticipant.videoTracks.forEach((publication) => {
				publication.track.stop();
				localParticipant?.unpublishTrack(publication.track);
			});
			videoEnabled = false;
		} else {
			// Enable video
			const Video = await import('twilio-video');
			const videoTrack = await Video.createLocalVideoTrack();
			await localParticipant.publishTrack(videoTrack);

			// Attach to local video element
			const localVideo = document.getElementById('local-video');
			if (localVideo) {
				const element = videoTrack.attach();
				localVideo.appendChild(element);
			}

			videoEnabled = true;
		}
	}

	function leaveCall() {
		if (room) {
			room.disconnect();
		}
		goto('/');
	}

	function copyCallLink() {
		const link = `${window.location.origin}/call/${callId}/setup`;
		navigator.clipboard.writeText(link);
	}
</script>

<svelte:head>
	<title>Call - Awwdio</title>
</svelte:head>

<div class="min-h-screen flex flex-col bg-twilio-gray-10 dark:bg-twilio-gray-100">
	<!-- Header -->
	<div class="bg-twilio-gray-0 dark:bg-twilio-gray-90 shadow-sm px-4 py-3 flex items-center justify-between">
		<div class="flex items-center gap-4">
			<h1 class="text-lg font-semibold">Awwdio Call</h1>
			<button
				onclick={copyCallLink}
				class="text-sm px-3 py-1 rounded bg-twilio-gray-20 dark:bg-twilio-gray-80 hover:bg-twilio-gray-30 dark:hover:bg-twilio-gray-70"
			>
				Copy Invite Link
			</button>
		</div>
		<button
			onclick={leaveCall}
			class="px-4 py-2 bg-twilio-red-60 hover:bg-twilio-red-70 text-white rounded-lg font-semibold"
		>
			Leave Call
		</button>
	</div>

	{#if error}
		<div class="m-4 p-4 bg-twilio-red-10 dark:bg-twilio-red-100 text-twilio-red-70 dark:text-twilio-red-30 rounded-lg">
			{error}
		</div>
	{/if}

	{#if connecting}
		<div class="flex-1 flex items-center justify-center">
			<div class="text-center">
				<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-twilio-blue-60 mx-auto mb-4"></div>
				<p class="text-lg">Connecting to call...</p>
			</div>
		</div>
	{:else}
		<!-- Participants Grid -->
		<div class="flex-1 p-4">
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 h-full">
				<!-- Local Participant -->
				{#if localParticipant}
					<div class="bg-twilio-gray-90 rounded-lg overflow-hidden relative aspect-video">
						<div id="local-video" class="media-container w-full h-full flex items-center justify-center">
							{#if !videoEnabled}
								<div class="flex flex-col items-center">
									<div class="w-20 h-20 rounded-full bg-twilio-blue-60 flex items-center justify-center text-white text-2xl font-bold mb-2">
										{(user?.displayName || user?.email || 'You')[0].toUpperCase()}
									</div>
								</div>
							{/if}
						</div>
						<div class="absolute bottom-2 left-2 bg-black bg-opacity-50 px-3 py-1 rounded text-white text-sm">
							{user?.displayName || user?.email || 'You'} (You)
						</div>
						<div class="absolute top-2 right-2 flex gap-1">
							<span class="px-2 py-1 bg-twilio-green-60 text-white text-xs rounded">Web</span>
						</div>
					</div>
				{/if}

				<!-- Remote Participants -->
				{#each remoteParticipants as participant (participant.sid)}
					<div id="participant-{participant.sid}" class="bg-twilio-gray-90 rounded-lg overflow-hidden relative aspect-video">
						<div class="media-container w-full h-full flex items-center justify-center">
							<div class="flex flex-col items-center">
								<div class="w-20 h-20 rounded-full bg-twilio-purple-60 flex items-center justify-center text-white text-2xl font-bold mb-2">
									{participant.identity[0].toUpperCase()}
								</div>
							</div>
						</div>
						<div class="absolute bottom-2 left-2 bg-black bg-opacity-50 px-3 py-1 rounded text-white text-sm">
							{participant.identity}
						</div>
						<div class="absolute top-2 right-2 flex gap-1">
							<span class="px-2 py-1 bg-twilio-green-60 text-white text-xs rounded">Web</span>
						</div>
					</div>
				{/each}
			</div>
		</div>

		<!-- Controls -->
		<div class="bg-twilio-gray-0 dark:bg-twilio-gray-90 px-4 py-6 shadow-lg">
			<div class="flex justify-center gap-4">
				<button
					onclick={toggleAudio}
					class="p-4 rounded-full {audioEnabled ? 'bg-twilio-gray-20 dark:bg-twilio-gray-80' : 'bg-twilio-red-60'} hover:opacity-80 transition-opacity"
					aria-label={audioEnabled ? 'Mute' : 'Unmute'}
				>
					{#if audioEnabled}
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
						</svg>
					{:else}
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" clip-rule="evenodd" />
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" />
						</svg>
					{/if}
				</button>

				<button
					onclick={toggleVideo}
					class="p-4 rounded-full {videoEnabled ? 'bg-twilio-gray-20 dark:bg-twilio-gray-80' : 'bg-twilio-gray-40 dark:bg-twilio-gray-70'} hover:opacity-80 transition-opacity"
					aria-label={videoEnabled ? 'Turn off video' : 'Turn on video'}
				>
					{#if videoEnabled}
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
						</svg>
					{:else}
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
						</svg>
					{/if}
				</button>
			</div>
		</div>
	{/if}
</div>
