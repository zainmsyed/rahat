import { apiBaseUrl } from './config';

export type LookaheadItem = {
	name: string;
	window: TimeWindow;
	duration_minutes: number;
	ready_at?: string;
};

export type LookaheadOmittedItem = {
	name: string;
	window: TimeWindow;
	reason: string;
};

export type TimeWindow = 'morning' | 'afternoon' | 'evening';

export type LookaheadDay = {
	date: string;
	label: string;
	windows: Record<TimeWindow, LookaheadItem[]>;
	blocked_windows: Record<TimeWindow, string[]>;
	omitted_items: LookaheadOmittedItem[];
	small_task_only_reason?: string;
	window_budgets_minutes: Record<TimeWindow, number>;
};

export type LookaheadResponse = {
	user: {
		display_name: string;
		timezone: string;
	};
	days: LookaheadDay[];
};

export const windows: TimeWindow[] = ['morning', 'afternoon', 'evening'];

export async function getLookaheadPlan(token: string) {
	const response = await fetch(`${apiBaseUrl}/lookahead/plan?token=${encodeURIComponent(token)}`);
	if (!response.ok) {
		throw new Error(await response.text());
	}
	return (await response.json()) as LookaheadResponse;
}

export function readLookaheadTokenFromUrl() {
	if (typeof window === 'undefined') {
		return '';
	}
	return new URL(window.location.href).searchParams.get('token') ?? '';
}

export function formatReadyTime(value?: string) {
	if (!value) {
		return '';
	}
	return new Date(value).toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' });
}
