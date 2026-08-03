import '@testing-library/jest-dom';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import ReviewPage from './+page.svelte';

const mocks = vi.hoisted(() => ({
	goto: vi.fn(),
	getState: vi.fn(),
	finishOnboarding: vi.fn(),
	getStoredOnboardingToken: vi.fn(),
	readTokenFromUrl: vi.fn(),
	clearStoredOnboardingToken: vi.fn()
}));

vi.mock('$app/navigation', () => ({
	goto: (path: string) => mocks.goto(path)
}));

vi.mock('$lib/api/onboarding', () => ({
	buildOnboardingSteps: (_state: unknown, _hasToken: boolean, finished: boolean) =>
		Array.from({ length: 6 }, (_, id) => ({
			id,
			title: id === 5 ? 'Review' : `Step ${id + 1}`,
			required: true,
			description: '',
			complete: finished || id < 5
		})),
		clearStoredOnboardingToken: () => mocks.clearStoredOnboardingToken(),
		finishOnboarding: (token: string) => mocks.finishOnboarding(token),
		formatStepLabel: () => 'Step 6 of 6',
		formatTaskSummary: (task: { duration_minutes: number }) => `${task.duration_minutes} minutes`,
		getState: (token: string) => mocks.getState(token),
		getStoredOnboardingToken: () => mocks.getStoredOnboardingToken(),
		readTokenFromUrl: () => mocks.readTokenFromUrl()
}));

const state = {
	has_profile: true,
	telegram_linked: true,
	calendar_connected: true,
	user: {
		display_name: 'Alex',
		timezone: 'America/New_York',
		daily_time_budget_minutes: 60,
		email: 'alex@example.com'
	},
	tasks: [
		{
			id: 'task-1',
			name: 'Write',
			description: '',
			duration_minutes: 30,
			cadence_type: 'interval',
			cadence_value: 1,
			priority: 'high',
			time_of_day_preference: 'morning',
			is_multistep: false,
			subtasks: []
		}
	],
	starter_templates: []
};

const finishResult = {
	profile: state.user,
	plan_date: '2026-08-04',
	task_count: 1,
	scheduled_count: 1,
	overflowed_count: 0,
	skipped_count: 0,
	summary: ['Your first routine is planned.'],
	scheduled_items: [{ name: 'Write', window: 'morning', ready_at: '2026-08-04T08:00:00Z' }],
	telegram_delivered: true
};

function preparePage() {
	mocks.getStoredOnboardingToken.mockReturnValue('test-token');
	mocks.readTokenFromUrl.mockReturnValue('');
	mocks.getState.mockResolvedValue(state);
	vi.stubGlobal('localStorage', {
		getItem: vi.fn().mockReturnValue('test-token'),
		setItem: vi.fn(),
		removeItem: vi.fn()
	});
}

describe('ReviewPage', () => {
	beforeEach(() => {
		vi.resetAllMocks();
	});

	it('redirects to onboarding start when no token is available', async () => {
		mocks.getStoredOnboardingToken.mockReturnValue('');
		mocks.readTokenFromUrl.mockReturnValue('');

		render(ReviewPage);
		await waitFor(() => expect(mocks.goto).toHaveBeenCalledWith('/onboarding'));
	});

	it('renders pre-filled profile, task, and calendar summaries', async () => {
		preparePage();

		render(ReviewPage);
		await waitFor(() => expect(screen.getByRole('heading', { name: 'Everything looks ready.' })).toBeInTheDocument());

		expect(screen.getByText('Alex')).toBeInTheDocument();
		expect(screen.getByText('America/New_York')).toBeInTheDocument();
		expect(screen.getByText('60 minutes')).toBeInTheDocument();
		expect(screen.getByText('Write')).toBeInTheDocument();
		expect(screen.getByText('Google Calendar')).toBeInTheDocument();
		expect(screen.getAllByRole('article')).toHaveLength(3);
		expect(screen.getByRole('button', { name: 'Back' })).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Finish onboarding' })).toBeEnabled();
	});

	it('finishes onboarding and preserves the redirect', async () => {
		preparePage();
		mocks.finishOnboarding.mockResolvedValue(finishResult);

		render(ReviewPage);
		await waitFor(() => expect(screen.getByRole('button', { name: 'Finish onboarding' })).toBeInTheDocument());
		await fireEvent.click(screen.getByRole('button', { name: 'Finish onboarding' }));

		await waitFor(() => expect(mocks.finishOnboarding).toHaveBeenCalledWith('test-token'));
		expect(mocks.clearStoredOnboardingToken).toHaveBeenCalled();
		expect(mocks.goto).not.toHaveBeenCalledWith('/');
		expect(screen.getByRole('heading', { name: 'Your first schedule is ready.' })).toBeInTheDocument();
		await fireEvent.click(screen.getByRole('button', { name: 'Continue to Rahat' }));
		expect(mocks.goto).toHaveBeenCalledWith('/');
	});

	it('shows an actionable error when finishing fails', async () => {
		preparePage();
		mocks.finishOnboarding.mockRejectedValue(new Error('Schedule service unavailable'));

		render(ReviewPage);
		await waitFor(() => expect(screen.getByRole('button', { name: 'Finish onboarding' })).toBeInTheDocument());
		await fireEvent.click(screen.getByRole('button', { name: 'Finish onboarding' }));

		await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('Schedule service unavailable'));
		expect(screen.getByRole('button', { name: 'Finish onboarding' })).toBeEnabled();
	});

	it('renders the completed schedule summary after a successful finish', async () => {
		preparePage();
		mocks.finishOnboarding.mockResolvedValue(finishResult);

		render(ReviewPage);
		await waitFor(() => expect(screen.getByRole('button', { name: 'Finish onboarding' })).toBeInTheDocument());
		await fireEvent.click(screen.getByRole('button', { name: 'Finish onboarding' }));

		await waitFor(() => expect(screen.getByRole('heading', { name: 'Your first schedule is ready.' })).toBeInTheDocument());
		expect(screen.getByText('2026-08-04')).toBeInTheDocument();
		expect(screen.getByText('Today\'s planned items')).toBeInTheDocument();
		expect(screen.getAllByText('Write')).toHaveLength(2);
	});
});
