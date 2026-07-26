import '@testing-library/jest-dom';
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
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
});
