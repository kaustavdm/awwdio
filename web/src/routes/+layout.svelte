<script lang="ts">
	import '../app.css';
	import { browser } from '$app/environment';
	import { onMount } from 'svelte';

	let darkMode = $state(false);

	onMount(() => {
		// Check for saved theme preference or default to system preference
		const savedTheme = localStorage.getItem('theme');
		if (savedTheme) {
			darkMode = savedTheme === 'dark';
		} else {
			darkMode = window.matchMedia('(prefers-color-scheme: dark)').matches;
		}
		updateTheme();
	});

	function updateTheme() {
		if (browser) {
			if (darkMode) {
				document.documentElement.classList.add('dark');
			} else {
				document.documentElement.classList.remove('dark');
			}
			localStorage.setItem('theme', darkMode ? 'dark' : 'light');
		}
	}

	function toggleTheme() {
		darkMode = !darkMode;
		updateTheme();
	}

	let { children } = $props();
</script>

<div class="min-h-screen">
	<!-- Theme toggle button -->
	<button
		onclick={toggleTheme}
		class="fixed top-4 right-4 p-2 rounded-lg bg-twilio-gray-20 dark:bg-twilio-gray-90 hover:bg-twilio-gray-30 dark:hover:bg-twilio-gray-80 transition-colors z-50 border border-twilio-gray-30 dark:border-twilio-gray-70"
		aria-label="Toggle theme"
	>
		{#if darkMode}
			<svg
				xmlns="http://www.w3.org/2000/svg"
				class="h-6 w-6"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
				/>
			</svg>
		{:else}
			<svg
				xmlns="http://www.w3.org/2000/svg"
				class="h-6 w-6"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
				/>
			</svg>
		{/if}
	</button>

	{@render children()}
</div>
