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

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

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
