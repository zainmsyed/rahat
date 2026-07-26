import '@testing-library/jest-dom';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import TelegramPage from './+page.svelte';

const mockGoto = vi.fn();
const getState = vi.fn();
const getTelegramStatus = vi.fn();
const skipTelegram = vi.fn();
const getStoredOnboardingToken = vi.fn();
const readTokenFromUrl = vi.fn();
const setStoredOnboardingToken = vi.fn();
const syncTokenInUrl = vi.fn();
const clearStoredOnboardingToken = vi.fn();

vi.mock('$app/navigation', () => ({ goto: (path: string) => mockGoto(path) }));
vi.mock('qrcode', () => ({
	default: {
		toDataURL: vi.fn(() => Promise.resolve('data:image/png;base64,abc'))
	}
}));
vi.mock('$lib/api/onboarding', () => ({
	buildOnboardingSteps: () => [
		{ id: 0, title: 'Start', required: true, description: '', complete: true },
		{ id: 1, title: 'Profile', required: true, description: '', complete: true },
		{ id: 2, title: 'Telegram', required: false, description: '', complete: false },
		{ id: 3, title: 'Calendar', required: false, description: '', complete: false },
		{ id: 4, title: 'Tasks', required: true, description: '', complete: false },
		{ id: 5, title: 'Review', required: true, description: '', complete: false }
	],
	clearStoredOnboardingToken: () => clearStoredOnboardingToken(),
	formatStepLabel: (step: { id: number; required: boolean }) =>
		`Step ${step.id + 1} · ${step.required ? 'Required' : 'Recommended'}`,
	getState: (token: string) => getState(token),
	getStoredOnboardingToken: () => getStoredOnboardingToken(),
	getTelegramStatus: (token: string) => getTelegramStatus(token),
	readTokenFromUrl: () => readTokenFromUrl(),
	setStoredOnboardingToken: (token: string) => setStoredOnboardingToken(token),
	skipTelegram: (token: string) => skipTelegram(token),
	syncTokenInUrl: (token: string) => syncTokenInUrl(token)
}));

describe('TelegramPage', () => {
	beforeEach(() => {
		vi.resetAllMocks();
		getStoredOnboardingToken.mockReturnValue('stored-token');
		readTokenFromUrl.mockReturnValue('url-token');
		vi.stubGlobal('localStorage', {
			getItem: vi.fn(),
			setItem: vi.fn(),
			removeItem: vi.fn()
		});
	});

	it('redirects to onboarding start when no token is available', async () => {
		getStoredOnboardingToken.mockReturnValue('');
		readTokenFromUrl.mockReturnValue('');

		render(TelegramPage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(mockGoto).toHaveBeenCalledWith('/onboarding');
	});

	it('shows unavailable state and continue action', async () => {
		getState.mockResolvedValue({
			has_profile: true,
			telegram_linked: false,
			calendar_connected: false,
			tasks: [],
			starter_templates: []
		});
		getTelegramStatus.mockResolvedValue({ available: false, linked: false });

		render(TelegramPage);
		await waitFor(() =>
			expect(screen.getByText(/Telegram is not configured right now/i)).toBeInTheDocument()
		);

		expect(screen.getByRole('button', { name: /Continue with email only/i })).toBeInTheDocument();
	});

	it('shows connected state and continue action', async () => {
		getState.mockResolvedValue({
			has_profile: true,
			telegram_linked: true,
			calendar_connected: false,
			tasks: [],
			starter_templates: []
		});
		getTelegramStatus.mockResolvedValue({
			available: true,
			linked: true,
			bot_username: 'rahat_bot'
		});

		render(TelegramPage);
		await waitFor(() => expect(screen.getByText('Connected')).toBeInTheDocument());

		expect(screen.getByRole('button', { name: /Continue to calendar/i })).toBeInTheDocument();
	});

	it('shows connect tile with deep link when not connected', async () => {
		getState.mockResolvedValue({
			has_profile: true,
			telegram_linked: false,
			calendar_connected: false,
			tasks: [],
			starter_templates: []
		});
		getTelegramStatus.mockResolvedValue({
			available: true,
			linked: false,
			bot_username: 'rahat_bot',
			deep_link: 'https://t.me/rahat_bot?start=abc',
			code: 'ABC123'
		});

		render(TelegramPage);
		await waitFor(() => expect(screen.getByText('Not connected')).toBeInTheDocument());

		const tile = screen.getByRole('link', { name: /Telegram/i });
		expect(tile).toHaveAttribute('href', 'https://t.me/rahat_bot?start=abc');
		expect(screen.getByText('ABC123')).toBeInTheDocument();
	});
});
