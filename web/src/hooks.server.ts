import { redirect, type Handle } from '@sveltejs/kit';
import { apiBaseUrl } from '$lib/api/config';
const authOptionalPrefixes = ['/login', '/lookahead', '/onboarding'];
const protectedPrefixes = ['/', '/tasks'];

async function loadCurrentUser(cookieHeader: string | null) {
	if (!cookieHeader || !cookieHeader.includes('rahat_session=')) {
		return null;
	}
	try {
		const response = await fetch(`${apiBaseUrl}/auth/session`, {
			headers: {
				cookie: cookieHeader
			}
		});
		if (!response.ok) {
			return null;
		}
		const body = (await response.json()) as { authenticated: boolean; user?: App.SessionUser };
		if (!body.authenticated || !body.user) {
			return null;
		}
		return body.user;
	} catch {
		return null;
	}
}

export const handle: Handle = async ({ event, resolve }) => {
	event.locals.sessionUser = await loadCurrentUser(event.request.headers.get('cookie'));
	const pathname = event.url.pathname;
	if (!event.locals.sessionUser && protectedPrefixes.some((prefix) => pathname === prefix || (prefix !== '/' && pathname.startsWith(prefix + '/')))) {
		if (!authOptionalPrefixes.some((prefix) => pathname === prefix || pathname.startsWith(prefix + '/'))) {
			throw redirect(303, '/login');
		}
		if (pathname === '/') {
			throw redirect(303, '/login');
		}
	}
	return resolve(event);
};
