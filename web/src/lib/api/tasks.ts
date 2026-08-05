import { apiBaseUrl } from './config';
import type { OnboardingTask, TaskDraft } from './onboarding';

export type ManagedTask = OnboardingTask & {
	is_paused: boolean;
	archived_at?: string;
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
	const response = await fetch(`${apiBaseUrl}${path}`, {
		credentials: 'include',
		headers: {
			'Content-Type': 'application/json',
			Accept: 'application/json',
			...(init?.headers ?? {})
		},
		...init
	});
	if (!response.ok) throw new Error(await response.text());
	if (response.status === 204) return undefined as T;
	return (await response.json()) as T;
}

export function listTasks() {
	return request<ManagedTask[]>('/api/tasks');
}

export function createTask(task: TaskDraft) {
	return request<ManagedTask>('/api/tasks', { method: 'POST', body: JSON.stringify(task) });
}

export function updateTask(taskId: string, task: TaskDraft) {
	return request<ManagedTask>(`/api/tasks/${encodeURIComponent(taskId)}`, { method: 'PUT', body: JSON.stringify(task) });
}

export function setTaskPaused(taskId: string, paused: boolean) {
	return request<ManagedTask>(`/api/tasks/${encodeURIComponent(taskId)}/pause`, {
		method: 'POST',
		body: JSON.stringify({ paused })
	});
}

export function removeTask(taskId: string) {
	return request<void>(`/api/tasks/${encodeURIComponent(taskId)}`, { method: 'DELETE' });
}

export function toDraft(task: ManagedTask): TaskDraft {
	return {
		name: task.name,
		description: task.description,
		duration_minutes: task.duration_minutes,
		cadence_type: task.cadence_type,
		cadence_value: task.cadence_value,
		priority: task.priority,
		time_of_day_preference: task.time_of_day_preference,
		day_preference: task.day_preference ?? 'any',
		subtasks: task.subtasks.map((subtask) => ({ ...subtask }))
	};
}
