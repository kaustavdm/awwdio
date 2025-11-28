<script lang="ts">
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth';

	let user = $state<any>(null);

	authStore.subscribe((value) => {
		user = value;
	});

	async function startCall() {
		if (!user) {
			// Redirect to login with a flag to create call after login
			goto('/login?intent=create-call');
			return;
		}

		// Create a new call and redirect to setup
		try {
			// TODO: Call API to create a new call
			const callId = crypto.randomUUID(); // Temporary - will be replaced with API call
			goto(`/call/${callId}/setup`);
		} catch (error) {
			console.error('Failed to create call:', error);
		}
	}
</script>

<svelte:head>
	<title>Awwdio - Audio & Video Conversations</title>
</svelte:head>

<div class="flex flex-col items-center justify-center min-h-screen p-4">
	<div class="text-center mb-12">
		<h1 class="text-5xl font-bold mb-4 text-twilio-gray-100 dark:text-twilio-gray-0">Awwdio</h1>
		<p class="text-xl text-twilio-gray-60 dark:text-twilio-gray-40">
			Lightweight audio & video conversations
		</p>
	</div>

	<button
		onclick={startCall}
		class="group relative w-64 h-64 rounded-full bg-gradient-to-br from-twilio-blue-60 to-twilio-purple-60 hover:from-twilio-blue-70 hover:to-twilio-purple-70 transition-all duration-300 transform hover:scale-105 shadow-2xl"
	>
		<div class="flex flex-col items-center justify-center h-full">
			<svg
				xmlns="http://www.w3.org/2000/svg"
				class="h-20 w-20 text-white mb-4"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"
				/>
			</svg>
			<span class="text-2xl font-bold text-white">Start Call</span>
		</div>
	</button>

	{#if user}
		<div class="mt-8 text-center">
			<p class="text-sm text-twilio-gray-60 dark:text-twilio-gray-40">
				Logged in as <span class="font-semibold text-twilio-gray-100 dark:text-twilio-gray-0">{user.displayName || user.contact}</span>
			</p>
			<button
				onclick={() => authStore.logout()}
				class="mt-2 text-sm text-twilio-red-60 dark:text-twilio-red-30 hover:underline"
			>
				Logout
			</button>
		</div>
	{/if}
</div>
