import '@testing-library/jest-dom';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import ConnectTile from './ConnectTile.svelte';

describe('ConnectTile', () => {
	it('renders name, subtitle, and disconnected status', () => {
		render(ConnectTile, {
			props: { icon: '📅', name: 'Google Calendar', subtitle: 'Read-only access', connected: false }
		});

		expect(screen.getByText('Google Calendar')).toBeInTheDocument();
		expect(screen.getByText('Read-only access')).toBeInTheDocument();
		expect(screen.getByText('Not connected')).toBeInTheDocument();
	});

	it('shows connected status and styles', () => {
		render(ConnectTile, {
			props: { icon: '📅', name: 'Google Calendar', connected: true }
		});

		expect(screen.getByText('Connected')).toBeInTheDocument();
		expect(screen.getByRole('button')).toHaveClass('connected');
	});

	it('emits a click event on button tiles', async () => {
		render(ConnectTile, {
			props: { icon: '✈️', name: 'Telegram', connected: false }
		});
		const onClick = vi.fn();
		const tile = screen.getByRole('button');
		tile.addEventListener('click', onClick);

		await fireEvent.click(tile);
		expect(onClick).toHaveBeenCalledTimes(1);
	});

	it('renders an anchor when href is provided', () => {
		render(ConnectTile, {
			props: {
				icon: '📅',
				name: 'Google Calendar',
				connected: false,
				href: 'https://accounts.google.test/oauth',
				target: '_blank',
				rel: 'noopener noreferrer'
			}
		});

		const link = screen.getByRole('link');
		expect(link).toHaveAttribute('href', 'https://accounts.google.test/oauth');
		expect(link).toHaveAttribute('target', '_blank');
	});
});
