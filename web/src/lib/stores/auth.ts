import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export interface User {
	channel: 'email' | 'sms';
	contact: string; // Email address or phone number
	displayName?: string;
	token?: string;
}

function createAuthStore() {
	const { subscribe, set, update } = writable<User | null>(null);

	// Load user from localStorage on initialization
	if (browser) {
		const storedUser = localStorage.getItem('user');
		if (storedUser) {
			try {
				set(JSON.parse(storedUser));
			} catch (e) {
				localStorage.removeItem('user');
			}
		}
	}

	return {
		subscribe,
		login: (user: User) => {
			set(user);
			if (browser) {
				localStorage.setItem('user', JSON.stringify(user));
				// Store token separately for easier access
				if (user.token) {
					localStorage.setItem('token', user.token);
				}
			}
		},
		logout: () => {
			set(null);
			if (browser) {
				localStorage.removeItem('user');
				localStorage.removeItem('token');
			}
		},
		updateDisplayName: (displayName: string) => {
			update((user) => {
				if (user) {
					const updatedUser = { ...user, displayName };
					if (browser) {
						localStorage.setItem('user', JSON.stringify(updatedUser));
					}
					return updatedUser;
				}
				return user;
			});
		}
	};
}

export const authStore = createAuthStore();
