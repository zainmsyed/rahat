import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/svelte';
import TasksPage from './+page.svelte';

const listTasks = vi.fn();
const createTask = vi.fn();
const updateTask = vi.fn();
const setTaskPaused = vi.fn();
const removeTask = vi.fn();

vi.mock('$lib/api/tasks', () => ({
	listTasks: () => listTasks(),
	createTask: (draft: unknown) => createTask(draft),
	updateTask: (id: string, draft: unknown) => updateTask(id, draft),
	setTaskPaused: (id: string, paused: boolean) => setTaskPaused(id, paused),
	removeTask: (id: string) => removeTask(id),
	toDraft: (task: ManagedTaskFixture) => ({
		name: task.name,
		description: task.description,
		duration_minutes: task.duration_minutes,
		cadence_type: task.cadence_type,
		cadence_value: task.cadence_value,
		priority: task.priority,
		time_of_day_preference: task.time_of_day_preference,
		subtasks: task.subtasks.map((subtask) => ({ ...subtask }))
	})
}));

type ManagedTaskFixture = {
	id: string;
	name: string;
	description: string;
	duration_minutes: number;
	cadence_type: 'interval' | 'count';
	cadence_value: number;
	priority: 'high' | 'medium' | 'low';
	time_of_day_preference: 'any' | 'morning' | 'afternoon' | 'evening';
	is_multistep: boolean;
	is_paused: boolean;
	archived_at?: string;
	subtasks: Array<{
		id?: string;
		name: string;
		duration_minutes: number;
		time_of_day_preference: 'any' | 'morning' | 'afternoon' | 'evening';
		dependency_type?: 'required_same_day' | 'soft_followup';
		min_gap_after_previous_minutes: number;
	}>;
};

const activeTask: ManagedTaskFixture = {
	id: 'active-1',
	name: 'Water plants',
	description: 'Kitchen first',
	duration_minutes: 15,
	cadence_type: 'interval',
	cadence_value: 2,
	priority: 'medium',
	time_of_day_preference: 'morning',
	is_multistep: false,
	is_paused: false,
	subtasks: []
};

const pausedTask: ManagedTaskFixture = {
	...activeTask,
	id: 'paused-1',
	name: 'Laundry',
	is_paused: true
};

const removedTask: ManagedTaskFixture = {
	...activeTask,
	id: 'removed-1',
	name: 'Old routine',
	archived_at: '2026-07-25T00:00:00Z'
};

beforeEach(() => {
	vi.resetAllMocks();
	listTasks.mockResolvedValue([activeTask, pausedTask, removedTask]);
	createTask.mockResolvedValue(activeTask);
	updateTask.mockResolvedValue(activeTask);
	setTaskPaused.mockResolvedValue(activeTask);
	removeTask.mockResolvedValue(undefined);
});

describe('task management page', () => {
	it('groups active, paused, and removed routines', async () => {
		render(TasksPage);
		await screen.findByText('Water plants');
		expect(screen.getByRole('heading', { name: 'Active' })).toBeTruthy();
		expect(screen.getByRole('heading', { name: 'Paused' })).toBeTruthy();
		expect(screen.getByRole('heading', { name: 'Removed' })).toBeTruthy();
		expect(screen.getByText('Laundry')).toBeTruthy();
		expect(screen.getByText('Old routine')).toBeTruthy();
		expect(screen.getByText('Preserved for history')).toBeTruthy();
		expect(screen.queryByRole('link', { name: 'Lookahead' })).toBeNull();
	});

	it('creates a routine with the shared editor', async () => {
		render(TasksPage);
		await screen.findByText('Water plants');
		await fireEvent.click(screen.getByRole('button', { name: 'Add a routine' }));
		const name = screen.getByPlaceholderText('Example: Wipe down the kitchen');
		await fireEvent.input(name, { target: { value: 'New routine' } });
		await fireEvent.click(screen.getByRole('button', { name: 'Create routine' }));
		await waitFor(() => expect(createTask).toHaveBeenCalledWith(expect.objectContaining({ name: 'New routine' })));
	});

	it('edits an existing routine and preserves its mapped values', async () => {
		render(TasksPage);
		await screen.findByText('Water plants');
		await fireEvent.click(screen.getAllByRole('button', { name: 'Edit' })[0]);
		const name = screen.getByDisplayValue('Water plants');
		await fireEvent.input(name, { target: { value: 'Water indoor plants' } });
		await fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));
		await waitFor(() => expect(updateTask).toHaveBeenCalledWith('active-1', expect.objectContaining({ name: 'Water indoor plants' })));
	});

	it('confirms pause and resume actions before calling the API', async () => {
		render(TasksPage);
		await screen.findByText('Water plants');
		await fireEvent.click(screen.getByRole('switch', { name: 'Active' }));
		let dialog = screen.getByRole('dialog');
		expect(within(dialog).getByText('Pause this routine?')).toBeTruthy();
		await fireEvent.click(within(dialog).getByRole('button', { name: 'Confirm' }));
		await waitFor(() => expect(setTaskPaused).toHaveBeenCalledWith('active-1', true));

		await fireEvent.click(screen.getByRole('switch', { name: 'Paused' }));
		dialog = screen.getByRole('dialog');
		await fireEvent.click(within(dialog).getByRole('button', { name: 'Confirm' }));
		await waitFor(() => expect(setTaskPaused).toHaveBeenCalledWith('paused-1', false));
	});

	it('allows removal confirmation to be cancelled and then accepted', async () => {
		render(TasksPage);
		await screen.findByText('Water plants');
		const removeButtons = screen.getAllByRole('button', { name: 'Remove' });
		await fireEvent.click(removeButtons[0]);
		let dialog = screen.getByRole('dialog');
		expect(within(dialog).getByText('Remove this routine?')).toBeTruthy();
		await fireEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }));
		expect(removeTask).not.toHaveBeenCalled();

		await fireEvent.click(screen.getAllByRole('button', { name: 'Remove' })[0]);
		dialog = screen.getByRole('dialog');
		await fireEvent.click(within(dialog).getByRole('button', { name: 'Confirm' }));
		await waitFor(() => expect(removeTask).toHaveBeenCalledWith('active-1'));
	});

	it('shows authentication and action failures', async () => {
		listTasks.mockRejectedValueOnce(new Error('missing session'));
		const { unmount } = render(TasksPage);
		expect(await screen.findByText('missing session')).toBeTruthy();
		unmount();

		listTasks.mockResolvedValue([activeTask]);
		setTaskPaused.mockRejectedValueOnce(new Error('Could not pause routine'));
		render(TasksPage);
		await screen.findByText('Water plants');
		await fireEvent.click(screen.getByRole('switch', { name: 'Active' }));
		await fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Confirm' }));
		expect(await screen.findByText('Could not pause routine')).toBeTruthy();
	});
});
