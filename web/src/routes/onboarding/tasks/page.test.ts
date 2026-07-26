import '@testing-library/jest-dom';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import TasksPage from './+page.svelte';

const mockGoto = vi.fn();
const getState = vi.fn();
const addStarterTask = vi.fn();
const getStoredOnboardingToken = vi.fn();
const readTokenFromUrl = vi.fn();
const setStoredOnboardingToken = vi.fn();
const syncTokenInUrl = vi.fn();
const clearStoredOnboardingToken = vi.fn();

vi.mock('$app/navigation', () => ({ goto: (path: string) => mockGoto(path) }));
vi.mock('$lib/api/onboarding', () => ({
	addStarterTask: (token: string, templateId: string) => addStarterTask(token, templateId),
	buildOnboardingSteps: () => [
		{ id: 0, title: 'Start', required: true, description: '', complete: true },
		{ id: 1, title: 'Profile', required: true, description: '', complete: true },
		{ id: 2, title: 'Telegram', required: false, description: '', complete: false },
		{ id: 3, title: 'Calendar', required: false, description: '', complete: false },
		{ id: 4, title: 'Tasks', required: true, description: '', complete: false },
		{ id: 5, title: 'Review', required: true, description: '', complete: false }
	],
	clearStoredOnboardingToken: () => clearStoredOnboardingToken(),
	createTask: vi.fn(),
	deleteTask: vi.fn(),
	emptyTaskDraft: () => ({
		name: '',
		description: '',
		duration_minutes: 20,
		cadence_type: 'interval',
		cadence_value: 1,
		priority: 'medium',
		time_of_day_preference: 'morning',
		subtasks: []
	}),
	formatStepLabel: (step: { id: number; required: boolean }) =>
		`Step ${step.id + 1} · ${step.required ? 'Required' : 'Recommended'}`,
	formatTaskFrequency: () => 'Every day',
	getState: (token: string) => getState(token),
	getStoredOnboardingToken: () => getStoredOnboardingToken(),
	readTokenFromUrl: () => readTokenFromUrl(),
	setStoredOnboardingToken: (token: string) => setStoredOnboardingToken(token),
	syncTokenInUrl: (token: string) => syncTokenInUrl(token),
	updateTask: vi.fn()
}));

describe('TasksPage', () => {
	beforeEach(() => {
		vi.resetAllMocks();
		getStoredOnboardingToken.mockReturnValue('stored-token');
		readTokenFromUrl.mockReturnValue('url-token');
		vi.stubGlobal('localStorage', {
			getItem: vi.fn(),
			setItem: vi.fn(),
			removeItem: vi.fn()
		});
	});

	it('redirects to onboarding start when no token is available', async () => {
		getStoredOnboardingToken.mockReturnValue('');
		readTokenFromUrl.mockReturnValue('');

		render(TasksPage);
		await new Promise((resolve) => setTimeout(resolve, 10));

		expect(mockGoto).toHaveBeenCalledWith('/onboarding');
	});

	it('renders starter templates and saved tasks', async () => {
		getState.mockResolvedValue({
			has_profile: true,
			telegram_linked: false,
			calendar_connected: false,
			starter_templates: [
				{
					id: 'starter-1',
					name: 'Tidy up',
					description: 'A quick tidy routine.',
					duration_minutes: 15,
					cadence_type: 'interval',
					cadence_value: 1,
					priority: 'medium',
					time_of_day_preference: 'any',
					is_multistep: false,
					subtasks: []
				}
			],
			tasks: [
				{
					id: 'task-1',
					name: 'Custom task',
					description: '',
					duration_minutes: 30,
					cadence_type: 'interval',
					cadence_value: 2,
					priority: 'high',
					time_of_day_preference: 'morning',
					is_multistep: false,
					subtasks: []
				}
			]
		});

		render(TasksPage);
		await waitFor(() => expect(screen.getByText('Tidy up')).toBeInTheDocument());

		expect(screen.getByText('Custom task')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: /Review my setup/i })).toBeEnabled();
	});

	it('adds a starter task when its tile is clicked', async () => {
		getState.mockResolvedValue({
			has_profile: true,
			telegram_linked: false,
			calendar_connected: false,
			starter_templates: [
				{
					id: 'starter-1',
					name: 'Tidy up',
					description: 'A quick tidy routine.',
					duration_minutes: 15,
					cadence_type: 'interval',
					cadence_value: 1,
					priority: 'medium',
					time_of_day_preference: 'any',
					is_multistep: false,
					subtasks: []
				}
			],
			tasks: []
		});
		addStarterTask.mockResolvedValue({});

		render(TasksPage);
		await waitFor(() => expect(screen.getByText('Tidy up')).toBeInTheDocument());

		const tile = screen.getByRole('button', { name: /Tidy up/i });
		await fireEvent.click(tile);

		await waitFor(() => expect(addStarterTask).toHaveBeenCalledWith('url-token', 'starter-1'));
	});
});
