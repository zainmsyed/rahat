import '@testing-library/jest-dom';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import OnboardingPage from './+page.svelte';

const mockGoto = vi.fn();
const createSession = vi.fn();
const getState = vi.fn();
const setStoredOnboardingToken = vi.fn();
const syncTokenInUrl = vi.fn();
const clearStoredOnboardingToken = vi.fn();

vi.mock('$app/navigation', () => ({ goto: (path: string) => mockGoto(path) }));
vi.mock('$lib/api/onboarding', () => ({
	buildOnboardingSteps: () => [
		{ id: 0, title: 'Start', required: true, description: '', complete: false },
		{ id: 1, title: 'Profile', required: true, description: '', complete: false },
		{ id: 2, title: 'Telegram', required: false, description: '', complete: false },
		{ id: 3, title: 'Calendar', required: false, description: '', complete: false },
		{ id: 4, title: 'Tasks', required: true, description: '', complete: false },
		{ id: 5, title: 'Review', required: true, description: '', complete: false }
	],
	createSession: (invite: string) => createSession(invite),
	getState: (token: string) => getState(token),
	getStoredOnboardingToken: () => '',
	nextOnboardingPath: () => '/onboarding/profile',
	readInviteCodeFromUrl: () => new URL(window.location.href).searchParams.get('invite') ?? '',
	readTokenFromUrl: () => '',
	setStoredOnboardingToken: (token: string) => setStoredOnboardingToken(token),
	syncTokenInUrl: (token: string) => syncTokenInUrl(token),
	clearStoredOnboardingToken: () => clearStoredOnboardingToken()
}));

describe('OnboardingPage', () => {
	beforeEach(() => {
		vi.resetAllMocks();
		vi.stubGlobal('localStorage', {
			getItem: vi.fn(),
			setItem: vi.fn(),
			removeItem: vi.fn()
		});
	});

	it('renders the invite form and shows a validation error for an empty code', async () => {
		vi.stubGlobal('window', {
			...window,
			location: { href: 'http://localhost:5200/onboarding' },
			history: { replaceState: vi.fn() }
		});

		const { container } = render(OnboardingPage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(screen.getByLabelText('Invite code *')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Start onboarding' })).toBeInTheDocument();

		const form = container.querySelector('form');
		if (!form) throw new Error('form not found');
		await fireEvent.submit(form);
		expect(screen.getByRole('alert')).toHaveTextContent(
			'Please enter your invite code to begin.'
		);
	});

	it('automatically starts setup when an invite is present in the URL', async () => {
		vi.stubGlobal('window', {
			...window,
			location: { href: 'http://localhost:5200/onboarding?invite=rahat-beta' },
			history: { replaceState: vi.fn() }
		});
		createSession.mockResolvedValue({ token: 'onboarding-token' });
		getState.mockResolvedValue({ has_profile: false, telegram_linked: false, calendar_connected: false, tasks: [], starter_templates: [] });

		render(OnboardingPage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(createSession).toHaveBeenCalledWith('rahat-beta');
		expect(setStoredOnboardingToken).toHaveBeenCalledWith('onboarding-token');
		expect(syncTokenInUrl).toHaveBeenCalledWith('onboarding-token');
		expect(mockGoto).toHaveBeenCalledWith('/onboarding/profile');
	});
});
