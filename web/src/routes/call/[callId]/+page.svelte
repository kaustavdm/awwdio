<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy, untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth';
	import { apiPost } from '$lib/api';
	import { isPhoneIdentity } from '$lib/utils/phone';
	import ParticipantTile from '$lib/components/ParticipantTile.svelte';
	import InvitePhoneModal from '$lib/components/InvitePhoneModal.svelte';
	import type { Room, LocalParticipant, RemoteParticipant, RemoteTrack, LocalTrackPublication, RemoteTrackPublication } from 'twilio-video';

	let callId = $derived($page.params.callId ?? '');
	let user = $state<any>(null);
	let room: Room | null = null;
	let localParticipant: LocalParticipant | null = $state<LocalParticipant | null>(null);
	let remoteParticipants = $state<RemoteParticipant[]>([]);
	let error = $state('');
	let connecting = $state(false);
	let audioEnabled = $state(true);
	let videoEnabled = $state(false);
	let localAudioTrack = $state<MediaStreamTrack | null>(null);
	let localVideoTrack = $state<MediaStreamTrack | null>(null);
	let remoteTrackMap = $state<Map<string, { audio: MediaStreamTrack | null; video: MediaStreamTrack | null }>>(new Map());
	let showInviteModal = $state(false);

	authStore.subscribe((value) => {
		user = value;
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
			// Get access token from API (uses authenticated fetch)
			const response = await apiPost<{ token: string }>('/api/video/token', {
				room: callId
			});

			if (response.error || !response.data) {
				throw new Error(response.error || 'Failed to get access token');
			}

			const data = response.data;

			// Import Twilio Video dynamically
			const Video = await import('twilio-video');

			// Connect to the room
			room = await Video.connect(data.token, {
				name: callId,
				audio: true,
				video: false
			});

			localParticipant = room.localParticipant;

			// Capture local audio track for waveform
			localParticipant.audioTracks.forEach((pub: LocalTrackPublication) => {
				if (pub.track) {
					localAudioTrack = (pub.track as any).mediaStreamTrack;
				}
			});

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

	// Attach a remote audio/video track: use SDK's attach() for audio playback,
	// and store the mediaStreamTrack for waveform visualization.
	function attachTrack(track: RemoteTrack, participantSid: string) {
		const entry = remoteTrackMap.get(participantSid) || { audio: null, video: null };
		if (track.kind === 'audio') {
			// Use SDK's attach() to create a proper <audio> element for playback
			const audioEl = (track as any).attach() as HTMLAudioElement;
			audioEl.setAttribute('data-participant', participantSid);
			document.body.appendChild(audioEl);
			entry.audio = (track as any).mediaStreamTrack || null;
		} else if (track.kind === 'video') {
			entry.video = (track as any).mediaStreamTrack || null;
		}
		remoteTrackMap = new Map(remoteTrackMap.set(participantSid, entry));
	}

	// Detach a remote audio/video track and clean up DOM elements.
	function detachTrack(track: RemoteTrack, participantSid: string) {
		const entry = remoteTrackMap.get(participantSid) || { audio: null, video: null };
		if (track.kind === 'audio') {
			// Remove SDK-created audio elements
			(track as any).detach().forEach((el: HTMLElement) => el.remove());
			entry.audio = null;
		} else if (track.kind === 'video') {
			entry.video = null;
		}
		remoteTrackMap = new Map(remoteTrackMap.set(participantSid, entry));
	}

	function handleParticipantConnected(participant: RemoteParticipant) {
		remoteParticipants = [...remoteParticipants, participant];

		// Handle existing tracks
		participant.tracks.forEach((publication: RemoteTrackPublication) => {
			if (publication.isSubscribed && publication.track) {
				attachTrack(publication.track as RemoteTrack, participant.sid);
			}
		});

		// Handle new tracks
		participant.on('trackSubscribed', (track: RemoteTrack) => {
			attachTrack(track, participant.sid);
		});

		participant.on('trackUnsubscribed', (track: RemoteTrack) => {
			detachTrack(track, participant.sid);
		});
	}

	function handleParticipantDisconnected(participant: RemoteParticipant) {
		// Detach all tracks for this participant
		participant.tracks.forEach((publication: RemoteTrackPublication) => {
			if (publication.track) {
				(publication.track as any).detach().forEach((el: HTMLElement) => el.remove());
			}
		});
		remoteParticipants = remoteParticipants.filter(p => p.sid !== participant.sid);
		remoteTrackMap.delete(participant.sid);
		remoteTrackMap = new Map(remoteTrackMap);
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
			localParticipant.videoTracks.forEach((publication) => {
				publication.track.stop();
				localParticipant?.unpublishTrack(publication.track);
			});
			localVideoTrack = null;
			videoEnabled = false;
		} else {
			const Video = await import('twilio-video');
			const videoTrack = await Video.createLocalVideoTrack();
			await localParticipant.publishTrack(videoTrack);
			localVideoTrack = videoTrack.mediaStreamTrack;
			videoEnabled = true;
		}
	}

	function leaveCall() {
		if (room) {
			room.disconnect();
		}
		goto(`/call/${callId}/summary`);
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
			<button
				onclick={() => showInviteModal = true}
				class="text-sm px-3 py-1 rounded bg-twilio-green-60 hover:bg-twilio-green-70 text-white"
			>
				Invite via Phone
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
					<ParticipantTile
						identity={user?.contact || 'You'}
						type="browser"
						isLocal={true}
						audioTrack={localAudioTrack}
						videoTrack={localVideoTrack}
						audioEnabled={audioEnabled}
						videoEnabled={videoEnabled}
						displayName={user?.displayName}
					/>
				{/if}

				<!-- Remote Participants -->
				{#each remoteParticipants as participant (participant.sid)}
					{@const tracks = remoteTrackMap.get(participant.sid)}
					<ParticipantTile
						identity={participant.identity}
						type={isPhoneIdentity(participant.identity) ? 'pstn' : 'browser'}
						audioTrack={tracks?.audio || null}
						videoTrack={tracks?.video || null}
						audioEnabled={true}
						videoEnabled={!!tracks?.video}
					/>
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

	<InvitePhoneModal
		roomName={callId}
		open={showInviteModal}
		onclose={() => showInviteModal = false}
	/>
</div>
