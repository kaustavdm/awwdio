<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { authStore } from '$lib/stores/auth';

	let step = $state<'contact' | 'otp' | 'displayName'>('contact');
	let channel = $state<'email' | 'sms'>('email');
	let contact = $state('');
	let otp = $state('');
	let displayName = $state('');
	let token = $state('');
	let error = $state('');
	let loading = $state(false);

	let intent = $state<string | null>(null);

	$effect(() => {
		intent = $page.url.searchParams.get('intent');
	});

	async function sendOTP() {
		// Validate input based on channel
		if (channel === 'email') {
			if (!contact || !contact.includes('@')) {
				error = 'Please enter a valid email address';
				return;
			}
		} else if (channel === 'sms') {
			if (!contact || contact.length < 10) {
				error = 'Please enter a valid phone number';
				return;
			}
		}

		loading = true;
		error = '';

		try {
			const response = await fetch('/api/auth/send-otp', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ channel, to: contact })
			});

			if (!response.ok) {
				throw new Error('Failed to send OTP');
			}

			step = 'otp';
		} catch (e) {
			error = 'Failed to send verification code. Please try again.';
			console.error(e);
		} finally {
			loading = false;
		}
	}

	async function verifyOTP() {
		if (!otp || otp.length < 6) {
			error = 'Please enter a valid verification code';
			return;
		}

		loading = true;
		error = '';

		try {
			const response = await fetch('/api/auth/verify-otp', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ channel, to: contact, otp })
			});

			if (!response.ok) {
				throw new Error('Invalid verification code');
			}

			const data = await response.json();
			token = data.token;

			step = 'displayName';
		} catch (e) {
			error = 'Invalid verification code. Please try again.';
			console.error(e);
		} finally {
			loading = false;
		}
	}

	function completeLogin() {
		// Save user to auth store
		authStore.login({
			channel,
			contact,
			displayName: displayName || undefined,
			token
		});

		// Redirect based on intent
		if (intent === 'create-call') {
			const callId = crypto.randomUUID(); // Temporary - will be replaced with API call
			goto(`/call/${callId}/setup`);
		} else {
			goto('/');
		}
	}
</script>

<svelte:head>
	<title>Login - Awwdio</title>
</svelte:head>

<div class="flex items-center justify-center min-h-screen p-4">
	<div class="w-full max-w-md">
		<div class="text-center mb-8">
			<h1 class="text-3xl font-bold mb-2 text-twilio-gray-100 dark:text-twilio-gray-0">Welcome to Awwdio</h1>
			<p class="text-twilio-gray-60 dark:text-twilio-gray-40">Sign in to continue</p>
		</div>

		<div class="bg-twilio-gray-0 dark:bg-twilio-gray-90 rounded-lg shadow-xl p-8 border border-twilio-gray-20 dark:border-twilio-gray-80">
			{#if error}
				<div class="mb-4 p-3 bg-twilio-red-10 dark:bg-twilio-red-100 text-twilio-red-70 dark:text-twilio-red-30 rounded border border-twilio-red-30">
					{error}
				</div>
			{/if}

			{#if step === 'contact'}
				<form onsubmit={(e) => { e.preventDefault(); sendOTP(); }}>
					<!-- Channel Selection -->
					<div class="mb-4">
						<label class="block mb-2 text-sm font-medium">Verification Method</label>
						<div class="flex gap-2">
							<button
								type="button"
								onclick={() => { channel = 'email'; contact = ''; error = ''; }}
								class="flex-1 py-2 px-4 rounded-lg border transition-colors {channel === 'email'
									? 'bg-twilio-blue-60 text-white border-twilio-blue-60'
									: 'bg-twilio-gray-0 dark:bg-twilio-gray-80 text-twilio-gray-100 dark:text-twilio-gray-0 border-twilio-gray-30 dark:border-twilio-gray-70 hover:border-twilio-blue-60'}"
							>
								Email
							</button>
							<button
								type="button"
								onclick={() => { channel = 'sms'; contact = ''; error = ''; }}
								class="flex-1 py-2 px-4 rounded-lg border transition-colors {channel === 'sms'
									? 'bg-twilio-blue-60 text-white border-twilio-blue-60'
									: 'bg-twilio-gray-0 dark:bg-twilio-gray-80 text-twilio-gray-100 dark:text-twilio-gray-0 border-twilio-gray-30 dark:border-twilio-gray-70 hover:border-twilio-blue-60'}"
							>
								SMS
							</button>
						</div>
					</div>

					<!-- Contact Input -->
					{#if channel === 'email'}
						<label class="block mb-2 text-sm font-medium" for="contact">Email Address</label>
						<input
							id="contact"
							type="email"
							bind:value={contact}
							placeholder="your@email.com"
							class="w-full px-4 py-3 rounded-lg border border-twilio-gray-30 dark:border-twilio-gray-70 bg-twilio-gray-0 dark:bg-twilio-gray-80 text-twilio-gray-100 dark:text-twilio-gray-0 focus:outline-none focus:ring-2 focus:ring-twilio-blue-60 mb-4"
							required
						/>
					{:else}
						<label class="block mb-2 text-sm font-medium" for="contact">Phone Number</label>
						<input
							id="contact"
							type="tel"
							bind:value={contact}
							placeholder="+1234567890"
							class="w-full px-4 py-3 rounded-lg border border-twilio-gray-30 dark:border-twilio-gray-70 bg-twilio-gray-0 dark:bg-twilio-gray-80 text-twilio-gray-100 dark:text-twilio-gray-0 focus:outline-none focus:ring-2 focus:ring-twilio-blue-60 mb-4"
							required
						/>
						<p class="text-xs text-twilio-gray-60 dark:text-twilio-gray-40 mb-4">
							Include country code (e.g., +1 for US)
						</p>
					{/if}

					<button
						type="submit"
						disabled={loading}
						class="w-full py-3 bg-twilio-blue-60 hover:bg-twilio-blue-70 disabled:bg-twilio-gray-40 disabled:cursor-not-allowed text-white font-semibold rounded-lg transition-colors"
					>
						{loading ? 'Sending...' : 'Send Verification Code'}
					</button>
				</form>
			{:else if step === 'otp'}
				<div class="mb-4">
					<p class="text-sm text-twilio-gray-60 dark:text-twilio-gray-40">
						We've sent a verification code {channel === 'email' ? 'to' : 'via SMS to'} <strong class="text-twilio-gray-100 dark:text-twilio-gray-0">{contact}</strong>
					</p>
				</div>

				<form onsubmit={(e) => { e.preventDefault(); verifyOTP(); }}>
					<label class="block mb-2 text-sm font-medium" for="otp">Verification Code</label>
					<input
						id="otp"
						type="text"
						bind:value={otp}
						placeholder="000000"
						maxlength="6"
						class="w-full px-4 py-3 rounded-lg border border-twilio-gray-30 dark:border-twilio-gray-70 bg-twilio-gray-0 dark:bg-twilio-gray-80 text-twilio-gray-100 dark:text-twilio-gray-0 focus:outline-none focus:ring-2 focus:ring-twilio-blue-60 mb-4 text-center text-2xl tracking-widest"
						required
					/>

					<button
						type="submit"
						disabled={loading}
						class="w-full py-3 bg-twilio-blue-60 hover:bg-twilio-blue-70 disabled:bg-twilio-gray-40 disabled:cursor-not-allowed text-white font-semibold rounded-lg transition-colors mb-2"
					>
						{loading ? 'Verifying...' : 'Verify Code'}
					</button>

					<button
						type="button"
						onclick={() => { step = 'contact'; otp = ''; error = ''; }}
						class="w-full py-2 text-sm text-twilio-gray-60 dark:text-twilio-gray-40 hover:text-twilio-gray-100 dark:hover:text-twilio-gray-0"
					>
						Use a different {channel === 'email' ? 'email' : 'phone number'}
					</button>
				</form>
			{:else if step === 'displayName'}
				<div class="mb-4">
					<p class="text-sm text-twilio-green-60 dark:text-twilio-green-30 mb-4">
						Verification successful!
					</p>
				</div>

				<form onsubmit={(e) => { e.preventDefault(); completeLogin(); }}>
					<label class="block mb-2 text-sm font-medium" for="displayName">
						What should we call you? <span class="text-twilio-gray-50">(optional)</span>
					</label>
					<input
						id="displayName"
						type="text"
						bind:value={displayName}
						placeholder="Your name"
						class="w-full px-4 py-3 rounded-lg border border-twilio-gray-30 dark:border-twilio-gray-70 bg-twilio-gray-0 dark:bg-twilio-gray-80 text-twilio-gray-100 dark:text-twilio-gray-0 focus:outline-none focus:ring-2 focus:ring-twilio-blue-60 mb-4"
					/>

					<button
						type="submit"
						class="w-full py-3 bg-twilio-blue-60 hover:bg-twilio-blue-70 text-white font-semibold rounded-lg transition-colors"
					>
						Continue
					</button>
				</form>
			{/if}
		</div>

		<div class="text-center mt-4">
			<button onclick={() => goto('/')} class="text-sm text-twilio-gray-60 dark:text-twilio-gray-40 hover:text-twilio-blue-60 dark:hover:text-twilio-blue-30">
				← Back to home
			</button>
		</div>
	</div>
</div>
