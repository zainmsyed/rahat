import '@testing-library/jest-dom';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import LoginPage from './+page.svelte';

const mockGoto = vi.fn();
const exchangeAccessLink = vi.fn();
const getCurrentSession = vi.fn();
const logout = vi.fn();

vi.mock('$app/navigation', () => ({ goto: (path: string) => mockGoto(path) }));
vi.mock('$lib/api/auth', () => ({
	exchangeAccessLink: (token: string) => exchangeAccessLink(token),
	getCurrentSession: () => getCurrentSession(),
	logout: () => logout()
}));

describe('LoginPage', () => {
	beforeEach(() => {
		vi.resetAllMocks();
		vi.stubGlobal('window', {
			...window,
			location: { href: 'http://localhost:5200/login' }
		});
	});

	it('shows signed-out instructions when no session exists', async () => {
		getCurrentSession.mockResolvedValue({ authenticated: false });
		render(LoginPage);
		expect(await screen.findByText('Use your beta access link.')).toBeTruthy();
	});

	it('exchanges an access token from the URL and redirects home', async () => {
		vi.stubGlobal('window', {
			...window,
			location: { href: 'http://localhost:5200/login?token=demo-token' }
		});
		exchangeAccessLink.mockResolvedValue({ authenticated: true, user: { display_name: 'Tester' } });
		render(LoginPage);
		await new Promise((resolve) => setTimeout(resolve, 10));
		expect(exchangeAccessLink).toHaveBeenCalledWith('demo-token');
		expect(mockGoto).toHaveBeenCalledWith('/tasks');
	});

	it('exchanges a token entered into the form and redirects home', async () => {
		getCurrentSession.mockResolvedValue({ authenticated: false });
		exchangeAccessLink.mockResolvedValue({ authenticated: true, user: { display_name: 'Tester' } });

		render(LoginPage);
		await screen.findByText('Use your beta access link.');

		const tokenInput = screen.getByLabelText('Beta access token *');
		await fireEvent.input(tokenInput, { target: { value: 'typed-token' } });

		const signInButton = screen.getByRole('button', { name: /Sign in/i });
		await fireEvent.click(signInButton);

		await new Promise((resolve) => setTimeout(resolve, 10));
		expect(exchangeAccessLink).toHaveBeenCalledWith('typed-token');
		expect(mockGoto).toHaveBeenCalledWith('/tasks');
	});

	it('shows an error when the access link exchange fails', async () => {
		getCurrentSession.mockResolvedValue({ authenticated: false });
		exchangeAccessLink.mockRejectedValue(new Error('Link expired'));

		render(LoginPage);
		await screen.findByText('Use your beta access link.');

		const tokenInput = screen.getByLabelText('Beta access token *');
		await fireEvent.input(tokenInput, { target: { value: 'bad-token' } });

		const signInButton = screen.getByRole('button', { name: /Sign in/i });
		await fireEvent.click(signInButton);

		await new Promise((resolve) => setTimeout(resolve, 10));
		expect(screen.getByText('Link expired')).toBeInTheDocument();
	});
});
