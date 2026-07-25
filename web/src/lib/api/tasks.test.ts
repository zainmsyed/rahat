import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createTask, listTasks, removeTask, setTaskPaused, toDraft, updateTask } from './tasks';

const fetchMock = vi.fn();

beforeEach(() => {
	fetchMock.mockReset();
	vi.stubGlobal('fetch', fetchMock);
});

describe('task management api', () => {
	it('lists, creates, edits, pauses, resumes, and removes tasks with credentials', async () => {
		fetchMock.mockResolvedValue({ ok: true, status: 200, json: async () => [] });
		await listTasks();
		expect(fetchMock).toHaveBeenLastCalledWith(expect.stringContaining('/tasks'), expect.objectContaining({ credentials: 'include' }));

		fetchMock.mockResolvedValue({ ok: true, status: 201, json: async () => ({ id: 'task-1', subtasks: [] }) });
		await createTask({ name: 'Task', description: '', duration_minutes: 10, cadence_type: 'interval', cadence_value: 1, priority: 'medium', time_of_day_preference: 'morning', subtasks: [] });
		expect(fetchMock.mock.calls.at(-1)?.[1]).toMatchObject({ method: 'POST' });

		fetchMock.mockResolvedValue({ ok: true, status: 200, json: async () => ({ id: 'task-1', subtasks: [] }) });
		await updateTask('task-1', { name: 'Task', description: '', duration_minutes: 10, cadence_type: 'interval', cadence_value: 1, priority: 'medium', time_of_day_preference: 'morning', subtasks: [] });
		expect(fetchMock.mock.calls.at(-1)?.[0]).toContain('/tasks/task-1');

		await setTaskPaused('task-1', true);
		expect(fetchMock.mock.calls.at(-1)?.[1]?.body).toBe(JSON.stringify({ paused: true }));
		await setTaskPaused('task-1', false);
		expect(fetchMock.mock.calls.at(-1)?.[1]?.body).toBe(JSON.stringify({ paused: false }));

		fetchMock.mockResolvedValue({ ok: true, status: 204, json: async () => ({}) });
		await removeTask('task-1');
		expect(fetchMock.mock.calls.at(-1)?.[1]).toMatchObject({ method: 'DELETE' });
	});

	it('surfaces authentication and validation failures', async () => {
		fetchMock.mockResolvedValue({ ok: false, status: 401, text: async () => 'missing session' });
		await expect(listTasks()).rejects.toThrow('missing session');
		fetchMock.mockResolvedValue({ ok: false, status: 400, text: async () => 'task name is required' });
		await expect(createTask({ name: '', description: '', duration_minutes: 10, cadence_type: 'interval', cadence_value: 1, priority: 'medium', time_of_day_preference: 'morning', subtasks: [] })).rejects.toThrow('task name is required');
	});

	it('keeps internal subtask metadata when mapping an existing task to a draft', () => {
		const draft = toDraft({ id: 'task-1', name: 'Laundry', description: '', duration_minutes: 30, cadence_type: 'interval', cadence_value: 2, priority: 'medium', time_of_day_preference: 'morning', is_multistep: true, is_paused: false, subtasks: [{ id: 'sub-1', name: 'Wash', duration_minutes: 10, time_of_day_preference: 'morning', dependency_type: 'soft_followup', min_gap_after_previous_minutes: 45 }] });
		expect(draft.subtasks[0].dependency_type).toBe('soft_followup');
		expect(draft.subtasks[0].min_gap_after_previous_minutes).toBe(45);
	});
});
