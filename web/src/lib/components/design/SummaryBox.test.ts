import '@testing-library/jest-dom';
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import SummaryBox from './SummaryBox.svelte';

describe('SummaryBox', () => {
	it('renders the value, unit, and hint', () => {
		render(SummaryBox, {
			props: {
				id: 'budget',
				value: 120,
				unit: 'minutes per day',
				hint: 'Friendly default: 45 minutes.'
			}
		});

		expect(screen.getByText('120')).toBeInTheDocument();
		expect(screen.getByText('minutes per day')).toBeInTheDocument();
		expect(screen.getByText('Friendly default: 45 minutes.')).toBeInTheDocument();
	});

	it('uses the id on the wrapper and on the hint element', () => {
		const { container } = render(SummaryBox, {
			props: {
				id: 'budget',
				value: 30,
				unit: 'minutes',
				hint: 'Move the slider to change it.'
			}
		});

		expect(container.querySelector('#budget')).toBeInTheDocument();
		expect(container.querySelector('#budget-hint')).toHaveTextContent(
			'Move the slider to change it.'
		);
	});
});
