import '@testing-library/jest-dom';
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import LandingPage from './+page.svelte';

describe('landing page', () => {
	it('renders a centered card with display heading, lede, and primary CTA', () => {
		render(LandingPage);

		expect(screen.getByRole('heading', { name: 'Rahat' })).toBeInTheDocument();
		expect(screen.getByText(/calm, adaptive routine planner/i)).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Manage routines' })).toHaveAttribute('href', '/tasks');
	});
});
