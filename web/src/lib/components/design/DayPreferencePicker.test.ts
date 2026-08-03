import '@testing-library/jest-dom';
import { describe, expect, it } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/svelte';
import DayPreferencePicker from './DayPreferencePicker.svelte';

describe('DayPreferencePicker', () => {
	it('renders guided cards and changes the selected value', async () => {
		render(DayPreferencePicker, { props: { value: 'any', variant: 'cards' } });

		expect(screen.getByRole('button', { name: /Any day is fine/ })).toHaveAttribute('aria-pressed', 'true');
		expect(screen.getByText('Planned Monday to Friday, never on weekends.')).toBeInTheDocument();

		await fireEvent.click(screen.getByRole('button', { name: /Weekends only/ }));
		expect(screen.getByRole('button', { name: /Weekends only/ })).toHaveAttribute('aria-pressed', 'true');
	});

	it('renders compact segmented options', () => {
		render(DayPreferencePicker, { props: { value: 'weekday', variant: 'segmented' } });

		expect(screen.getByRole('button', { name: 'Weekdays' })).toHaveAttribute('aria-pressed', 'true');
		expect(screen.queryByText('Planned Monday to Friday, never on weekends.')).toBeNull();
	});
});
