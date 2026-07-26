import '@testing-library/jest-dom';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import CalendarPage from './+page.svelte';

const mockGoto = vi.fn();
const mockReplaceState = vi.fn();
vi.mock('$app/navigation', () => ({
	goto: (path: string) => mockGoto(path),
	replaceState: (url: string, _state: unknown) => mockReplaceState(url)
}));

function mockResponse(body: unknown, status = 200) {
	return {
		ok: status >= 200 && status < 300,
		status,
		text: async () => JSON.stringify(body),
		json: async () => body
	};
}

describe('CalendarPage', () => {
	beforeEach(() => {
		vi.resetAllMocks();
		vi.stubGlobal('localStorage', {
			getItem: vi.fn(),
			setItem: vi.fn(),
			removeItem: vi.fn()
		});
	});

	it('redirects to onboarding start when no token is available', async () => {
		vi.stubGlobal('window', {
			...window,
			location: { href: 'http://localhost:5200/onboarding/calendar' },
			history: { replaceState: vi.fn() }
		});

		render(CalendarPage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(mockGoto).toHaveBeenCalledWith('/onboarding');
	});

	it('shows unavailable state when google oauth is not configured', async () => {
		vi.stubGlobal('localStorage', {
			getItem: vi.fn().mockReturnValue('test-token'),
			setItem: vi.fn(),
			removeItem: vi.fn()
		});
		vi.stubGlobal('window', {
			...window,
			location: { href: 'http://localhost:5200/onboarding/calendar?token=test-token' },
			history: { replaceState: vi.fn() }
		});
		vi.stubGlobal('fetch', vi.fn(async (url: string) => {
			if (url.includes('/onboarding/state')) {
				return mockResponse({
					has_profile: true,
					telegram_linked: true,
					calendar_connected: false,
					tasks: [],
					starter_templates: []
				});
			}
			return mockResponse({ available: false, connected: false });
		}));

		render(CalendarPage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(screen.getByText(/Google Calendar is not available right now/i)).toBeInTheDocument();
		expect(screen.getByRole('button', { name: /Continue to tasks/i })).toBeInTheDocument();
	});

	it('shows connected state when calendar is already connected', async () => {
		vi.stubGlobal('localStorage', {
			getItem: vi.fn().mockReturnValue('test-token'),
			setItem: vi.fn(),
			removeItem: vi.fn()
		});
		vi.stubGlobal('window', {
			...window,
			location: { href: 'http://localhost:5200/onboarding/calendar?token=test-token' },
			history: { replaceState: vi.fn() }
		});
		vi.stubGlobal('fetch', vi.fn(async (url: string) => {
			if (url.includes('/onboarding/state')) {
				return mockResponse({
					has_profile: true,
					telegram_linked: true,
					calendar_connected: true,
					tasks: [],
					starter_templates: []
				});
			}
			return mockResponse({ available: true, connected: true });
		}));

		render(CalendarPage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(screen.getByText(/Google Calendar is connected/i)).toBeInTheDocument();
		expect(screen.getByRole('button', { name: /Disconnect calendar/i })).toBeInTheDocument();
	});

	it('shows connect action when oauth is configured and calendar is not connected', async () => {
		vi.stubGlobal('localStorage', {
			getItem: vi.fn().mockReturnValue('test-token'),
			setItem: vi.fn(),
			removeItem: vi.fn()
		});
		vi.stubGlobal('window', {
			...window,
			location: { href: 'http://localhost:5200/onboarding/calendar?token=test-token' },
			history: { replaceState: vi.fn() }
		});
		vi.stubGlobal('fetch', vi.fn(async (url: string) => {
			if (url.includes('/onboarding/state')) {
				return mockResponse({
					has_profile: true,
					telegram_linked: true,
					calendar_connected: false,
					tasks: [],
					starter_templates: []
				});
			}
			return mockResponse({ available: true, connected: false, auth_url: 'https://accounts.google.test/oauth' });
		}));

		render(CalendarPage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(screen.getByRole('link', { name: /Google Calendar/i })).toHaveAttribute(
			'href',
			'https://accounts.google.test/oauth'
		);
		expect(screen.getByRole('button', { name: /Skip and continue/i })).toBeInTheDocument();
	});
});
