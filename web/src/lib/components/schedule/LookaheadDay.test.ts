import '@testing-library/jest-dom';
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import LookaheadDay from './LookaheadDay.svelte';
import type { LookaheadDay as LookaheadDayData } from '$lib/api/lookahead';

const day: LookaheadDayData = {
	date: '2026-07-24',
	label: 'Today',
	windows: {
		morning: [{ name: 'Laundry', window: 'morning', duration_minutes: 20, ready_at: '2026-07-24T08:00:00Z' }],
		afternoon: [],
		evening: [{ name: 'Review mail', window: 'evening', duration_minutes: 10 }]
	},
	blocked_windows: {
		morning: [],
		afternoon: ['Dentist (medium)'],
		evening: []
	},
	omitted_items: [{ name: 'Deep clean', window: 'afternoon', reason: 'Calendar blocked the afternoon window: Dentist (medium)' }],
	small_task_only_reason: '',
	window_budgets_minutes: { morning: 30, afternoon: 0, evening: 15 }
};

describe('LookaheadDay', () => {
	it('renders tasks grouped by window with blocked explanations and omitted items', () => {
		render(LookaheadDay, { props: { day } });

		expect(screen.getByText('Today')).toBeInTheDocument();
		expect(screen.getByText('Laundry')).toBeInTheDocument();
		expect(screen.getByText('Review mail')).toBeInTheDocument();
		expect(screen.getByText('Dentist (medium)')).toBeInTheDocument();
		expect(screen.getByText('Deep clean')).toBeInTheDocument();
		expect(screen.getByText(/Calendar blocked the afternoon window/i)).toBeInTheDocument();
	});

	it('does not render editing or completion controls', () => {
		render(LookaheadDay, { props: { day } });

		expect(screen.queryByRole('button')).not.toBeInTheDocument();
		expect(screen.queryByText(/edit/i)).not.toBeInTheDocument();
		expect(screen.queryByText(/complete/i)).not.toBeInTheDocument();
	});
});
