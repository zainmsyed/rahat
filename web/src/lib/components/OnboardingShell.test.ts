import '@testing-library/jest-dom';
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import OnboardingShellWrapper from './OnboardingShellWrapper.svelte';

const steps = [
	{ id: 0, title: 'Welcome', required: true, description: '', complete: false },
	{ id: 1, title: 'Profile', required: true, description: '', complete: false },
	{ id: 2, title: 'Tasks', required: true, description: '', complete: false },
	{ id: 3, title: 'Telegram', required: true, description: '', complete: false },
	{ id: 4, title: 'Calendar', required: false, description: '', complete: false },
	{ id: 5, title: 'Review', required: false, description: '', complete: false }
];

describe('OnboardingShell', () => {
	it('renders the wordmark, step counter, title, intro, and slot', () => {
		render(OnboardingShellWrapper, {
			props: { steps, currentStep: 2, title: 'Pick your routines', intro: 'Choose what feels right.' }
		});

		expect(screen.getByText('Rahat')).toBeInTheDocument();
		expect(screen.getByText('3 of 6')).toBeInTheDocument();
		expect(screen.getByText('Step 3 of 6')).toBeInTheDocument();
		expect(screen.getByRole('heading', { name: 'Pick your routines' })).toBeInTheDocument();
		expect(screen.getByText('Choose what feels right.')).toBeInTheDocument();
		expect(screen.getByText('Slot content rendered.')).toBeInTheDocument();
	});

	it('shows the done state when finished', () => {
		render(OnboardingShellWrapper, {
			props: { steps, currentStep: 5, title: 'Ready to go', intro: 'You are ready.', finished: true }
		});

		expect(screen.getByText('Done')).toBeInTheDocument();
		expect(screen.getByText('All set')).toBeInTheDocument();
		expect(screen.getByRole('heading', { name: 'Ready to go' })).toBeInTheDocument();
	});
});
