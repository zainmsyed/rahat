import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import TaskEditor from './TaskEditor.svelte';
import { emptyTaskDraft } from '$lib/api/onboarding';

describe('TaskEditor', () => {
	it('renders the task form with default draft', () => {
		render(TaskEditor, { props: { draft: emptyTaskDraft() } });

		expect(screen.getByPlaceholderText('Example: Wipe down the kitchen')).toBeDefined();
		expect(screen.getByText('Save task')).toBeDefined();
	});

	it('updates the task name when typed', async () => {
		render(TaskEditor, { props: { draft: emptyTaskDraft() } });

		const nameInput = screen.getByPlaceholderText('Example: Wipe down the kitchen');
		await fireEvent.input(nameInput, { target: { value: 'New task' } });

		expect((nameInput as HTMLInputElement).value).toBe('New task');
	});

	it('renders cancel button', () => {
		render(TaskEditor, { props: { draft: emptyTaskDraft() } });

		expect(screen.getByText('Cancel')).toBeDefined();
	});

	it('adds and removes subtasks', async () => {
		render(TaskEditor, { props: { draft: emptyTaskDraft() } });

		expect(screen.getByText('No steps added. That is okay for a simple one-step task.')).toBeDefined();

		const addStepButton = screen.getByText('Add a step');
		await fireEvent.click(addStepButton);

		expect(screen.getByText('Step 1')).toBeDefined();

		const removeButton = screen.getByText('Remove');
		await fireEvent.click(removeButton);

		expect(screen.getByText('No steps added. That is okay for a simple one-step task.')).toBeDefined();
	});
});
