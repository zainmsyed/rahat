import '@testing-library/jest-dom';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import CallbackPage from './+page.svelte';

const mockGoto = vi.fn();
const mockReplaceState = vi.fn();
const getStoredOnboardingToken = vi.fn();
const clearStoredOnboardingToken = vi.fn();
const syncTokenInUrl = vi.fn();

vi.mock('$app/navigation', () => ({
	goto: (path: string) => mockGoto(path),
	replaceState: (url: string, _state: unknown) => mockReplaceState(url)
}));
vi.mock('$lib/api/onboarding', () => ({
	apiBaseUrl: 'http://localhost:8080',
	buildOnboardingSteps: () => [
		{ id: 0, title: 'Start', required: true, description: '', complete: true },
		{ id: 1, title: 'Profile', required: true, description: '', complete: true },
		{ id: 2, title: 'Telegram', required: false, description: '', complete: false },
		{ id: 3, title: 'Calendar', required: false, description: '', complete: false },
		{ id: 4, title: 'Tasks', required: true, description: '', complete: false },
		{ id: 5, title: 'Review', required: true, description: '', complete: false }
	],
	clearStoredOnboardingToken: () => clearStoredOnboardingToken(),
	getStoredOnboardingToken: () => getStoredOnboardingToken(),
	syncTokenInUrl: (token: string) => syncTokenInUrl(token)
}));

describe('CalendarCallbackPage', () => {
	beforeEach(() => {
		vi.resetAllMocks();
		getStoredOnboardingToken.mockReturnValue('stored-token');
		vi.stubGlobal('localStorage', {
			getItem: vi.fn(),
			setItem: vi.fn(),
			removeItem: vi.fn()
		});
	});

	function setUrl(href: string) {
		vi.stubGlobal('window', {
			...window,
			location: { href },
			history: { replaceState: vi.fn() }
		});
	}

	function mockFetch(response: { ok: boolean; text?: () => Promise<string> }) {
		vi.stubGlobal('fetch', vi.fn(() => Promise.resolve(response)));
	}

	it('redirects to onboarding start when no token is available', async () => {
		getStoredOnboardingToken.mockReturnValue('');
		setUrl('http://localhost:5200/onboarding/calendar/callback?state=s&code=c');

		render(CallbackPage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(mockGoto).toHaveBeenCalledWith('/onboarding');
	});

	it('shows an error when Google returns an error parameter', async () => {
		setUrl('http://localhost:5200/onboarding/calendar/callback?error=access_denied');

		render(CallbackPage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(screen.getByText(/Google returned an error/i)).toBeInTheDocument();
		expect(screen.getByRole('button', { name: /Back to calendar step/i })).toBeInTheDocument();
		expect(screen.getByRole('button', { name: /Restart onboarding/i })).toBeInTheDocument();
	});

	it('shows an error when state or code is missing', async () => {
		setUrl('http://localhost:5200/onboarding/calendar/callback?state=s');

		render(CallbackPage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(screen.getByText(/missing information needed to connect/i)).toBeInTheDocument();
	});

	it('redirects to calendar step on successful connection', async () => {
		setUrl('http://localhost:5200/onboarding/calendar/callback?state=s&code=c');
		mockFetch({ ok: true, text: async () => '{}' });

		render(CallbackPage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(mockGoto).toHaveBeenCalledWith('/onboarding/calendar');
	});

	it('shows an error when the connect request fails', async () => {
		setUrl('http://localhost:5200/onboarding/calendar/callback?state=s&code=c');
		mockFetch({ ok: false, text: async () => 'invalid grant' });

		render(CallbackPage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(screen.getByText(/invalid grant/i)).toBeInTheDocument();
	});
});
