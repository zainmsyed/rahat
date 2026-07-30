import { browser } from '$app/environment';
import { redirect } from '@sveltejs/kit';
import { getCurrentSession } from '$lib/api/auth';
import type { LayoutLoad } from './$types';

const protectedPrefixes = ['/tasks'];

function isProtected(pathname: string) {
	return protectedPrefixes.some(
		(prefix) => pathname === prefix || pathname.startsWith(prefix + '/')
	);
}

export const load: LayoutLoad = async ({ url }) => {
	if (!browser) {
		return {};
	}

	if (!isProtected(url.pathname)) {
		return {};
	}

	const session = await getCurrentSession().catch(() => ({ authenticated: false }));
	if (!session.authenticated) {
		redirect(303, '/login');
	}

	return {};
};
