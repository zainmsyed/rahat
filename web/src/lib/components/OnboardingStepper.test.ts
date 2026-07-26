import '@testing-library/jest-dom';
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import OnboardingStepper from './OnboardingStepper.svelte';

const steps = [
	{ id: 0, title: 'Welcome', required: true, description: '', complete: false },
	{ id: 1, title: 'Profile', required: true, description: '', complete: false },
	{ id: 2, title: 'Tasks', required: true, description: '', complete: false },
	{ id: 3, title: 'Telegram', required: true, description: '', complete: false },
	{ id: 4, title: 'Review', required: false, description: '', complete: false }
];

describe('OnboardingStepper', () => {
	it('renders a progressbar with the correct percentage', () => {
		render(OnboardingStepper, { props: { steps, currentStep: 1 } });

		const bar = screen.getByRole('progressbar');
		expect(bar).toHaveAttribute('aria-valuenow', '25');
		expect(bar).toHaveAttribute('aria-valuemin', '0');
		expect(bar).toHaveAttribute('aria-valuemax', '100');
	});

	it('shows 0% at the first step and 100% when finished', () => {
		const { rerender } = render(OnboardingStepper, { props: { steps, currentStep: 0 } });
		expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '0');

		rerender({ steps, currentStep: 4, finished: true });
		expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '100');
	});
});
