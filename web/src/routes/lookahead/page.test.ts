import '@testing-library/jest-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import LookaheadPage from './+page.svelte';

const getLookaheadPlan = vi.fn();
const readLookaheadTokenFromUrl = vi.fn();

vi.mock('$lib/api/lookahead', () => ({
	getLookaheadPlan: (token: string) => getLookaheadPlan(token),
	readLookaheadTokenFromUrl: () => readLookaheadTokenFromUrl(),
	windows: ['morning', 'afternoon', 'evening']
}));

const planFixture = {
	user: { display_name: 'Zain', timezone: 'America/Chicago' },
	range_days: 2,
	days: [
		{
			date: '2026-07-26',
			label: 'Today',
			windows: {
				morning: [{ name: 'Morning walk', window: 'morning', duration_minutes: 15 }],
				afternoon: [],
				evening: [{ name: 'Evening tidy', window: 'evening', duration_minutes: 15 }]
			},
			blocked_windows: { morning: [], afternoon: ['Dentist (medium)'], evening: [] },
			omitted_items: [
				{
					name: 'Deep clean',
					window: 'afternoon',
					reason: 'Calendar blocked the afternoon window: Dentist (medium)'
				}
			],
			overflowed: [],
			skipped: [],
			reasons: {},
			small_task_only_reason: '',
			window_budgets_minutes: { morning: 30, afternoon: 0, evening: 15 }
		},
		{
			date: '2026-07-27',
			label: 'Tomorrow',
			windows: { morning: [], afternoon: [], evening: [] },
			blocked_windows: { morning: [], afternoon: [], evening: [] },
			omitted_items: [],
			overflowed: [],
			skipped: [],
			reasons: {},
			small_task_only_reason: '',
			window_budgets_minutes: { morning: 60, afternoon: 60, evening: 60 }
		}
	]
};

beforeEach(() => {
	vi.resetAllMocks();
	getLookaheadPlan.mockResolvedValue(planFixture);
});

describe('lookahead page', () => {
	it('shows a loading state while fetching the plan', () => {
		readLookaheadTokenFromUrl.mockReturnValue('valid-token');
		getLookaheadPlan.mockImplementation(() => new Promise(() => {}));
		render(LookaheadPage);

		expect(screen.getByText('Loading your lookahead…')).toBeInTheDocument();
	});

	it('shows an error when the token is missing', async () => {
		readLookaheadTokenFromUrl.mockReturnValue('');
		render(LookaheadPage);

		expect(await screen.findByText(/missing its access token/i)).toBeInTheDocument();
		expect(getLookaheadPlan).not.toHaveBeenCalled();
	});

	it('shows an error when the plan request fails', async () => {
		readLookaheadTokenFromUrl.mockReturnValue('bad-token');
		getLookaheadPlan.mockRejectedValueOnce(new Error('token expired'));
		render(LookaheadPage);

		expect(await screen.findByText(/token expired/)).toBeInTheDocument();
		expect(screen.getByText(/Ask Rahat for a fresh link/i)).toBeInTheDocument();
	});

	it('renders the user, days, windows, and omitted items after loading', async () => {
		readLookaheadTokenFromUrl.mockReturnValue('valid-token');
		render(LookaheadPage);

		await waitFor(() => expect(screen.getByText(/Zain/)).toBeInTheDocument());
		expect(screen.getByText(/America\/Chicago/)).toBeInTheDocument();
		expect(screen.getByText('Today')).toBeInTheDocument();
		expect(screen.getByText('Tomorrow')).toBeInTheDocument();
		expect(screen.getByText('Morning walk')).toBeInTheDocument();
		expect(screen.getByText('Evening tidy')).toBeInTheDocument();
		expect(screen.getByText('Dentist (medium)')).toBeInTheDocument();
		expect(screen.getByText('Deep clean')).toBeInTheDocument();
	});

	it('does not render editing or completion controls', async () => {
		readLookaheadTokenFromUrl.mockReturnValue('valid-token');
		render(LookaheadPage);

		await waitFor(() => expect(screen.getByText(/Zain/)).toBeInTheDocument());
		expect(screen.queryByRole('button', { name: /edit/i })).not.toBeInTheDocument();
		expect(screen.queryByRole('button', { name: /complete/i })).not.toBeInTheDocument();
	});
});
