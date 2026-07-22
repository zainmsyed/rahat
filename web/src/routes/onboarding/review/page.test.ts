import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import ReviewPage from './+page.svelte';

const mockGoto = vi.fn();
vi.mock('$app/navigation', () => ({
	goto: (path: string) => mockGoto(path)
}));

describe('ReviewPage', () => {
	beforeEach(() => {
		vi.resetAllMocks();
		vi.stubGlobal('localStorage', {
			getItem: vi.fn(),
			setItem: vi.fn(),
			removeItem: vi.fn()
		});
	});

	it('redirects to onboarding start when no token is available', async () => {
		vi.stubGlobal('window', {
			...window,
			location: { href: 'http://localhost:5200/onboarding/review' },
			history: { replaceState: vi.fn() }
		});

		render(ReviewPage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(mockGoto).toHaveBeenCalledWith('/onboarding');
	});
});
