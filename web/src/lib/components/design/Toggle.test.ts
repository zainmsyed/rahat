import '@testing-library/jest-dom';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import Toggle from './Toggle.svelte';

describe('Toggle', () => {
	it('renders unchecked by default', () => {
		render(Toggle, { props: { id: 'pause', label: 'Active' } });

		const switchEl = screen.getByRole('switch');
		expect(switchEl).toHaveAttribute('aria-checked', 'false');
		expect(screen.getByText('Active')).toBeInTheDocument();
	});

	it('renders checked when the prop is set', () => {
		render(Toggle, { props: { id: 'pause', checked: true } });

		expect(screen.getByRole('switch')).toHaveAttribute('aria-checked', 'true');
	});

	it('toggles state and dispatches a change event on click', async () => {
		render(Toggle, { props: { id: 'pause' } });
		const handler = vi.fn();
		const switchEl = screen.getByRole('switch');
		switchEl.addEventListener('change', handler);

		await fireEvent.click(switchEl);

		expect(switchEl).toHaveAttribute('aria-checked', 'true');
		expect(handler).toHaveBeenCalledWith(expect.objectContaining({ detail: { checked: true } }));

		await fireEvent.click(switchEl);
		expect(switchEl).toHaveAttribute('aria-checked', 'false');
		expect(handler).toHaveBeenLastCalledWith(
			expect.objectContaining({ detail: { checked: false } })
		);
	});
});
