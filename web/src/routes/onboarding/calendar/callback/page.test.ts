import '@testing-library/jest-dom';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import CallbackPage from './+page.svelte';

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

describe('CalendarCallbackPage', () => {
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
			location: { href: 'http://localhost:5200/onboarding/calendar/callback?state=abc&code=123' },
			history: { replaceState: vi.fn() }
		});

		render(CallbackPage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(mockGoto).toHaveBeenCalledWith('/onboarding');
	});

	it('shows an error when Google returns an error parameter', async () => {
		vi.stubGlobal('localStorage', {
			getItem: vi.fn().mockReturnValue('test-token'),
			setItem: vi.fn(),
			removeItem: vi.fn()
		});
		vi.stubGlobal('window', {
			...window,
			location: { href: 'http://localhost:5200/onboarding/calendar/callback?error=access_denied' },
			history: { replaceState: vi.fn() }
		});

		render(CallbackPage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(screen.getByText(/Calendar connection did not finish/i)).toBeInTheDocument();
		expect(screen.getByText(/access_denied/i)).toBeInTheDocument();
	});

	it('exchanges the code and redirects to the calendar step on success', async () => {
		const fetchMock = vi.fn(async (url: string) => {
			if (url.includes('/calendar/google/connect')) {
				return mockResponse({ provider: 'google' });
			}
			return mockResponse({});
		});
		vi.stubGlobal('localStorage', {
			getItem: vi.fn().mockReturnValue('test-token'),
			setItem: vi.fn(),
			removeItem: vi.fn()
		});
		vi.stubGlobal('window', {
			...window,
			location: {
				href: 'http://localhost:5200/onboarding/calendar/callback?state=abc123&code=secret'
			},
			history: { replaceState: vi.fn() }
		});
		vi.stubGlobal('fetch', fetchMock);

		render(CallbackPage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(fetchMock).toHaveBeenCalledWith(
			expect.stringContaining('/calendar/google/connect?state=abc123&code=secret'),
			expect.objectContaining({ method: 'POST' })
		);
		expect(mockGoto).toHaveBeenCalledWith('/onboarding/calendar');
	});
});
