import { apiBaseUrl } from './config';

export type SessionUser = {
	id: string;
	display_name: string;
	timezone: string;
	daily_time_budget_minutes: number;
	email: string;
};

export type CurrentSession = {
	authenticated: boolean;
	user?: SessionUser;
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
	const response = await fetch(`${apiBaseUrl}${path}`, {
		credentials: 'include',
		headers: {
			'Content-Type': 'application/json',
			...(init?.headers ?? {})
		},
		...init
	});
	if (!response.ok) {
		throw new Error(await response.text());
	}
	if (response.status === 204) {
		return undefined as T;
	}
	return (await response.json()) as T;
}

export function exchangeAccessLink(token: string) {
	return request<CurrentSession>('/auth/access-link/exchange', {
		method: 'POST',
		body: JSON.stringify({ token })
	});
}

export function getCurrentSession() {
	return request<CurrentSession>('/auth/session');
}

export function logout() {
	return request<void>('/auth/logout', { method: 'POST', body: JSON.stringify({}) });
}
