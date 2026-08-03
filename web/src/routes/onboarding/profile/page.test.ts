import '@testing-library/jest-dom';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import ProfilePage from './+page.svelte';

const mockGoto = vi.fn();
const getState = vi.fn();
const saveProfile = vi.fn();
const getStoredOnboardingToken = vi.fn();
const readTokenFromUrl = vi.fn();
const setStoredOnboardingToken = vi.fn();
const syncTokenInUrl = vi.fn();
const clearStoredOnboardingToken = vi.fn();

vi.mock('$app/navigation', () => ({ goto: (path: string) => mockGoto(path) }));
vi.mock('$lib/api/onboarding', () => ({
	buildOnboardingSteps: () => [
		{ id: 0, title: 'Start', required: true, description: '', complete: true },
		{ id: 1, title: 'Profile', required: true, description: '', complete: false },
		{ id: 2, title: 'Telegram', required: false, description: '', complete: false },
		{ id: 3, title: 'Calendar', required: false, description: '', complete: false },
		{ id: 4, title: 'Tasks', required: true, description: '', complete: false },
		{ id: 5, title: 'Review', required: true, description: '', complete: false }
	],
	clearStoredOnboardingToken: () => clearStoredOnboardingToken(),
	getState: (token: string) => getState(token),
	getStoredOnboardingToken: () => getStoredOnboardingToken(),
	readTokenFromUrl: () => readTokenFromUrl(),
	saveProfile: (token: string, profile: unknown) => saveProfile(token, profile),
	setStoredOnboardingToken: (token: string) => setStoredOnboardingToken(token),
	syncTokenInUrl: (token: string) => syncTokenInUrl(token)
}));

describe('ProfilePage', () => {
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

		render(ProfilePage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(mockGoto).toHaveBeenCalledWith('/onboarding');
	});

	it('renders the profile form and saves the profile', async () => {
		getState.mockResolvedValue({
			has_profile: false,
			telegram_linked: false,
			calendar_connected: false,
			tasks: [],
			starter_templates: []
		});
		saveProfile.mockResolvedValue({
			display_name: 'Alex',
			timezone: 'America/New_York',
			daily_time_budget_minutes: 90,
			email: 'alex@example.com'
		});

		render(ProfilePage);
		await waitFor(() => expect(screen.getByLabelText('Name *')).toBeInTheDocument());

		const nameInput = screen.getByLabelText('Name *');
		const timezoneInput = screen.getByLabelText('Timezone *');
		const emailInput = screen.getByLabelText('Email for recaps (optional)');
		const slider = screen.getByRole('slider');

		await fireEvent.input(nameInput, { target: { value: 'Alex' } });
		await fireEvent.input(timezoneInput, { target: { value: 'America/New_York' } });
		await fireEvent.input(emailInput, { target: { value: 'alex@example.com' } });
		await fireEvent.input(slider, { target: { value: '90' } });

		const summary = screen.getByRole('status');
		expect(summary).toHaveTextContent('90');
		expect(summary).toHaveTextContent('minutes per day');

		const saveButton = screen.getByRole('button', { name: /Save and continue/i });
		await fireEvent.click(saveButton);

		await waitFor(() =>
			expect(saveProfile).toHaveBeenCalledWith('url-token', {
				display_name: 'Alex',
				timezone: 'America/New_York',
				daily_time_budget_minutes: 90,
				email: 'alex@example.com'
			})
		);
		expect(mockGoto).toHaveBeenCalledWith('/onboarding/telegram');
	});

	it('pre-fills the form from an existing profile', async () => {
		getState.mockResolvedValue({
			has_profile: true,
			telegram_linked: false,
			calendar_connected: false,
			tasks: [],
			starter_templates: [],
			user: {
				display_name: 'Jordan',
				timezone: 'Europe/London',
				daily_time_budget_minutes: 120,
				email: 'jordan@example.com'
			}
		});

		render(ProfilePage);
		await waitFor(() => expect(screen.getByLabelText('Name *')).toHaveValue('Jordan'));

		expect(screen.getByLabelText('Timezone *')).toHaveValue('Europe/London');
		expect(screen.getByLabelText('Email for recaps (optional)')).toHaveValue(
			'jordan@example.com'
		);
		const summary = screen.getByRole('status');
		expect(summary).toHaveTextContent('120');
		expect(summary).toHaveTextContent('minutes per day');
	});

	it('renders proportional budget tick labels aligned to the slider track', async () => {
		getState.mockResolvedValue({
			has_profile: false,
			telegram_linked: false,
			calendar_connected: false,
			tasks: [],
			starter_templates: []
		});

		render(ProfilePage);
		await waitFor(() => expect(screen.getByLabelText('Name *')).toBeInTheDocument());

		const slider = screen.getByRole('slider');
		const budgetInput = screen.getByRole('spinbutton', { name: 'Daily task-time budget value' });
		const tickContainer = slider.parentElement?.querySelector('.slider-ticks');
		expect(tickContainer).toHaveClass('slider-ticks');

		const ticks = [15, 60, 120, 240, 480];
		const expectedPositions = [0, 9.6774193548, 22.5806451613, 48.3870967742, 100];
		ticks.forEach((tick, index) => {
			const label = tickContainer?.querySelector(`[data-budget-tick="${tick}"]`);
			expect(label).toHaveTextContent(String(tick));
			expect(Number.parseFloat(label?.getAttribute('style')?.match(/[\d.]+/)?.[0] ?? 'NaN')).toBeCloseTo(
				expectedPositions[index],
				8
			);
		});
		expect(tickContainer?.querySelector('[data-budget-tick="15"]')).toHaveClass('first');
		expect(tickContainer?.querySelector('[data-budget-tick="480"]')).toHaveClass('last');

		for (const tick of ticks) {
			await fireEvent.input(slider, { target: { value: String(tick) } });
			expect(slider).toHaveValue(String(tick));
			expect(budgetInput).toHaveValue(tick);
			expect(screen.getByRole('status')).toHaveTextContent(String(tick));
		}

		await fireEvent.input(budgetInput, { target: { value: '73' } });
		expect(slider).toHaveValue('73');
		expect(screen.getByRole('status')).toHaveTextContent('73');

		await fireEvent.input(budgetInput, { target: { value: '500' } });
		expect(slider).toHaveValue('73');
		expect(screen.getByRole('status')).toHaveTextContent('73');
		await fireEvent.change(budgetInput, { target: { value: '500' } });
		expect(budgetInput).toHaveValue(480);
		expect(slider).toHaveValue('480');
		expect(screen.getByRole('status')).toHaveTextContent('480');

		await fireEvent.input(budgetInput, { target: { value: '73.5' } });
		expect(slider).toHaveValue('480');
		await fireEvent.change(budgetInput, { target: { value: '73.5' } });
		expect(budgetInput).toHaveValue(74);
		expect(slider).toHaveValue('74');
	});

	it('shows validation errors and does not submit when required fields are invalid', async () => {
		getState.mockResolvedValue({
			has_profile: false,
			telegram_linked: false,
			calendar_connected: false,
			tasks: [],
			starter_templates: []
		});

		render(ProfilePage);
		await waitFor(() => expect(screen.getByLabelText('Name *')).toBeInTheDocument());

		const nameInput = screen.getByLabelText('Name *');
		await fireEvent.input(nameInput, { target: { value: '' } });

		const emailInput = screen.getByLabelText('Email for recaps (optional)');
		await fireEvent.input(emailInput, { target: { value: 'not-an-email' } });

		const saveButton = screen.getByRole('button', { name: /Save and continue/i });
		await fireEvent.click(saveButton);

		await waitFor(() => {
			expect(screen.getByText('Your name is required.')).toBeInTheDocument();
			expect(screen.getByText('If you add an email, please use a valid one.')).toBeInTheDocument();
		});
		expect(saveProfile).not.toHaveBeenCalled();
	});
});
