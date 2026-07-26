import '@testing-library/jest-dom';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import ToggleTestWrapper from './ToggleTestWrapper.svelte';

describe('Toggle', () => {
	it('renders unchecked by default', () => {
		render(ToggleTestWrapper, { props: { id: 'pause', label: 'Active' } });

		const switchEl = screen.getByRole('switch');
		expect(switchEl).toHaveAttribute('aria-checked', 'false');
		expect(screen.getByText('Active')).toBeInTheDocument();
	});

	it('renders checked when the prop is set', () => {
		render(ToggleTestWrapper, { props: { id: 'pause', checked: true } });

		expect(screen.getByRole('switch')).toHaveAttribute('aria-checked', 'true');
	});

	it('toggles state and dispatches a change event on click', async () => {
		const handler = vi.fn();
		render(ToggleTestWrapper, { props: { id: 'pause', onChange: handler } });

		const switchEl = screen.getByRole('switch');
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
