import '@testing-library/jest-dom';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import Tile from './Tile.svelte';

describe('Tile', () => {
	it('renders title and subtitle', () => {
		render(Tile, { props: { title: 'Clean kitchen', subtitle: '20 min · daily', icon: '🧹' } });
		expect(screen.getByText('Clean kitchen')).toBeInTheDocument();
		expect(screen.getByText('20 min · daily')).toBeInTheDocument();
	});

	it('shows a selected state', () => {
		render(Tile, { props: { title: 'Clean kitchen', selected: true } });
		expect(screen.getByRole('button')).toHaveClass('sel');
		expect(screen.getByText('✓')).toBeInTheDocument();
	});

	it('emits a click event', async () => {
		const onClick = vi.fn();
		render(Tile, { props: { title: 'Clean kitchen' } });
		const tile = screen.getByRole('button');
		tile.addEventListener('click', onClick);
		await fireEvent.click(tile);
		expect(onClick).toHaveBeenCalledTimes(1);
	});
});
