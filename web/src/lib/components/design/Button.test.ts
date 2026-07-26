import '@testing-library/jest-dom';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import Button from './Button.svelte';

describe('Button', () => {
	it('renders a primary button by default', () => {
		render(Button, { props: { variant: 'primary' } });
		const button = screen.getByRole('button');
		expect(button).toHaveClass('btn-primary');
		expect(button).not.toHaveClass('btn-secondary');
	});

	it('renders secondary and text variants', () => {
		const { rerender } = render(Button, { props: { variant: 'secondary' } });
		expect(screen.getByRole('button')).toHaveClass('btn-secondary');

		rerender({ variant: 'text' });
		expect(screen.getByRole('button')).toHaveClass('btn-text');
	});

	it('forwards the disabled state', () => {
		render(Button, { props: { disabled: true } });
		expect(screen.getByRole('button')).toBeDisabled();
	});

	it('emits a click event', async () => {
		const onClick = vi.fn();
		render(Button, { props: { variant: 'primary' } });
		const button = screen.getByRole('button');
		button.addEventListener('click', onClick);
		await fireEvent.click(button);
		expect(onClick).toHaveBeenCalledTimes(1);
	});

	it('renders an anchor when href is provided', () => {
		render(Button, { props: { variant: 'primary', href: '/tasks' } });
		const link = screen.getByRole('link');
		expect(link).toHaveAttribute('href', '/tasks');
		expect(link).toHaveClass('btn-primary');
	});
});
