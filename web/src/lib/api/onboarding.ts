export type CadenceType = 'interval' | 'count';
export type Priority = 'high' | 'medium' | 'low';
export type TimeOfDayPreference = 'any' | 'morning' | 'afternoon' | 'evening';

export type OnboardingProfile = {
	display_name: string;
	timezone: string;
	daily_time_budget_minutes: number;
	email: string;
};

export type OnboardingSubtask = {
	id?: string;
	name: string;
	duration_minutes: number;
	time_of_day_preference: TimeOfDayPreference;
	min_gap_after_previous_minutes: number;
};

export type OnboardingTask = {
	id: string;
	name: string;
	description: string;
	duration_minutes: number;
	cadence_type: CadenceType;
	cadence_value: number;
	priority: Priority;
	time_of_day_preference: TimeOfDayPreference;
	is_multistep: boolean;
	subtasks: OnboardingSubtask[];
};

export type StarterTemplate = {
	id: string;
	slug: string;
	name: string;
	description: string;
	duration_minutes: number;
	cadence_type: CadenceType;
	cadence_value: number;
	priority: Priority;
	time_of_day_preference: TimeOfDayPreference;
	is_multistep: boolean;
	subtasks: OnboardingSubtask[];
};

export type OnboardingState = {
	has_profile: boolean;
	user?: OnboardingProfile;
	tasks: OnboardingTask[];
	starter_templates: StarterTemplate[];
};

export type OnboardingFinishResult = {
	profile: OnboardingProfile;
	plan_date: string;
	task_count: number;
	scheduled_count: number;
	overflowed_count: number;
	skipped_count: number;
	summary: string[];
	scheduled_items: { name: string; window: string; ready_at?: string }[];
	next_checkpoint?: string;
};

export type TaskDraft = {
	name: string;
	description: string;
	duration_minutes: number;
	cadence_type: CadenceType;
	cadence_value: number;
	priority: Priority;
	time_of_day_preference: TimeOfDayPreference;
	subtasks: OnboardingSubtask[];
};

export type OnboardingStep = {
	id: number;
	title: string;
	required: boolean;
	description: string;
	complete: boolean;
};

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';
const storageKey = 'rahat-onboarding-token';

async function request<T>(path: string, init?: RequestInit): Promise<T> {
	const response = await fetch(`${apiBaseUrl}${path}`, {
		headers: {
			'Content-Type': 'application/json',
			...(init?.headers ?? {})
		},
		...init
	});

	if (!response.ok) {
		throw new Error(await response.text());
	}

	if (response.status === 204) {
		return undefined as T;
	}

	return (await response.json()) as T;
}

export function createSession(inviteCode: string) {
	return request<{ token: string }>('/onboarding/session', {
		method: 'POST',
		body: JSON.stringify({ invite_code: inviteCode })
	});
}

export function getState(token: string) {
	return request<OnboardingState>(`/onboarding/state?token=${encodeURIComponent(token)}`);
}

export function saveProfile(token: string, profile: OnboardingProfile) {
	return request<OnboardingProfile>(`/onboarding/profile?token=${encodeURIComponent(token)}`, {
		method: 'POST',
		body: JSON.stringify(profile)
	});
}

export function addStarterTask(token: string, templateId: string) {
	return request<OnboardingTask>(`/onboarding/tasks/from-template?token=${encodeURIComponent(token)}`, {
		method: 'POST',
		body: JSON.stringify({ template_id: templateId })
	});
}

export function createTask(token: string, task: TaskDraft) {
	return request<OnboardingTask>(`/onboarding/tasks?token=${encodeURIComponent(token)}`, {
		method: 'POST',
		body: JSON.stringify(task)
	});
}

export function updateTask(token: string, taskId: string, task: TaskDraft) {
	return request<OnboardingTask>(
		`/onboarding/tasks/${encodeURIComponent(taskId)}?token=${encodeURIComponent(token)}`,
		{
			method: 'PUT',
			body: JSON.stringify(task)
		}
	);
}

export function deleteTask(token: string, taskId: string) {
	return request<void>(`/onboarding/tasks/${encodeURIComponent(taskId)}?token=${encodeURIComponent(token)}`, {
		method: 'DELETE'
	});
}

export function finishOnboarding(token: string) {
	return request<OnboardingFinishResult>(`/onboarding/finish?token=${encodeURIComponent(token)}`, {
		method: 'POST',
		body: JSON.stringify({})
	});
}

export function emptyTaskDraft(): TaskDraft {
	return {
		name: '',
		description: '',
		duration_minutes: 20,
		cadence_type: 'interval',
		cadence_value: 1,
		priority: 'medium',
		time_of_day_preference: 'morning',
		subtasks: []
	};
}

export function getStoredOnboardingToken() {
	if (typeof localStorage === 'undefined') {
		return '';
	}
	return localStorage.getItem(storageKey) ?? '';
}

export function setStoredOnboardingToken(token: string) {
	if (typeof localStorage === 'undefined') {
		return;
	}
	localStorage.setItem(storageKey, token);
}

export function clearStoredOnboardingToken() {
	if (typeof localStorage === 'undefined') {
		return;
	}
	localStorage.removeItem(storageKey);
}

export function syncTokenInUrl(token: string) {
	if (typeof window === 'undefined') {
		return;
	}
	const url = new URL(window.location.href);
	url.searchParams.set('token', token);
	url.searchParams.delete('invite');
	window.history.replaceState({}, '', url);
}

export function readInviteCodeFromUrl() {
	if (typeof window === 'undefined') {
		return '';
	}
	return new URL(window.location.href).searchParams.get('invite') ?? '';
}

export function readTokenFromUrl() {
	if (typeof window === 'undefined') {
		return '';
	}
	return new URL(window.location.href).searchParams.get('token') ?? '';
}

export function formatTaskFrequency(task: Pick<OnboardingTask, 'cadence_type' | 'cadence_value'>) {
	return task.cadence_type === 'interval'
		? `Every ${task.cadence_value} day(s)`
		: `${task.cadence_value} time(s) each week`;
}

export function formatTaskSummary(task: OnboardingTask) {
	const parts: string[] = [`${task.duration_minutes} min`];
	if (task.cadence_type && task.cadence_value) {
		parts.push(formatTaskFrequency(task));
	}
	if (task.time_of_day_preference && task.time_of_day_preference !== 'any') {
		parts.push(`best in the ${task.time_of_day_preference}`);
	}
	if (task.subtasks.length > 0) {
		parts.push(`${task.subtasks.length} step(s)`);
	}
	return parts.join(' · ');
}

export function buildOnboardingSteps(state: OnboardingState, hasSession: boolean, finished = false): OnboardingStep[] {
	return [
		{
			id: 0,
			title: 'Start with your invite code',
			required: true,
			description: 'Enter the invite code you were given so Rahat can open your guided setup.',
			complete: hasSession
		},
		{
			id: 1,
			title: 'Tell Rahat about you',
			required: true,
			description: 'Add your name, timezone, daily task-time budget, and an optional email for recaps.',
			complete: state.has_profile
		},
		{
			id: 2,
			title: 'Pick at least one task',
			required: true,
			description: 'Choose from the starter ideas or add your own task and steps in plain language.',
			complete: state.tasks.length > 0
		},
		{
			id: 3,
			title: 'Review and finish',
			required: true,
			description: 'Confirm everything looks right, let Rahat seed your first schedule, and read what happens next.',
			complete: finished
		}
	];
}

export function nextOnboardingPath(state: OnboardingState) {
	if (!state.has_profile) {
		return '/onboarding/profile';
	}
	if (state.tasks.length === 0) {
		return '/onboarding/tasks';
	}
	return '/onboarding/review';
}
