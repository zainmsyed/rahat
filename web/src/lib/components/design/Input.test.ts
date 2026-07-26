import '@testing-library/jest-dom';
import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import Input from './Input.svelte';

describe('Input', () => {
	it('renders a labeled input', () => {
		render(Input, { props: { id: 'name', label: 'First name', value: 'Sarah' } });
		expect(screen.getByLabelText('First name')).toBeInTheDocument();
		expect(screen.getByRole('textbox')).toHaveValue('Sarah');
	});

	it('shows a required marker on the label', () => {
		render(Input, { props: { id: 'email', label: 'Email', required: true } });
		expect(screen.getByLabelText('Email *')).toBeInTheDocument();
	});

	it('displays an error message and adds an error class', () => {
		render(Input, { props: { id: 'code', label: 'Invite code', error: 'Required' } });
		expect(screen.getByText('Required')).toBeInTheDocument();
		expect(screen.getByRole('textbox')).toHaveClass('error');
	});

	it('renders a numeric value and min/max attributes', () => {
		render(Input, {
			props: {
				id: 'budget',
				label: 'Budget',
				type: 'number',
				value: 45,
				min: 15,
				max: 480
			}
		});

		const input = screen.getByRole('spinbutton') as HTMLInputElement;
		expect(input).toHaveValue(45);
		expect(input).toHaveAttribute('min', '15');
		expect(input).toHaveAttribute('max', '480');
	});

	it('updates a numeric value on input', async () => {
		render(Input, {
			props: { id: 'budget', label: 'Budget', type: 'number', value: 45 }
		});

		const input = screen.getByRole('spinbutton') as HTMLInputElement;
		await fireEvent.input(input, { target: { value: '90' } });
		expect(input).toHaveValue(90);
	});

	it('styles the field as errored when the invalid prop is set', () => {
		const { container } = render(Input, {
			props: { id: 'code', label: 'Invite code', invalid: true }
		});

		expect(screen.getByRole('textbox')).toHaveClass('error');
		expect(container.querySelector('.error-text')).not.toBeInTheDocument();
	});
});
