import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import { authStore } from './stores/auth';

export interface ApiError {
	error: string;
}

export interface ApiResponse<T> {
	data?: T;
	error?: string;
	status: number;
}

/**
 * Authenticated fetch wrapper that:
 * - Adds Authorization: Bearer token header
 * - Handles 401 responses by redirecting to login
 * - Returns typed response data
 */
export async function apiFetch<T>(
	url: string,
	options: RequestInit = {}
): Promise<ApiResponse<T>> {
	const token = authStore.getToken();

	const headers: HeadersInit = {
		'Content-Type': 'application/json',
		...(options.headers || {})
	};

	if (token) {
		(headers as Record<string, string>)['Authorization'] = `Bearer ${token}`;
	}

	try {
		const response = await fetch(url, {
			...options,
			headers
		});

		// Handle 401 - redirect to login
		if (response.status === 401) {
			if (browser) {
				authStore.logout();
				goto('/login');
			}
			return {
				error: 'Authentication required',
				status: 401
			};
		}

		// Parse response
		const data = await response.json();

		if (!response.ok) {
			return {
				error: (data as ApiError).error || 'Request failed',
				status: response.status
			};
		}

		return {
			data: data as T,
			status: response.status
		};
	} catch (err) {
		return {
			error: err instanceof Error ? err.message : 'Network error',
			status: 0
		};
	}
}

/**
 * GET request with authentication
 */
export async function apiGet<T>(url: string): Promise<ApiResponse<T>> {
	return apiFetch<T>(url, { method: 'GET' });
}

/**
 * POST request with authentication
 */
export async function apiPost<T>(url: string, body: unknown): Promise<ApiResponse<T>> {
	return apiFetch<T>(url, {
		method: 'POST',
		body: JSON.stringify(body)
	});
}
